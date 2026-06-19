package main

import (
	"sort"
	"strconv"
	"strings"
)

type Map struct {
	Entries     map[string]Binding
	IsImmutable bool
}

func NewMapValue(entries map[string]Binding, isImmutable bool) Value {
	return Value{
		Kind: ValueMap,
		Map: &Map{
			Entries:     entries,
			IsImmutable: isImmutable,
		},
	}
}

func (state *valueFormatState) formatMap(m *Map) string {
	if m == nil || len(m.Entries) == 0 {
		return "{:}"
	}

	if state.maps[m] {
		return "<cyclical map>"
	}

	state.maps[m] = true
	defer delete(state.maps, m)

	keys := sortedMapKeys(m)

	parts := make([]string, 0, len(keys))

	for _, key := range keys {
		binding := m.Entries[key]
		parts = append(parts, strconv.Quote(key)+": "+state.formatValue(binding.Value))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func sortedMapKeys(m *Map) []string {
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
