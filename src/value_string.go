package main

import "fmt"

type String struct {
	Runes              []rune
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
		Text: NewStringWithFrozenFromNative(value, isFrozen),
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
		Runes:    append([]rune(nil), runes...),
		IsFrozen: isFrozen,
	}
}

func NewStringWithFrozenFromNative(value string, isFrozen bool) *String {
	return &String{
		Runes:              []rune(value),
		nativeCache:        value,
		isNativeCacheValid: true,
		IsFrozen:           isFrozen,
	}
}

func (s *String) EntryCount() int {
	if s == nil {
		return 0
	}

	return len(s.Runes)
}

func (s *String) Len() *Number {
	return NewSmallNumber(int64(s.EntryCount()))
}

func (s *String) IsEmpty() bool {
	return s.EntryCount() == 0
}

func (s *String) Get(index int) (rune, bool, error) {
	if s == nil {
		return 0, false, fmt.Errorf("invalid string")
	}

	if index < 0 || index >= len(s.Runes) {
		return 0, false, nil
	}

	return s.Runes[index], true, nil
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

	if index > len(s.Runes) {
		return fmt.Errorf(
			"string index %d is past the next append position %d",
			index,
			len(s.Runes),
		)
	}

	if index == len(s.Runes) {
		s.Runes = append(s.Runes, value)
		s.invalidateNativeCache()
		return nil
	}

	s.Runes[index] = value
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

	s.Runes = nil
	s.invalidateNativeCache()
	return nil
}

func (s *String) CloneRunes() []rune {
	if s == nil {
		return []rune{}
	}

	return append([]rune(nil), s.Runes...)
}

func (s *String) ForEach(fn func(index int, r rune) error) error {
	if s == nil {
		return fmt.Errorf("invalid string")
	}

	for index := 0; index < len(s.Runes); index++ {
		if err := fn(index, s.Runes[index]); err != nil {
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
		s.nativeCache = string(s.Runes)
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

	if value.Text.EntryCount() != 1 {
		return 0, fmt.Errorf("string index assignment value must contain exactly one rune")
	}

	r, _, err := value.Text.Get(0)
	return r, err
}
