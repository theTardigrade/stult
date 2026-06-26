package main

import "testing"

func TestForEachStultRangeValueStreamsValues(t *testing.T) {
	values := []string{}

	produced, err := forEachStultRangeValue(
		NewNumberValueFromInt(1),
		NewNumberValueFromInt(3),
		NewVoidValue(),
		true,
		func(value Value) error {
			values = append(values, value.Number.Format(DefaultDecimalPlacesToDisplay))
			return nil
		},
	)
	if err != nil {
		t.Fatalf("forEachStultRangeValue() failed: %v", err)
	}

	if !produced {
		t.Fatalf("expected produced=true")
	}

	want := []string{"1", "2", "3"}
	if len(values) != len(want) {
		t.Fatalf("got %d values, want %d", len(values), len(want))
	}

	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("value[%d] = %q, want %q", index, values[index], want[index])
		}
	}
}

func TestForEachStultRangeValueReportsEmptyRange(t *testing.T) {
	produced, err := forEachStultRangeValue(
		NewNumberValueFromInt(1),
		NewNumberValueFromInt(1),
		NewVoidValue(),
		false,
		func(value Value) error {
			t.Fatalf("empty range should not produce value %v", value)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("forEachStultRangeValue() failed: %v", err)
	}

	if produced {
		t.Fatalf("expected produced=false for empty exclusive range")
	}
}
