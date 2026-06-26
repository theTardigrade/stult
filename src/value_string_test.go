package main

import "testing"

func TestStringStoresNativeOnlyFromNativeString(t *testing.T) {
	value := NewStringValue("cat")
	text := value.Text

	if text == nil {
		t.Fatalf("expected string value")
	}

	if text.storageState != stringStorageNative {
		t.Fatalf("native string constructor storage = %v, want %v", text.storageState, stringStorageNative)
	}

	if text.native != "cat" {
		t.Fatalf("native = %q, want %q", text.native, "cat")
	}

	if text.runes != nil {
		t.Fatalf("native string constructor should not populate runes")
	}

	if got := text.String(); got != "cat" {
		t.Fatalf("String() = %q, want %q", got, "cat")
	}

	if text.storageState != stringStorageNative {
		t.Fatalf("String() should keep native-only storage when native already exists")
	}
}

func TestStringStoresRunesOnlyFromRunes(t *testing.T) {
	value := NewStringValueWithFrozenFromRunes([]rune{'d', 'o', 'g'}, false)
	text := value.Text

	if text == nil {
		t.Fatalf("expected string value")
	}

	if text.storageState != stringStorageRunes {
		t.Fatalf("rune constructor storage = %v, want %v", text.storageState, stringStorageRunes)
	}

	if text.native != "" {
		t.Fatalf("rune constructor should not populate native string; got %q", text.native)
	}

	if got := text.String(); got != "dog" {
		t.Fatalf("String() = %q, want %q", got, "dog")
	}

	if text.storageState != stringStorageBoth {
		t.Fatalf("String() should populate native storage and keep runes")
	}

	if text.native != "dog" {
		t.Fatalf("native = %q, want %q", text.native, "dog")
	}
}

func TestStringRuneAccessMaterializesRunesFromNative(t *testing.T) {
	value := NewStringValue("cat")
	text := value.Text

	r, ok, err := text.Get(1)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if !ok || r != 'a' {
		t.Fatalf("Get(1) = %q, %v; want %q, true", r, ok, 'a')
	}

	if text.storageState != stringStorageBoth {
		t.Fatalf("Get() should materialize runes and keep native storage")
	}

	if string(text.runes) != "cat" {
		t.Fatalf("runes = %q, want %q", string(text.runes), "cat")
	}
}

func TestStringRuneCountDoesNotMaterializeRunes(t *testing.T) {
	value := NewStringValue("a🐶c")
	text := value.Text

	if got := text.RuneCount(); got != 3 {
		t.Fatalf("RuneCount() = %d, want 3", got)
	}

	if text.storageState != stringStorageNative {
		t.Fatalf("RuneCount() should not materialize runes")
	}

	if text.runes != nil {
		t.Fatalf("RuneCount() should not populate runes")
	}
}

func TestStringNativeInvalidatesOnMutation(t *testing.T) {
	value := NewStringValue("cat")
	text := value.Text

	if err := text.Set(0, 'b'); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	if text.storageState != stringStorageRunes {
		t.Fatalf("Set() should leave rune-only storage after invalidating native")
	}

	if text.native != "" {
		t.Fatalf("Set() should clear native storage; got %q", text.native)
	}

	if got := text.String(); got != "bat" {
		t.Fatalf("String() after Set() = %q, want %q", got, "bat")
	}

	if text.storageState != stringStorageBoth || text.native != "bat" {
		t.Fatalf("String() should refresh native storage to %q", "bat")
	}

	if err := text.Clear(); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	if text.storageState != stringStorageRunes {
		t.Fatalf("Clear() should leave rune-only storage")
	}

	if got := text.String(); got != "" {
		t.Fatalf("String() after Clear() = %q, want empty string", got)
	}

	if text.storageState != stringStorageBoth || text.native != "" {
		t.Fatalf("String() should refresh native storage to empty string")
	}
}

func TestStringCloneRunesDoesNotAliasStorage(t *testing.T) {
	value := NewStringValue("cat")
	text := value.Text

	runes := text.CloneRunes()
	runes[0] = 'b'

	if got := text.String(); got != "cat" {
		t.Fatalf("CloneRunes() should not expose mutable storage; got %q", got)
	}
}

func TestStringForEachPreservesRuneOrder(t *testing.T) {
	value := NewStringValue("abc")
	text := value.Text

	seen := []rune{}
	if err := text.ForEach(func(index int, r rune) error {
		if index != len(seen) {
			t.Fatalf("ForEach index = %d, want %d", index, len(seen))
		}

		seen = append(seen, r)
		return nil
	}); err != nil {
		t.Fatalf("ForEach() failed: %v", err)
	}

	if string(seen) != "abc" {
		t.Fatalf("ForEach runes = %q, want %q", string(seen), "abc")
	}
}
