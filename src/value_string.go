package main

import "fmt"

type String struct {
	runes              []rune
	nativeCache        string
	isNativeCacheValid bool
	IsFrozen           bool
}

func NewStringValue(value string) Value {
	return NewStringValueWithFrozen(value, false)
}

func NewStringValueWithFrozen(value string, isFrozen bool) Value {
	return Value{
		Kind: ValueString,
		Text: NewStringFromNative(value, isFrozen),
	}
}

func NewStringValueWithFrozenFromRunes(runes []rune, isFrozen bool) Value {
	return Value{
		Kind: ValueString,
		Text: NewString(runes, isFrozen),
	}
}

func NewString(runes []rune, isFrozen bool) *String {
	return &String{
		runes:    append([]rune(nil), runes...),
		IsFrozen: isFrozen,
	}
}

func NewStringFromNative(value string, isFrozen bool) *String {
	return &String{
		runes:              []rune(value),
		nativeCache:        value,
		isNativeCacheValid: true,
		IsFrozen:           isFrozen,
	}
}

func (s *String) RuneCount() int {
	if s == nil {
		return 0
	}

	return len(s.runes)
}

func (s *String) Len() *Number {
	return NewSmallNumber(int64(s.RuneCount()))
}

func (s *String) IsEmpty() bool {
	return s.RuneCount() == 0
}

func (s *String) Get(index int) (rune, bool, error) {
	if s == nil {
		return 0, false, fmt.Errorf("invalid string")
	}

	if index < 0 || index >= len(s.runes) {
		return 0, false, nil
	}

	return s.runes[index], true, nil
}

func (s *String) Set(index int, value rune) error {
	if s == nil {
		return fmt.Errorf("invalid string")
	}

	if s.IsFrozen {
		return fmt.Errorf("cannot modify frozen string")
	}

	if index < 0 {
		return fmt.Errorf("string index cannot be negative")
	}

	if index > len(s.runes) {
		return fmt.Errorf(
			"string index %d is past the next append position %d",
			index,
			len(s.runes),
		)
	}

	if index == len(s.runes) {
		s.runes = append(s.runes, value)
		s.invalidateNativeCache()
		return nil
	}

	s.runes[index] = value
	s.invalidateNativeCache()
	return nil
}

func (s *String) Clear() error {
	if s == nil {
		return fmt.Errorf("invalid string")
	}

	if s.IsFrozen {
		return fmt.Errorf("cannot modify frozen string")
	}

	s.runes = nil
	s.invalidateNativeCache()
	return nil
}

func (s *String) CloneRunes() []rune {
	if s == nil {
		return []rune{}
	}

	return append([]rune(nil), s.runes...)
}

func (s *String) ForEach(fn func(index int, r rune) error) error {
	if s == nil {
		return fmt.Errorf("invalid string")
	}

	for index := 0; index < len(s.runes); index++ {
		if err := fn(index, s.runes[index]); err != nil {
			return err
		}
	}

	return nil
}

func (s *String) String() string {
	if s == nil {
		return ""
	}

	if !s.isNativeCacheValid {
		s.nativeCache = string(s.runes)
		s.isNativeCacheValid = true
	}

	return s.nativeCache
}

func (s *String) invalidateNativeCache() {
	s.nativeCache = ""
	s.isNativeCacheValid = false
}

func stringAssignmentRune(value Value) (rune, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueString {
		return 0, fmt.Errorf("string index assignment requires a string value")
	}

	if value.Text == nil {
		return 0, fmt.Errorf("string index assignment requires a valid string value")
	}

	if value.Text.RuneCount() != 1 {
		return 0, fmt.Errorf("string index assignment value must contain exactly one rune")
	}

	r, _, err := value.Text.Get(0)
	return r, err
}
