package main

import "testing"

func TestStringNativeCacheFromNativeString(t *testing.T) {
	value := NewStringValue("cat")
	text := value.Text

	if text == nil {
		t.Fatalf("expected string value")
	}

	if !text.isNativeCacheValid {
		t.Fatalf("native string constructor should populate native cache")
	}

	if text.nativeCache != "cat" {
		t.Fatalf("native cache = %q, want %q", text.nativeCache, "cat")
	}

	if got := text.String(); got != "cat" {
		t.Fatalf("String() = %q, want %q", got, "cat")
	}
}

func TestStringNativeCacheFromRunes(t *testing.T) {
	value := NewStringValueWithFrozenFromRunes([]rune{'d', 'o', 'g'}, false)
	text := value.Text

	if text == nil {
		t.Fatalf("expected string value")
	}

	if text.isNativeCacheValid {
		t.Fatalf("rune constructor should leave native cache invalid")
	}

	if got := text.String(); got != "dog" {
		t.Fatalf("String() = %q, want %q", got, "dog")
	}

	if !text.isNativeCacheValid {
		t.Fatalf("String() should populate native cache")
	}

	if text.nativeCache != "dog" {
		t.Fatalf("native cache = %q, want %q", text.nativeCache, "dog")
	}
}

func TestStringNativeCacheInvalidatesOnMutation(t *testing.T) {
	value := NewStringValue("cat")
	text := value.Text

	if err := text.Set(0, 'b'); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	if text.isNativeCacheValid {
		t.Fatalf("Set() should invalidate native cache")
	}

	if got := text.String(); got != "bat" {
		t.Fatalf("String() after Set() = %q, want %q", got, "bat")
	}

	if !text.isNativeCacheValid || text.nativeCache != "bat" {
		t.Fatalf("String() should refresh native cache to %q", "bat")
	}

	if err := text.Clear(); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	if text.isNativeCacheValid {
		t.Fatalf("Clear() should invalidate native cache")
	}

	if got := text.String(); got != "" {
		t.Fatalf("String() after Clear() = %q, want empty string", got)
	}

	if !text.isNativeCacheValid || text.nativeCache != "" {
		t.Fatalf("String() should refresh native cache to empty string")
	}
}
