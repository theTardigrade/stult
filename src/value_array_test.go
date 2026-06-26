package main

import (
	"math/big"
	"testing"
)

func mustArrayElement(t *testing.T, array *Array, index int64) Value {
	t.Helper()

	value, exists, err := array.Get(NewSmallNumber(index))
	if err != nil {
		t.Fatalf("Get(%d) failed: %v", index, err)
	}

	if !exists {
		t.Fatalf("expected element at index %d", index)
	}

	return value
}

func mustArrayString(t *testing.T, array *Array, index int64) string {
	t.Helper()

	value := mustArrayElement(t, array, index)
	if value.Kind != ValueString || value.Text == nil {
		t.Fatalf("expected string element at index %d, got %v", index, value.Kind)
	}

	return value.Text.String()
}

func TestNewArrayValueDoesNotAliasInputSlice(t *testing.T) {
	source := []Value{NewStringValue("original")}
	array := NewArrayValue(source, false).Array

	source[0] = NewStringValue("changed")

	if got := mustArrayString(t, array, 0); got != "original" {
		t.Fatalf("expected array to keep original element, got %q", got)
	}
}

func TestArrayCapacityHintIsBounded(t *testing.T) {
	array := NewArrayValue([]Value{
		NewNumberValueFromInt(1),
		NewNumberValueFromInt(2),
		NewNumberValueFromInt(3),
	}, false).Array

	if got := array.capacityHintHostLimited(10); got != 3 {
		t.Fatalf("expected capacity hint 3, got %d", got)
	}

	if got := array.capacityHintHostLimited(2); got != 2 {
		t.Fatalf("expected bounded capacity hint 2, got %d", got)
	}

	if got := array.capacityHintHostLimited(0); got != 0 {
		t.Fatalf("expected zero capacity hint for zero limit, got %d", got)
	}

	largeLength := new(big.Int).Mul(arrayOrdinaryLimitBig, big.NewInt(2))
	largeArray := &Array{Length: NewBigIntNumber(largeLength)}
	if got := largeArray.capacityHintHostLimited(32); got != 32 {
		t.Fatalf("expected large array capacity hint to be capped at 32, got %d", got)
	}
}

func TestArraySetAppendAndOutOfBounds(t *testing.T) {
	array := NewArrayValue([]Value{NewStringValue("a")}, false).Array

	if err := array.Set(NewSmallNumber(1), NewStringValue("b")); err != nil {
		t.Fatalf("Set at append position failed: %v", err)
	}

	if got := array.Len().Format(DefaultDecimalPlacesToDisplay); got != "2" {
		t.Fatalf("expected array length 2 after append-position set, got %s", got)
	}

	if got := mustArrayString(t, array, 1); got != "b" {
		t.Fatalf("expected appended value %q, got %q", "b", got)
	}

	if err := array.Set(NewSmallNumber(3), NewStringValue("c")); err == nil {
		t.Fatalf("expected Set past next append position to fail")
	}
}

func TestFrozenArrayRejectsMutation(t *testing.T) {
	array := NewArrayValue([]Value{NewStringValue("a")}, true).Array

	if err := array.Set(NewSmallNumber(0), NewStringValue("b")); err == nil {
		t.Fatalf("expected Set on frozen array to fail")
	}

	if err := array.Append(NewStringValue("b")); err == nil {
		t.Fatalf("expected Append on frozen array to fail")
	}

	if err := array.Clear(); err == nil {
		t.Fatalf("expected Clear on frozen array to fail")
	}

	if got := mustArrayString(t, array, 0); got != "a" {
		t.Fatalf("expected frozen array to remain unchanged, got %q", got)
	}
}

func TestArrayForEachPreservesOrder(t *testing.T) {
	array := NewArrayValue([]Value{
		NewStringValue("a"),
		NewStringValue("b"),
		NewStringValue("c"),
	}, false).Array

	var indexes []string
	var values []string

	if err := array.ForEach(func(index *Number, value Value) error {
		indexes = append(indexes, index.Format(DefaultDecimalPlacesToDisplay))
		values = append(values, value.Text.String())
		return nil
	}); err != nil {
		t.Fatalf("ForEach failed: %v", err)
	}

	expectedIndexes := []string{"0", "1", "2"}
	expectedValues := []string{"a", "b", "c"}

	for i := range expectedIndexes {
		if indexes[i] != expectedIndexes[i] || values[i] != expectedValues[i] {
			t.Fatalf("iteration %d = (%s, %s), want (%s, %s)", i, indexes[i], values[i], expectedIndexes[i], expectedValues[i])
		}
	}
}

func TestArrayOverflowStorageAcrossOrdinaryBoundary(t *testing.T) {
	overflow := &ArrayOverflow{Chunks: make(map[string][]Value)}
	firstOverflowIndex := new(big.Int).Set(arrayOrdinaryLimitBig)
	secondOverflowIndex := new(big.Int).Add(firstOverflowIndex, arrayOneBig)

	if err := overflow.AppendAt(firstOverflowIndex, NewStringValue("first")); err != nil {
		t.Fatalf("AppendAt first overflow index failed: %v", err)
	}

	if err := overflow.AppendAt(secondOverflowIndex, NewStringValue("second")); err != nil {
		t.Fatalf("AppendAt second overflow index failed: %v", err)
	}

	if err := overflow.SetExisting(secondOverflowIndex, NewStringValue("updated")); err != nil {
		t.Fatalf("SetExisting overflow index failed: %v", err)
	}

	value, exists, err := overflow.Get(secondOverflowIndex)
	if err != nil {
		t.Fatalf("Get overflow index failed: %v", err)
	}
	if !exists || value.Text.String() != "updated" {
		t.Fatalf("expected updated overflow value, got exists=%v value=%q", exists, value.Text.String())
	}

	var indexes []string
	var values []string
	length := new(big.Int).Add(arrayOrdinaryLimitBig, big.NewInt(2))
	if err := overflow.ForEach(length, func(index *Number, value Value) error {
		indexes = append(indexes, index.Format(DefaultDecimalPlacesToDisplay))
		values = append(values, value.Text.String())
		return nil
	}); err != nil {
		t.Fatalf("ForEach overflow failed: %v", err)
	}

	if len(indexes) != 2 || values[0] != "first" || values[1] != "updated" {
		t.Fatalf("unexpected overflow iteration indexes=%v values=%v", indexes, values)
	}
}
