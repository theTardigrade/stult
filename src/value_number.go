package main

import (
	"fmt"
	"math/big"
	"strings"
)

const FloatPrecision uint = 2048
const DefaultFractionDigits = 32

func NewNumberValueFromString(literal string) (Value, error) {
	n, _, err := big.ParseFloat(literal, 10, FloatPrecision, big.ToNearestEven)
	if err != nil {
		return Value{}, fmt.Errorf("invalid number %q", literal)
	}

	return Value{Kind: ValueNumber, Number: n}, nil
}

func NewNumberValueFromInt(value int) Value {
	n := newFloat().
		SetInt64(int64(value))

	return Value{Kind: ValueNumber, Number: n}
}

func NewNumberValueFromInt64(value int64) Value {
	n := newFloat().
		SetInt64(value)

	return Value{Kind: ValueNumber, Number: n}
}

func CloneNumber(x *big.Float) *big.Float {
	return newFloat().
		Set(x)
}

func newFloat() *big.Float {
	return newFloatWithPrecision(FloatPrecision)
}

func newFloatWithPrecision(precision uint) *big.Float {
	return new(big.Float).
		SetPrec(precision).
		SetMode(big.ToNearestEven)
}

func formatNumber(x *big.Float, fractionDigits int) string {
	if fractionDigits < 0 {
		fractionDigits = DefaultFractionDigits
	}

	text := x.Text('f', fractionDigits)

	return trimDecimalZeros(text)
}

func trimDecimalZeros(text string) string {
	if !strings.Contains(text, ".") {
		return text
	}

	text = strings.TrimRight(text, "0")
	text = strings.TrimRight(text, ".")

	if text == "-0" {
		return "0"
	}

	return text
}

func numberToArrayIndex(index Value) (int, error) {
	if index.Kind != ValueNumber {
		return 0, fmt.Errorf("array index must be a number")
	}

	i, accuracy := index.Number.Int64()
	if accuracy != big.Exact {
		return 0, fmt.Errorf("array index must be an integer")
	}

	return int(i), nil
}

func numberToInt64(value Value, name string) (int64, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueNumber {
		return 0, fmt.Errorf("%s must be a number", name)
	}

	out, accuracy := value.Number.Int64()
	if accuracy != big.Exact {
		return 0, fmt.Errorf("%s must be an integer", name)
	}

	return out, nil
}
