package main

import (
	"fmt"
	"math/big"
	"strings"
)

const arrayOrdinaryLimit int64 = 1 << 24
const arrayOverflowChunkSize int64 = 1 << 16

type Array struct {
	Ordinary    []Value
	Overflow    *ArrayOverflow
	Length      *Number
	IsImmutable bool
}

type ArrayOverflow struct {
	Chunks map[string][]Value
}

func NewArrayValue(elements []Value, isImmutable bool) Value {
	array := &Array{
		Length: NewSmallNumber(0),
	}

	for _, element := range elements {
		if err := array.appendUnchecked(element); err != nil {
			return Value{
				Kind:  ValueArray,
				Array: array,
			}
		}
	}

	array.IsImmutable = isImmutable

	return Value{
		Kind:  ValueArray,
		Array: array,
	}
}

func (array *Array) Len() *Number {
	if array == nil {
		return NewSmallNumber(0)
	}

	if array.Length != nil {
		return CloneNumber(array.Length)
	}

	return NewSmallNumber(int64(len(array.Ordinary)))
}

func (array *Array) Get(index *Number) (Value, bool, error) {
	if array == nil {
		return Value{}, false, fmt.Errorf("invalid array")
	}

	integer, err := arrayNumberToInteger(index)
	if err != nil {
		return Value{}, false, err
	}

	if integer.Sign() < 0 || integer.Cmp(array.lengthInteger()) >= 0 {
		return Value{}, false, nil
	}

	if integer.Cmp(big.NewInt(arrayOrdinaryLimit)) < 0 {
		hostIndex64 := integer.Int64()
		hostIndex := int(hostIndex64)
		if int64(hostIndex) != hostIndex64 || hostIndex < 0 || hostIndex >= len(array.Ordinary) {
			return Value{}, false, nil
		}

		return array.Ordinary[hostIndex], true, nil
	}

	if array.Overflow == nil {
		return Value{}, false, nil
	}

	return array.Overflow.Get(integer)
}

func (array *Array) Set(index *Number, value Value) error {
	if array == nil {
		return fmt.Errorf("invalid array")
	}

	if array.IsImmutable {
		return fmt.Errorf("cannot modify frozen array")
	}

	integer, err := arrayNumberToInteger(index)
	if err != nil {
		return err
	}

	if integer.Sign() < 0 {
		return fmt.Errorf("array index cannot be negative")
	}

	length := array.lengthInteger()
	comparison := integer.Cmp(length)
	if comparison > 0 {
		return fmt.Errorf(
			"array index %s is past the next append position %s",
			formatArrayIndex(index),
			array.Len().Format(DefaultDecimalPlacesToDisplay),
		)
	}

	if comparison == 0 {
		return array.appendUnchecked(value)
	}

	return array.setExisting(integer, value)
}

func (array *Array) Append(value Value) error {
	return array.Set(array.Len(), value)
}

func (array *Array) ForEach(fn func(index *Number, value Value) error) error {
	if array == nil {
		return fmt.Errorf("invalid array")
	}

	for index, value := range array.Ordinary {
		if err := fn(NewSmallNumber(int64(index)), value); err != nil {
			return err
		}
	}

	if array.Overflow == nil {
		return nil
	}

	return array.Overflow.ForEach(array.lengthInteger(), fn)
}

func (array *Array) Clear() error {
	if array == nil {
		return fmt.Errorf("invalid array")
	}

	if array.IsImmutable {
		return fmt.Errorf("cannot modify frozen array")
	}

	array.Ordinary = nil
	array.Overflow = nil
	array.Length = NewSmallNumber(0)

	return nil
}

func (array *Array) appendUnchecked(value Value) error {
	if array == nil {
		return fmt.Errorf("invalid array")
	}

	length := array.lengthInteger()

	if length.Cmp(big.NewInt(arrayOrdinaryLimit)) < 0 {
		hostIndex64 := length.Int64()
		hostIndex := int(hostIndex64)
		if int64(hostIndex) != hostIndex64 || hostIndex != len(array.Ordinary) {
			return fmt.Errorf("invalid array ordinary storage")
		}

		array.Ordinary = append(array.Ordinary, value)
	} else {
		if array.Overflow == nil {
			array.Overflow = &ArrayOverflow{Chunks: make(map[string][]Value)}
		}

		if err := array.Overflow.Set(length, value); err != nil {
			return err
		}
	}

	length.Add(length, big.NewInt(1))
	array.Length = NewBigIntNumber(length)

	return nil
}

func (array *Array) setExisting(index *big.Int, value Value) error {
	if index.Cmp(big.NewInt(arrayOrdinaryLimit)) < 0 {
		hostIndex64 := index.Int64()
		hostIndex := int(hostIndex64)
		if int64(hostIndex) != hostIndex64 || hostIndex < 0 || hostIndex >= len(array.Ordinary) {
			return fmt.Errorf("array index %s out of bounds", index.String())
		}

		array.Ordinary[hostIndex] = value
		return nil
	}

	if array.Overflow == nil {
		return fmt.Errorf("array index %s out of bounds", index.String())
	}

	return array.Overflow.Set(index, value)
}

func (array *Array) lengthInteger() *big.Int {
	if array == nil {
		return big.NewInt(0)
	}

	if array.Length == nil {
		return big.NewInt(int64(len(array.Ordinary)))
	}

	integer, accuracy := array.Length.Int(nil)
	if accuracy != big.Exact {
		return big.NewInt(int64(len(array.Ordinary)))
	}

	return integer
}

func (overflow *ArrayOverflow) Get(logicalIndex *big.Int) (Value, bool, error) {
	if overflow == nil || logicalIndex == nil {
		return Value{}, false, fmt.Errorf("invalid array overflow storage")
	}

	key, offset := arrayOverflowPosition(logicalIndex)
	chunk, ok := overflow.Chunks[key]
	if !ok || offset < 0 || offset >= len(chunk) {
		return Value{}, false, nil
	}

	return chunk[offset], true, nil
}

func (overflow *ArrayOverflow) Set(logicalIndex *big.Int, value Value) error {
	if overflow == nil || logicalIndex == nil {
		return fmt.Errorf("invalid array overflow storage")
	}

	key, offset := arrayOverflowPosition(logicalIndex)
	chunk := overflow.Chunks[key]
	if offset < 0 || offset > len(chunk) {
		return fmt.Errorf("invalid array overflow index %s", logicalIndex.String())
	}

	if offset == len(chunk) {
		chunk = append(chunk, value)
	} else {
		chunk[offset] = value
	}

	overflow.Chunks[key] = chunk
	return nil
}

func (overflow *ArrayOverflow) ForEach(length *big.Int, fn func(index *Number, value Value) error) error {
	if overflow == nil || length == nil {
		return nil
	}

	ordinaryLimit := big.NewInt(arrayOrdinaryLimit)
	if length.Cmp(ordinaryLimit) <= 0 {
		return nil
	}

	chunkSize := big.NewInt(arrayOverflowChunkSize)
	overflowCount := new(big.Int).Sub(length, ordinaryLimit)
	chunkCount := new(big.Int).Sub(overflowCount, big.NewInt(1))
	chunkCount.Quo(chunkCount, chunkSize)
	chunkCount.Add(chunkCount, big.NewInt(1))

	for chunkIndex := big.NewInt(0); chunkIndex.Cmp(chunkCount) < 0; chunkIndex.Add(chunkIndex, big.NewInt(1)) {
		chunk := overflow.Chunks[chunkIndex.String()]
		if len(chunk) == 0 {
			return fmt.Errorf("invalid array overflow storage")
		}

		for offset, value := range chunk {
			logicalIndex := new(big.Int).Mul(chunkIndex, chunkSize)
			logicalIndex.Add(logicalIndex, big.NewInt(int64(offset)))
			logicalIndex.Add(logicalIndex, ordinaryLimit)

			if logicalIndex.Cmp(length) >= 0 {
				return nil
			}

			if err := fn(NewBigIntNumber(logicalIndex), value); err != nil {
				return err
			}
		}
	}

	return nil
}

func arrayOverflowPosition(logicalIndex *big.Int) (string, int) {
	overflowIndex := new(big.Int).Sub(logicalIndex, big.NewInt(arrayOrdinaryLimit))
	chunkIndex := new(big.Int)
	offset := new(big.Int)
	chunkIndex.QuoRem(overflowIndex, big.NewInt(arrayOverflowChunkSize), offset)

	return chunkIndex.String(), int(offset.Int64())
}

func arrayNumberToInteger(index *Number) (*big.Int, error) {
	if index == nil {
		return nil, fmt.Errorf("array index must be a number")
	}

	integer, accuracy := index.Int(nil)
	if accuracy != big.Exact {
		return nil, fmt.Errorf("array index must be an integer")
	}

	return integer, nil
}

func formatArrayIndex(index *Number) string {
	if index == nil {
		return "<invalid>"
	}

	return index.Format(DefaultDecimalPlacesToDisplay)
}

func (state *valueFormatState) formatArray(a *Array) string {
	if a == nil {
		return "{}"
	}

	if state.arrays[a] {
		return "<cyclical array>"
	}

	state.arrays[a] = true
	defer delete(state.arrays, a)

	parts := make([]string, 0, len(a.Ordinary))

	if err := a.ForEach(func(_ *Number, element Value) error {
		parts = append(parts, state.formatValue(element))
		return nil
	}); err != nil {
		return "<invalid array>"
	}

	return "{" + strings.Join(parts, ", ") + "}"
}
