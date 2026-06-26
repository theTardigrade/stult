package main

import (
	"math/big"
	"testing"
)

func TestMapTrieSetGetLenAndUpdate(t *testing.T) {
	m := NewMap(nil, false)

	if !m.IsEmpty() {
		t.Fatalf("new map should be empty")
	}

	if err := m.Set("name", Binding{Value: NewStringValue("Ada")}); err != nil {
		t.Fatalf("Set(name) error: %v", err)
	}

	if err := m.Set("number", Binding{Value: NewNumberValueFromInt(1)}); err != nil {
		t.Fatalf("Set(number) error: %v", err)
	}

	if got := m.Len().String(); got != "2" {
		t.Fatalf("Len() = %s, want 2", got)
	}

	binding, exists := m.Get("name")
	if !exists {
		t.Fatalf("Get(name) did not find binding")
	}

	if got := binding.Value.Text.String(); got != "Ada" {
		t.Fatalf("Get(name) = %q, want Ada", got)
	}

	if err := m.Set("name", Binding{Value: NewStringValue("Grace")}); err != nil {
		t.Fatalf("Set(name update) error: %v", err)
	}

	if got := m.Len().String(); got != "2" {
		t.Fatalf("Len() after update = %s, want 2", got)
	}

	binding, exists = m.Get("name")
	if !exists {
		t.Fatalf("Get(name) after update did not find binding")
	}

	if got := binding.Value.Text.String(); got != "Grace" {
		t.Fatalf("Get(name) after update = %q, want Grace", got)
	}
}

func TestMapTrieKeyOrdering(t *testing.T) {
	m := NewMap(nil, false)

	for _, key := range []string{"beta", "", "alpha", "alphabet", "zebra"} {
		if err := m.Set(key, Binding{Value: NewStringValue(key)}); err != nil {
			t.Fatalf("Set(%q) error: %v", key, err)
		}
	}

	keys := m.Keys()
	want := []string{"", "alpha", "alphabet", "beta", "zebra"}

	if len(keys) != len(want) {
		t.Fatalf("len(Keys()) = %d, want %d", len(keys), len(want))
	}

	for index, key := range keys {
		if key != want[index] {
			t.Fatalf("Keys()[%d] = %q, want %q", index, key, want[index])
		}
	}
}

func TestMapTrieClear(t *testing.T) {
	m := NewMap(map[string]Binding{
		"a": Binding{Value: NewNumberValueFromInt(1)},
		"b": Binding{Value: NewNumberValueFromInt(2)},
	}, false)

	if err := m.Clear(); err != nil {
		t.Fatalf("Clear() error: %v", err)
	}

	if got := m.Len().String(); got != "0" {
		t.Fatalf("Len() after Clear() = %s, want 0", got)
	}

	if m.Has("a") || m.Has("b") {
		t.Fatalf("Clear() left old keys behind")
	}

	if err := m.Set("c", Binding{Value: NewNumberValueFromInt(3)}); err != nil {
		t.Fatalf("Set(c) after Clear() error: %v", err)
	}

	if !m.Has("c") {
		t.Fatalf("Set(c) after Clear() did not store key")
	}
}

func TestMapEntryCountCapsAtHostInt(t *testing.T) {
	m := NewMap(nil, false)
	m.entryCount = NewBigIntNumber(new(big.Int).Lsh(big.NewInt(1), 80))

	if got := m.EntryCount(); got != maxHostInt {
		t.Fatalf("EntryCount() = %d, want maxHostInt", got)
	}

	if got := m.Len().String(); got != "1208925819614629174706176" {
		t.Fatalf("Len() = %s, want exact big count", got)
	}
}
