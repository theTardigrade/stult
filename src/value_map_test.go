package main

import (
	"fmt"
	"testing"
)

func mutableTestBinding(value Value) Binding {
	return Binding{Value: value}
}

func TestMapLenTracksExactEntryCount(t *testing.T) {
	m := NewMap(nil, false)

	if got := m.Len().Format(DefaultDecimalPlacesToDisplay); got != "0" {
		t.Fatalf("expected empty map length 0, got %s", got)
	}

	if err := m.Set("name", mutableTestBinding(NewStringValue("Ada"))); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if got := m.Len().Format(DefaultDecimalPlacesToDisplay); got != "1" {
		t.Fatalf("expected map length 1 after inserting new key, got %s", got)
	}

	if err := m.Set("name", mutableTestBinding(NewStringValue("Grace"))); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if got := m.Len().Format(DefaultDecimalPlacesToDisplay); got != "1" {
		t.Fatalf("expected map length to stay 1 after replacing existing key, got %s", got)
	}

	if err := m.Set("language", mutableTestBinding(NewStringValue("Stult"))); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if got := m.Len().Format(DefaultDecimalPlacesToDisplay); got != "2" {
		t.Fatalf("expected map length 2 after second key, got %s", got)
	}
}

func TestMapClearResetsExactEntryCount(t *testing.T) {
	m := NewMap(map[string]Binding{
		"a": mutableTestBinding(NewNumberValueFromInt(1)),
		"b": mutableTestBinding(NewNumberValueFromInt(2)),
	}, false)

	if got := m.Len().Format(DefaultDecimalPlacesToDisplay); got != "2" {
		t.Fatalf("expected starting map length 2, got %s", got)
	}

	if err := m.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if got := m.Len().Format(DefaultDecimalPlacesToDisplay); got != "0" {
		t.Fatalf("expected cleared map length 0, got %s", got)
	}

	if !m.IsEmpty() {
		t.Fatalf("expected cleared map to be empty")
	}
}

func TestMapKeysRemainSortedWithHybridStorage(t *testing.T) {
	m := NewMap(nil, false)

	for _, key := range []string{"car", "cat", "apple", ""} {
		if err := m.Set(key, mutableTestBinding(NewStringValue(key))); err != nil {
			t.Fatalf("Set failed for %q: %v", key, err)
		}
	}

	keys := m.Keys()
	expected := []string{"", "apple", "car", "cat"}

	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d: %v", len(expected), len(keys), keys)
	}

	for i := range expected {
		if keys[i] != expected[i] {
			t.Fatalf("expected sorted key %d to be %q, got %q", i, expected[i], keys[i])
		}
	}
}

func TestMapUsesNativeTierUntilLimitThenOverflowTier(t *testing.T) {
	m := NewMap(nil, false)

	for i := 0; i < mapNativeEntryLimit; i++ {
		key := fmt.Sprintf("native-%06d", i)
		if err := m.Set(key, mutableTestBinding(NewNumberValueFromInt(i))); err != nil {
			t.Fatalf("Set failed for %q: %v", key, err)
		}
	}

	if len(m.Entries) != mapNativeEntryLimit {
		t.Fatalf("expected native tier to hold %d entries, got %d", mapNativeEntryLimit, len(m.Entries))
	}

	if m.overflow != nil && !m.overflow.IsEmpty() {
		t.Fatalf("expected overflow tier to stay empty until native tier is full")
	}

	if err := m.Set("overflow", mutableTestBinding(NewStringValue("overflow"))); err != nil {
		t.Fatalf("Set failed for overflow key: %v", err)
	}

	if len(m.Entries) != mapNativeEntryLimit {
		t.Fatalf("expected native tier size to remain %d after overflow insert, got %d", mapNativeEntryLimit, len(m.Entries))
	}

	if m.overflow == nil || m.overflow.IsEmpty() {
		t.Fatalf("expected overflow tier to contain entries after native tier is full")
	}

	if _, exists := m.Entries["overflow"]; exists {
		t.Fatalf("expected overflow key not to be copied into native tier")
	}

	binding, exists := m.Get("overflow")
	if !exists {
		t.Fatalf("expected overflow key to be readable")
	}

	if got := binding.Value.Text.String(); got != "overflow" {
		t.Fatalf("expected overflow value, got %q", got)
	}

	if got := m.Len().Format(DefaultDecimalPlacesToDisplay); got != fmt.Sprintf("%d", mapNativeEntryLimit+1) {
		t.Fatalf("expected exact map length %d, got %s", mapNativeEntryLimit+1, got)
	}
}

func TestMapUpdatesNativeEntriesAfterOverflowWithoutCopying(t *testing.T) {
	m := NewMap(nil, false)

	for i := 0; i < mapNativeEntryLimit; i++ {
		key := fmt.Sprintf("native-%06d", i)
		if err := m.Set(key, mutableTestBinding(NewNumberValueFromInt(i))); err != nil {
			t.Fatalf("Set failed for %q: %v", key, err)
		}
	}

	if err := m.Set("overflow", mutableTestBinding(NewStringValue("overflow"))); err != nil {
		t.Fatalf("Set failed for overflow key: %v", err)
	}

	if err := m.Set("native-000000", mutableTestBinding(NewStringValue("updated"))); err != nil {
		t.Fatalf("Set failed for native replacement: %v", err)
	}

	if got := m.Len().Format(DefaultDecimalPlacesToDisplay); got != fmt.Sprintf("%d", mapNativeEntryLimit+1) {
		t.Fatalf("expected replacement not to change length, got %s", got)
	}

	if _, exists := m.overflow.Get("native-000000"); exists {
		t.Fatalf("expected existing native key to stay in native tier")
	}

	binding, exists := m.Get("native-000000")
	if !exists {
		t.Fatalf("expected updated native key to be readable")
	}

	if got := binding.Value.Text.String(); got != "updated" {
		t.Fatalf("expected updated value, got %q", got)
	}
}

func TestMapForEachMergesNativeAndOverflowTiersInSortedOrder(t *testing.T) {
	m := NewMap(nil, false)

	for i := 0; i < mapNativeEntryLimit; i++ {
		key := fmt.Sprintf("n%06d", i)
		if err := m.Set(key, mutableTestBinding(NewNumberValueFromInt(i))); err != nil {
			t.Fatalf("Set failed for %q: %v", key, err)
		}
	}

	for _, key := range []string{"a", "z"} {
		if err := m.Set(key, mutableTestBinding(NewStringValue(key))); err != nil {
			t.Fatalf("Set failed for %q: %v", key, err)
		}
	}

	keys := m.Keys()

	if len(keys) != mapNativeEntryLimit+2 {
		t.Fatalf("expected %d keys, got %d", mapNativeEntryLimit+2, len(keys))
	}

	if keys[0] != "a" {
		t.Fatalf("expected first key to be overflow key %q, got %q", "a", keys[0])
	}

	if keys[1] != "n000000" {
		t.Fatalf("expected second key to be first native key %q, got %q", "n000000", keys[1])
	}

	if keys[len(keys)-1] != "z" {
		t.Fatalf("expected last key to be overflow key %q, got %q", "z", keys[len(keys)-1])
	}
}
