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

func NewMap(entries map[string]Binding, isFrozen bool) *Map {
	if entries == nil {
		entries = map[string]Binding{}
	}

	return &Map{
		Entries:  entries,
		IsFrozen: isFrozen,
	}
}

func NewMapValue(entries map[string]Binding, isFrozen bool) Value {
	return Value{
		Kind: ValueMap,
		Map:  NewMap(entries, isFrozen),
	}
}

func (m *Map) EntryCount() int {
	if m == nil || m.Entries == nil {
		return 0
	}

	return len(m.Entries)
}

func (m *Map) Len() *Number {
	return NewSmallNumber(int64(m.EntryCount()))
}

func (m *Map) IsEmpty() bool {
	return m.EntryCount() == 0
}

func (m *Map) Get(key string) (Binding, bool) {
	if m == nil || m.Entries == nil {
		return Binding{}, false
	}

	binding, exists := m.Entries[key]
	return binding, exists
}

func (m *Map) Has(key string) bool {
	_, exists := m.Get(key)
	return exists
}

func (m *Map) Set(key string, binding Binding) error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	if m.Entries == nil {
		m.Entries = map[string]Binding{}
	}

	m.Entries[key] = binding
	return nil
}

func (m *Map) Clear() error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	m.Entries = map[string]Binding{}
	return nil
}

func (m *Map) Keys() []string {
	if m == nil || m.Entries == nil {
		return []string{}
	}

	keys := make([]string, 0, len(m.Entries))

	for key := range m.Entries {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (m *Map) ForEach(fn func(key string, binding Binding) error) error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	for _, key := range m.Keys() {
		binding, exists := m.Get(key)
		if !exists {
			continue
		}

		if err := fn(key, binding); err != nil {
			return err
		}
	}

	return nil
}

func (state *valueFormatState) formatMap(m *Map) string {
	if m == nil {
		return "{:}"
	}

	prefix := ""
	if m.IsFrozen {
		prefix = "~"
	}

	if m.IsEmpty() {
		return prefix + "{:}"
	}

	if state.maps[m] {
		return prefix + "<cyclical map>"
	}

	state.maps[m] = true
	defer delete(state.maps, m)

	keys := sortedMapKeys(m)

	parts := make([]string, 0, len(keys))

	for _, key := range keys {
		binding, _ := m.Get(key)
		parts = append(parts, strconv.Quote(key)+": "+state.formatValue(binding.Value))
	}

	return prefix + "{" + strings.Join(parts, ", ") + "}"
}

func sortedMapKeys(m *Map) []string {
	return m.Keys()
}
