package main

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

const maxHostInt = int(^uint(0) >> 1)

type Map struct {
	root       *mapNode
	entryCount *Number
	IsFrozen   bool
}

type mapNode struct {
	binding    Binding
	hasBinding bool
	children   map[rune]*mapNode
}

func NewMap(entries map[string]Binding, isFrozen bool) *Map {
	m := &Map{
		root:       &mapNode{},
		entryCount: NewSmallNumber(0),
		IsFrozen:   isFrozen,
	}

	for key, binding := range entries {
		_ = m.Set(key, binding)
	}

	return m
}

func NewMapValue(entries map[string]Binding, isFrozen bool) Value {
	return Value{
		Kind: ValueMap,
		Map:  NewMap(entries, isFrozen),
	}
}

func (m *Map) ensureRoot() {
	if m.root == nil {
		m.root = &mapNode{}
	}

	if m.entryCount == nil {
		m.entryCount = NewSmallNumber(0)
	}
}

func (m *Map) EntryCount() int {
	if m == nil {
		return 0
	}

	count := m.Len()
	count64, accuracy := count.Int64()
	if accuracy == big.Above || count64 > int64(maxHostInt) {
		return maxHostInt
	}

	if accuracy == big.Below || count64 < 0 {
		return 0
	}

	return int(count64)
}

func (m *Map) Len() *Number {
	if m == nil || m.entryCount == nil {
		return NewSmallNumber(0)
	}

	return CloneNumber(m.entryCount)
}

func (m *Map) IsEmpty() bool {
	return m == nil || m.Len().Sign() == 0
}

func (m *Map) Get(key string) (Binding, bool) {
	if m == nil || m.root == nil {
		return Binding{}, false
	}

	node := m.root
	for _, r := range key {
		if node.children == nil {
			return Binding{}, false
		}

		child := node.children[r]
		if child == nil {
			return Binding{}, false
		}

		node = child
	}

	if !node.hasBinding {
		return Binding{}, false
	}

	return node.binding, true
}

func (m *Map) Has(key string) bool {
	_, exists := m.Get(key)
	return exists
}

func (m *Map) Set(key string, binding Binding) error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	m.ensureRoot()

	node := m.root
	for _, r := range key {
		if node.children == nil {
			node.children = map[rune]*mapNode{}
		}

		child := node.children[r]
		if child == nil {
			child = &mapNode{}
			node.children[r] = child
		}

		node = child
	}

	if !node.hasBinding {
		m.entryCount = numberAdd(m.entryCount, NewSmallNumber(1))
	}

	node.binding = binding
	node.hasBinding = true
	return nil
}

func (m *Map) Clear() error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	m.root = &mapNode{}
	m.entryCount = NewSmallNumber(0)
	return nil
}

func (m *Map) Keys() []string {
	keys := []string{}

	if m == nil {
		return keys
	}

	_ = m.ForEach(func(key string, _ Binding) error {
		keys = append(keys, key)
		return nil
	})

	return keys
}

func (m *Map) ForEach(fn func(key string, binding Binding) error) error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	if m.root == nil {
		return nil
	}

	return m.root.forEach(nil, fn)
}

func (node *mapNode) forEach(prefix []rune, fn func(key string, binding Binding) error) error {
	if node == nil {
		return nil
	}

	if node.hasBinding {
		if err := fn(string(prefix), node.binding); err != nil {
			return err
		}
	}

	if len(node.children) == 0 {
		return nil
	}

	runes := make([]rune, 0, len(node.children))
	for r := range node.children {
		runes = append(runes, r)
	}

	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	for _, r := range runes {
		if err := node.children[r].forEach(append(prefix, r), fn); err != nil {
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
