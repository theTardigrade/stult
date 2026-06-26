package main

import (
	"fmt"
	"unicode/utf8"
)

type stringStorageState uint8

const (
	stringStorageNative stringStorageState = iota
	stringStorageRunes
	stringStorageBoth
)

type String struct {
	runes        []rune
	native       string
	storageState stringStorageState
	IsFrozen     bool
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
		runes:        append([]rune(nil), runes...),
		storageState: stringStorageRunes,
		IsFrozen:     isFrozen,
	}
}

func NewStringFromNative(value string, isFrozen bool) *String {
	return &String{
		native:       value,
		storageState: stringStorageNative,
		IsFrozen:     isFrozen,
	}
}

func (s *String) RuneCount() int {
	if s == nil {
		return 0
	}

	if s.hasRunes() {
		return len(s.runes)
	}

	return utf8.RuneCountInString(s.native)
}

func (s *String) Len() *Number {
	return NewSmallNumber(int64(s.RuneCount()))
}

func (s *String) IsEmpty() bool {
	return s.RuneCount() == 0
}

func (s *String) Get(index int) (rune, bool, error) {
	if err := s.ensureRunes(); err != nil {
		return 0, false, err
	}

	if index < 0 || index >= len(s.runes) {
		return 0, false, nil
	}

	return s.runes[index], true, nil
}

func (s *String) Set(index int, value rune) error {
	if err := s.ensureRunes(); err != nil {
		return err
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
		s.invalidateNative()
		return nil
	}

	s.runes[index] = value
	s.invalidateNative()
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
	s.invalidateNative()
	return nil
}

func (s *String) CloneRunes() []rune {
	if s == nil {
		return []rune{}
	}

	if err := s.ensureRunes(); err != nil {
		return []rune{}
	}

	return append([]rune(nil), s.runes...)
}

func (s *String) ForEach(fn func(index int, r rune) error) error {
	if s == nil {
		return fmt.Errorf("invalid string")
	}

	if s.hasRunes() {
		for index, r := range s.runes {
			if err := fn(index, r); err != nil {
				return err
			}
		}
		return nil
	}

	index := 0
	for _, r := range s.native {
		if err := fn(index, r); err != nil {
			return err
		}
		index++
	}
	return nil
}

func (s *String) String() string {
	if s == nil {
		return ""
	}

	s.ensureNative()
	return s.native
}

func (s *String) ensureRunes() error {
	if s == nil {
		return fmt.Errorf("invalid string")
	}

	if s.hasRunes() {
		return nil
	}

	s.runes = []rune(s.native)
	s.storageState = stringStorageBoth
	return nil
}

func (s *String) ensureNative() {
	if s.hasNative() {
		return
	}

	s.native = string(s.runes)
	s.storageState = stringStorageBoth
}

func (s *String) hasRunes() bool {
	return s.storageState == stringStorageRunes || s.storageState == stringStorageBoth
}

func (s *String) hasNative() bool {
	return s.storageState == stringStorageNative || s.storageState == stringStorageBoth
}

func (s *String) invalidateNative() {
	s.native = ""
	s.storageState = stringStorageRunes
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
