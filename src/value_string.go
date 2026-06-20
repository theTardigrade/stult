package main

import "fmt"

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
