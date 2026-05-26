package main

import (
	"fmt"
	"math/big"
	"strings"
)

func NewStdTypesNumberMap() Value {
	entries := map[string]Binding{
		"FRACTION_DIGITS":  NewImmutableBinding(NewNumberValueFromInt(DefaultFractionDigits)),
		"MAX_SAFE_INTEGER": NewImmutableBinding(newNumberValueFromBigInt(maxSafeIntegerBigInt())),
		"MIN_SAFE_INTEGER": NewImmutableBinding(newNumberValueFromBigInt(minSafeIntegerBigInt())),
		"NEW":              NewImmutableBinding(NewBuiltinFunctionValue(stdTypesNumberNew)),
		"PRECISION":        NewImmutableBinding(NewNumberValueFromInt(int(FloatPrecision))),
	}

	return NewMapValue(entries, true)
}

func stdTypesNumberNew(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.NUMBER.NEW expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueNumber:
		return Value{
			Kind:   ValueNumber,
			Number: CloneNumber(value.Number),
		}, nil

	case ValueBool:
		if value.Bool {
			return NewNumberValueFromInt(1), nil
		}

		return NewNumberValueFromInt(0), nil

	case ValueString:
		if value.Text == nil {
			return NewVoidValue(), nil
		}

		text := strings.TrimSpace(value.Text.String())
		if text == "" {
			return NewVoidValue(), nil
		}

		number, err := NewNumberValueFromString(text)
		if err != nil {
			return NewVoidValue(), nil
		}

		return number, nil

	case ValueVoid,
		ValueMap,
		ValueArray,
		ValueEmptyCollection,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPES.NUMBER.NEW cannot convert unknown value kind")
	}
}

func maxSafeIntegerBigInt() *big.Int {
	one := big.NewInt(1)

	max := new(big.Int).Lsh(one, FloatPrecision)
	max.Sub(max, one)

	return max
}

func minSafeIntegerBigInt() *big.Int {
	min := maxSafeIntegerBigInt()
	min.Neg(min)

	return min
}

func newNumberValueFromBigInt(value *big.Int) Value {
	number := newFloat()
	number.SetInt(value)

	return Value{
		Kind:   ValueNumber,
		Number: number,
	}
}
