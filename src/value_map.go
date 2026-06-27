package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// mapNativeEntryLimit keeps ordinary and large practical maps in Go's
// optimized native map representation. After this point, new keys use the
// trie overflow tier so a Stult map does not depend entirely on one host map.
const mapNativeEntryLimit = 1 << 20

type Map struct {
	Entries       map[string]Binding
	overflow      *mapTrie
	entryCount    *Number
	IsFrozen      bool
	ValueContract *BindingContract
}

type mapTrie struct {
	root *mapNode
}

type mapNode struct {
	binding    Binding
	hasBinding bool
	children   map[rune]*mapNode
}

func NewMap(entries map[string]Binding, isFrozen bool) *Map {
	m := &Map{
		Entries:    map[string]Binding{},
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

func newMapTrie() *mapTrie {
	return &mapTrie{root: &mapNode{}}
}

func (m *Map) ensureStorage() {
	if m.Entries == nil {
		m.Entries = map[string]Binding{}
	}

	if m.entryCount == nil {
		m.entryCount = NewSmallNumber(int64(len(m.Entries)))
	}
}

func (m *Map) ensureOverflow() *mapTrie {
	if m.overflow == nil {
		m.overflow = newMapTrie()
	}

	return m.overflow
}

func (m *Map) Len() *Number {
	if m == nil {
		return NewSmallNumber(0)
	}

	m.ensureStorage()

	return CloneNumber(m.entryCount)
}

func (m *Map) IsEmpty() bool {
	return m == nil || m.Len().Sign() == 0
}

func (m *Map) Get(key string) (Binding, bool) {
	if m == nil {
		return Binding{}, false
	}

	m.ensureStorage()

	if binding, exists := m.Entries[key]; exists {
		return binding, true
	}

	if m.overflow == nil {
		return Binding{}, false
	}

	return m.overflow.Get(key)
}

func (m *Map) Has(key string) bool {
	_, exists := m.Get(key)
	return exists
}

func (m *Map) Set(key string, binding Binding) error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	if m.ValueContract != nil {
		if err := m.ValueContract.CheckAndLearn(fmt.Sprintf("map entry %q", key), binding.Value); err != nil {
			return err
		}
	}

	m.ensureStorage()

	if _, exists := m.Entries[key]; exists {
		m.Entries[key] = binding
		return nil
	}

	if m.overflow != nil {
		if _, exists := m.overflow.Get(key); exists {
			m.overflow.Set(key, binding)
			return nil
		}
	}

	if len(m.Entries) < mapNativeEntryLimit {
		m.Entries[key] = binding
	} else {
		m.ensureOverflow().Set(key, binding)
	}

	m.entryCount = numberAdd(m.entryCount, NewSmallNumber(1))
	return nil
}

func (m *Map) Clear() error {
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	m.Entries = map[string]Binding{}
	m.overflow = nil
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

	m.ensureStorage()

	nativeKeys := make([]string, 0, len(m.Entries))
	for key := range m.Entries {
		nativeKeys = append(nativeKeys, key)
	}
	sort.Strings(nativeKeys)

	if m.overflow == nil || m.overflow.IsEmpty() {
		for _, key := range nativeKeys {
			if err := fn(key, m.Entries[key]); err != nil {
				return err
			}
		}

		return nil
	}

	nativeIndex := 0

	if err := m.overflow.ForEach(func(key string, binding Binding) error {
		for nativeIndex < len(nativeKeys) && nativeKeys[nativeIndex] < key {
			nativeKey := nativeKeys[nativeIndex]
			if err := fn(nativeKey, m.Entries[nativeKey]); err != nil {
				return err
			}

			nativeIndex++
		}

		if err := fn(key, binding); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	for nativeIndex < len(nativeKeys) {
		nativeKey := nativeKeys[nativeIndex]
		if err := fn(nativeKey, m.Entries[nativeKey]); err != nil {
			return err
		}

		nativeIndex++
	}

	return nil
}

func (t *mapTrie) IsEmpty() bool {
	return t == nil || t.root == nil || (!t.root.hasBinding && len(t.root.children) == 0)
}

func (t *mapTrie) Get(key string) (Binding, bool) {
	if t == nil || t.root == nil {
		return Binding{}, false
	}

	node := t.root
	for _, r := range key {
		if node.children == nil {
			return Binding{}, false
		}

		next := node.children[r]
		if next == nil {
			return Binding{}, false
		}

		node = next
	}

	if !node.hasBinding {
		return Binding{}, false
	}

	return node.binding, true
}

func (t *mapTrie) Set(key string, binding Binding) bool {
	if t.root == nil {
		t.root = &mapNode{}
	}

	node := t.root
	for _, r := range key {
		if node.children == nil {
			node.children = map[rune]*mapNode{}
		}

		next := node.children[r]
		if next == nil {
			next = &mapNode{}
			node.children[r] = next
		}

		node = next
	}

	inserted := !node.hasBinding
	node.binding = binding
	node.hasBinding = true
	return inserted
}

func (t *mapTrie) ForEach(fn func(key string, binding Binding) error) error {
	if t == nil || t.root == nil {
		return nil
	}

	return t.root.forEach(nil, fn)
}

func (n *mapNode) forEach(prefix []rune, fn func(key string, binding Binding) error) error {
	if n == nil {
		return nil
	}

	if n.hasBinding {
		if err := fn(string(prefix), n.binding); err != nil {
			return err
		}
	}

	if len(n.children) == 0 {
		return nil
	}

	children := make([]rune, 0, len(n.children))
	for r := range n.children {
		children = append(children, r)
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i] < children[j]
	})

	for _, r := range children {
		if err := n.children[r].forEach(append(prefix, r), fn); err != nil {
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

	parts := []string{}

	_ = m.ForEach(func(key string, binding Binding) error {
		parts = append(parts, strconv.Quote(key)+": "+state.formatValue(binding.Value))
		return nil
	})

	return prefix + "{" + strings.Join(parts, ", ") + "}"
}

func sortedMapKeys(m *Map) []string {
	return m.Keys()
}
