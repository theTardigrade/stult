package main

import "math/big"

func NewStdTypesNumberMap() Value {
	entries := map[string]Binding{
		"PRECISION":        NewImmutableBinding(NewNumberValueFromInt(int(FloatPrecision))),
		"FRACTION_DIGITS":  NewImmutableBinding(NewNumberValueFromInt(DefaultFractionDigits)),
		"MAX_SAFE_INTEGER": NewImmutableBinding(newNumberValueFromBigInt(maxSafeIntegerBigInt())),
		"MIN_SAFE_INTEGER": NewImmutableBinding(newNumberValueFromBigInt(minSafeIntegerBigInt())),
	}

	return NewMapValue(entries, true)
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
