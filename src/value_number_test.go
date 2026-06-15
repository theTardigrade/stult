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
