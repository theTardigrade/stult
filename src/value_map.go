package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Map struct {
	Entries  map[string]Binding
	IsFrozen bool
}

func NewMapValue(entries map[string]Binding, isFrozen bool) Value {
	return Value{
		Kind: ValueMap,
		Map: &Map{
			Entries:  entries,
			IsFrozen: isFrozen,
		},
	}
}

func (m *Map) Len() *Number {
	if m == nil {
		return NewSmallNumber(0)
	}

	return NewSmallNumber(int64(len(m.Entries)))
}

func (m *Map) Get(key Value) (Value, bool, error) {
	keyText, err := mapKeyString(key)
	if err != nil {
		return Value{}, false, err
	}

	return m.GetFromString(keyText)
}

func (m *Map) GetFromString(key string) (Value, bool, error) {
	binding, exists, err := m.Binding(key)
	if err != nil || !exists {
		return Value{}, exists, err
	}

	return binding.Value, true, nil
}

func (m *Map) Binding(key string) (Binding, bool, error) {
	if m == nil {
		return Binding{}, false, fmt.Errorf("invalid map")
	}

	binding, exists := m.Entries[key]
	return binding, exists, nil
}

func (m *Map) Set(key Value, value Value) error {
	keyText, err := mapKeyString(key)
	if err != nil {
		return err
	}

	return m.SetString(keyText, value)
}

func (m *Map) SetString(key string, value Value) error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	if m.IsFrozen {
		return fmt.Errorf("cannot modify frozen map")
	}

	if m.Entries == nil {
		m.Entries = make(map[string]Binding)
	}

	binding, exists := m.Entries[key]
	if exists && binding.IsImmutable {
		return fmt.Errorf("cannot reassign immutable map entry %q", key)
	}

	if exists {
		binding.Value = value
		m.Entries[key] = binding
		return nil
	}

	m.Entries[key] = Binding{
		Value:       value,
		IsImmutable: isImmutableIdentifier(key),
	}

	return nil
}

func (m *Map) Define(key string, value Value, isImmutable bool) error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	if m.IsFrozen {
		return fmt.Errorf("cannot modify frozen map")
	}

	if m.Entries == nil {
		m.Entries = make(map[string]Binding)
	}

	if _, exists := m.Entries[key]; exists {
		return fmt.Errorf("map already has key %q", key)
	}

	m.Entries[key] = Binding{
		Value:       value,
		IsImmutable: isImmutable,
	}

	return nil
}

func (m *Map) SetBindingUnchecked(key string, binding Binding) error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	if m.Entries == nil {
		m.Entries = make(map[string]Binding)
	}

	m.Entries[key] = binding
	return nil
}

func (m *Map) ForEach(fn func(key string, binding Binding) error) error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	for _, key := range m.Keys() {
		binding, exists := m.Entries[key]
		if !exists {
			return fmt.Errorf("invalid map storage")
		}

		if err := fn(key, binding); err != nil {
			return err
		}
	}

	return nil
}

func (m *Map) Keys() []string {
	if m == nil {
		return []string{}
	}

	keys := make([]string, 0, len(m.Entries))
	for key := range m.Entries {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func (m *Map) Clear() error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	if m.IsFrozen {
		return fmt.Errorf("cannot modify frozen map")
	}

	m.Entries = make(map[string]Binding)
	return nil
}

func mapKeyString(key Value) (string, error) {
	key = resolveSpecializedValue(key)

	if key.Kind != ValueString || key.Text == nil {
		return "", fmt.Errorf("map index must be a string")
	}

	return key.Text.String(), nil
}

func (state *valueFormatState) formatMap(m *Map) string {
	if m == nil || m.Len().Sign() == 0 {
		return "{:}"
	}

	if state.maps[m] {
		return "<cyclical map>"
	}

	state.maps[m] = true
	defer delete(state.maps, m)

	parts := make([]string, 0, len(m.Entries))

	if err := m.ForEach(func(key string, binding Binding) error {
		parts = append(parts, strconv.Quote(key)+": "+state.formatValue(binding.Value))
		return nil
	}); err != nil {
		return "<invalid map>"
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func sortedMapKeys(m *Map) []string {
	return m.Keys()
}
