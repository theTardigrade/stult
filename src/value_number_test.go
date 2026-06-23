package main

import (
	"strings"
	"testing"
)

func TestHugeIntegerAddition(t *testing.T) {
	huge := "1" + strings.Repeat("0", 800)

	left, err := NewNumberFromString(huge)
	if err != nil {
		t.Fatal(err)
	}

	out := numberAdd(left, NewSmallNumber(1))

	got := out.Format(MaxDecimalPlaces)
	want := huge[:len(huge)-1] + "1"

	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestHugeIntegerWithDecimalAddition(t *testing.T) {
	nines := strings.Repeat("9", 800)

	left, err := NewNumberFromString(nines + ".25")
	if err != nil {
		t.Fatal(err)
	}

	right, err := NewNumberFromString("0.75")
	if err != nil {
		t.Fatal(err)
	}

	out := numberAdd(left, right)

	got := out.Format(MaxDecimalPlaces)
	want := "1" + strings.Repeat("0", 800)

	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDivisionRoundsToMaxDecimalPlaces(t *testing.T) {
	one := NewSmallNumber(1)
	three := NewSmallNumber(3)

	out, err := numberDivide(one, three)
	if err != nil {
		t.Fatal(err)
	}

	got := out.Format(MaxDecimalPlaces)

	if !strings.HasPrefix(got, "0.") {
		t.Fatalf("got %q, want decimal beginning with 0.", got)
	}

	fraction := strings.TrimPrefix(got, "0.")
	if len(fraction) != MaxDecimalPlaces {
		t.Fatalf("got %d decimal places, want %d", len(fraction), MaxDecimalPlaces)
	}

	for _, digit := range fraction {
		if digit != '3' {
			t.Fatalf("got %q, want only recurring 3s", got)
		}
	}
}

func TestMultiplicationRoundsToMaxDecimalPlaces(t *testing.T) {
	leftLiteral := "0." + strings.Repeat("1", 200)
	rightLiteral := "0." + strings.Repeat("2", 200)

	left, err := NewNumberFromString(leftLiteral)
	if err != nil {
		t.Fatal(err)
	}

	right, err := NewNumberFromString(rightLiteral)
	if err != nil {
		t.Fatal(err)
	}

	out := numberMultiply(left, right)

	got := out.Format(MaxDecimalPlaces)

	if strings.Contains(got, ".") {
		fraction := strings.SplitN(got, ".", 2)[1]
		if len(fraction) > MaxDecimalPlaces {
			t.Fatalf("got %d decimal places, want at most %d", len(fraction), MaxDecimalPlaces)
		}
	}
}

func TestHugeDecimalComparisonPreservesWholePart(t *testing.T) {
	huge := "1" + strings.Repeat("0", 800)

	left, err := NewNumberFromString(huge + ".1")
	if err != nil {
		t.Fatal(err)
	}

	right, err := NewNumberFromString(strings.Repeat("9", 800) + ".9")
	if err != nil {
		t.Fatal(err)
	}

	if numberCompare(left, right) <= 0 {
		t.Fatalf("expected %s to be greater than %s", left.Format(MaxDecimalPlaces), right.Format(MaxDecimalPlaces))
	}
}

func TestPercentageNumberLiteralScalesByHundred(t *testing.T) {
	tests := []struct {
		literal string
		want    string
	}{
		{"50%", "0.5"},
		{"99.9%", "0.999"},
		{".5%", "0.005"},
		{"1e2%", "1"},
	}

	for _, tt := range tests {
		number, err := NewNumberFromString(tt.literal)
		if err != nil {
			t.Fatalf("NewNumberFromString(%q) returned error: %v", tt.literal, err)
		}

		got := number.Format(MaxDecimalPlaces)
		if got != tt.want {
			t.Fatalf("NewNumberFromString(%q) = %q, want %q", tt.literal, got, tt.want)
		}
	}
}

func TestApostropheNumberSeparators(t *testing.T) {
	tests := []struct {
		literal string
		want    string
	}{
		{"1'000", "1000"},
		{"1'000'000'000", "1000000000"},
		{"123'456.789'123", "123456.789123"},
		{".123'456", "0.123456"},
		{"1'000%", "10"},
		{"1e1'0", "10000000000"},
	}

	for _, tt := range tests {
		number, err := NewNumberFromString(tt.literal)
		if err != nil {
			t.Fatalf("NewNumberFromString(%q) returned error: %v", tt.literal, err)
		}

		got := number.Format(MaxDecimalPlaces)
		if got != tt.want {
			t.Fatalf("NewNumberFromString(%q) = %q, want %q", tt.literal, got, tt.want)
		}
	}
}

func TestInvalidApostropheNumberSeparators(t *testing.T) {
	invalid := []string{
		"'1000",
		"1000'",
		"1''000",
		"1'.5",
		"1.'5",
		"1e'3",
		"1e3'",
		"50'%",
	}

	for _, literal := range invalid {
		if _, err := NewNumberFromString(literal); err == nil {
			t.Fatalf("NewNumberFromString(%q) returned nil error", literal)
		}
	}
}
