package main

import (
	"fmt"
	"math/big"
	"strings"
)

type Array struct {
	Elements    []Value
	IsImmutable bool
}

func NewArrayValue(elements []Value, isImmutable bool) Value {
	return Value{
		Kind: ValueArray,
		Array: &Array{
			Elements:    elements,
			IsImmutable: isImmutable,
		},
	}
}

func (array *Array) Len() *Number {
	if array == nil {
		return NewSmallNumber(0)
	}

	return NewSmallNumber(int64(len(array.Elements)))
}

func (array *Array) Get(index *Number) (Value, bool, error) {
	if array == nil {
		return Value{}, false, fmt.Errorf("invalid array")
	}

	hostIndex, valid, err := arrayNumberToHostIndex(index)
	if err != nil {
		return Value{}, false, err
	}

	if !valid || hostIndex < 0 || hostIndex >= int64(len(array.Elements)) {
		return Value{}, false, nil
	}

	return array.Elements[hostIndex], true, nil
}

func (array *Array) Set(index *Number, value Value) error {
	if array == nil {
		return fmt.Errorf("invalid array")
	}

	if array.IsImmutable {
		return fmt.Errorf("cannot modify frozen array")
	}

	hostIndex, valid, err := arrayNumberToHostIndex(index)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("array index %s out of bounds", formatArrayIndex(index))
	}

	if hostIndex < 0 {
		return fmt.Errorf("array index cannot be negative")
	}

	if hostIndex > int64(len(array.Elements)) {
		return fmt.Errorf(
			"array index %d is past the next append position %d",
			hostIndex,
			len(array.Elements),
		)
	}

	if hostIndex == int64(len(array.Elements)) {
		array.Elements = append(array.Elements, value)
	} else {
		array.Elements[hostIndex] = value
	}

	return nil
}

func (array *Array) Append(value Value) error {
	return array.Set(array.Len(), value)
}

func (array *Array) ForEach(fn func(index *Number, value Value) error) error {
	if array == nil {
		return fmt.Errorf("invalid array")
	}

	for index, value := range array.Elements {
		if err := fn(NewSmallNumber(int64(index)), value); err != nil {
			return err
		}
	}

	return nil
}

func (array *Array) Clear() error {
	if array == nil {
		return fmt.Errorf("invalid array")
	}

	if array.IsImmutable {
		return fmt.Errorf("cannot modify frozen array")
	}

	array.Elements = nil

	return nil
}

func arrayNumberToHostIndex(index *Number) (int64, bool, error) {
	if index == nil {
		return 0, false, fmt.Errorf("array index must be a number")
	}

	integer, accuracy := index.Int(nil)
	if accuracy != big.Exact {
		return 0, false, fmt.Errorf("array index must be an integer")
	}

	if !integer.IsInt64() {
		return 0, false, nil
	}

	hostIndex := integer.Int64()

	return hostIndex, true, nil
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

	parts := make([]string, 0, len(a.Elements))

	_ = a.ForEach(func(_ *Number, element Value) error {
		parts = append(parts, state.formatValue(element))
		return nil
	})

	return "{" + strings.Join(parts, ", ") + "}"
}
