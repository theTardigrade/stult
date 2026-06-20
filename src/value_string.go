package main

import (
	"fmt"
	"math/big"
)

type String struct {
	Runes    []rune
	IsFrozen bool
}

func NewStringValue(value string) Value {
	return NewStringValueWithFrozen(value, false)
}

func NewStringValueWithFrozen(value string, isFrozen bool) Value {
	return Value{
		Kind: ValueString,
		Text: &String{
			Runes:    []rune(value),
			IsFrozen: isFrozen,
		},
	}
}

func (s *String) String() string {
	if s == nil {
		return ""
	}

	return string(s.Runes)
}

func (s *String) Len() *Number {
	if s == nil {
		return NewSmallNumber(0)
	}

	return NewSmallNumber(int64(len(s.Runes)))
}

func (s *String) Get(index *Number) (Value, bool, error) {
	if s == nil {
		return Value{}, false, fmt.Errorf("invalid string")
	}

	integer, err := stringNumberToInteger(index)
	if err != nil {
		return Value{}, false, err
	}

	if integer.Sign() < 0 || integer.Cmp(s.lengthInteger()) >= 0 {
		return Value{}, false, nil
	}

	hostIndex, err := stringIntegerToHostIndex(integer)
	if err != nil {
		return Value{}, false, err
	}

	return NewStringValue(string(s.Runes[hostIndex])), true, nil
}

func (s *String) Set(index *Number, value Value) error {
	if s == nil {
		return fmt.Errorf("invalid string")
	}

	if s.IsFrozen {
		return fmt.Errorf("cannot modify frozen string")
	}

	integer, err := stringNumberToInteger(index)
	if err != nil {
		return err
	}

	if integer.Sign() < 0 {
		return fmt.Errorf("string index cannot be negative")
	}

	length := s.lengthInteger()
	comparison := integer.Cmp(length)
	if comparison > 0 {
		return fmt.Errorf(
			"string index %s is past the next append position %s",
			formatStringIndex(index),
			s.Len().Format(DefaultDecimalPlacesToDisplay),
		)
	}

	replacement, err := stringAssignmentRune(value)
	if err != nil {
		return err
	}

	if comparison == 0 {
		s.Runes = append(s.Runes, replacement)
		return nil
	}

	hostIndex, err := stringIntegerToHostIndex(integer)
	if err != nil {
		return err
	}

	s.Runes[hostIndex] = replacement
	return nil
}

func (s *String) Append(value Value) error {
	return s.Set(s.Len(), value)
}

func (s *String) ForEach(fn func(index *Number, value Value) error) error {
	if s == nil {
		return fmt.Errorf("invalid string")
	}

	for index := 0; index < len(s.Runes); index++ {
		if err := fn(NewSmallNumber(int64(index)), NewStringValue(string(s.Runes[index]))); err != nil {
			return err
		}
	}

	return nil
}

func (s *String) Clear() error {
	if s == nil {
		return fmt.Errorf("invalid string")
	}

	if s.IsFrozen {
		return fmt.Errorf("cannot modify frozen string")
	}

	s.Runes = nil
	return nil
}

func (s *String) lengthInteger() *big.Int {
	if s == nil {
		return big.NewInt(0)
	}

	return big.NewInt(int64(len(s.Runes)))
}

func stringAssignmentRune(value Value) (rune, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueString {
		return 0, fmt.Errorf("string index assignment requires a string value")
	}

	if value.Text == nil {
		return 0, fmt.Errorf("string index assignment requires a valid string value")
	}

	if len(value.Text.Runes) != 1 {
		return 0, fmt.Errorf("string index assignment value must contain exactly one rune")
	}

	return value.Text.Runes[0], nil
}

func stringNumberToInteger(index *Number) (*big.Int, error) {
	if index == nil {
		return nil, fmt.Errorf("string index must be a number")
	}

	integer, accuracy := index.Int(nil)
	if accuracy != big.Exact {
		return nil, fmt.Errorf("string index must be an integer")
	}

	return integer, nil
}

func stringIntegerToHostIndex(index *big.Int) (int, error) {
	if index == nil || !index.IsInt64() {
		return 0, fmt.Errorf("string index %s out of bounds", formatStringBigIntIndex(index))
	}

	hostIndex64 := index.Int64()
	hostIndex := int(hostIndex64)
	if int64(hostIndex) != hostIndex64 {
		return 0, fmt.Errorf("string index %s out of bounds", index.String())
	}

	return hostIndex, nil
}

func formatStringIndex(index *Number) string {
	if index == nil {
		return "<invalid>"
	}

	return index.Format(DefaultDecimalPlacesToDisplay)
}

func formatStringBigIntIndex(index *big.Int) string {
	if index == nil {
		return "<invalid>"
	}

	return index.String()
}
