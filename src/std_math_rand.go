package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func NewStdMathRandMap() Value {
	entries := map[string]Binding{
		"NUMBER":       NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathRandNumber)),
		"WHOLE_NUMBER": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathRandWholeNumber)),
	}

	return NewMapValue(entries, true)
}

func builtinStdMathRandNumber(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("MATH.RAND.NUMBER expected 2 arguments, got %d", len(args))
	}

	lower, err := stdMathNumberArg("MATH.RAND.NUMBER", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	upper, err := stdMathNumberArg("MATH.RAND.NUMBER", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	if numberCompare(lower.Number, upper.Number) >= 0 {
		return Value{}, fmt.Errorf(
			"MATH.RAND.NUMBER inclusive lower bound must be less than exclusive upper bound",
		)
	}

	out, err := stdMathRandNumber(lower.Number, upper.Number)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(out), nil
}

func builtinStdMathRandWholeNumber(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("MATH.RAND.WHOLE_NUMBER expected 2 arguments, got %d", len(args))
	}

	minimum, err := stdMathNumberArg("MATH.RAND.WHOLE_NUMBER", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	maximum, err := stdMathNumberArg("MATH.RAND.WHOLE_NUMBER", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	out, err := stdMathRandWholeNumber(minimum.Number, maximum.Number)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(out), nil
}

func stdMathRandNumber(lower *Number, upper *Number) (*Number, error) {
	lowerCoefficient := stdMathRandScaledCoefficient(lower, MaxDecimalPlaces)
	upperCoefficient := stdMathRandScaledCoefficient(upper, MaxDecimalPlaces)

	span := new(big.Int)
	span.Sub(upperCoefficient, lowerCoefficient)

	if span.Sign() <= 0 {
		return nil, fmt.Errorf(
			"MATH.RAND.NUMBER inclusive lower bound must be less than exclusive upper bound",
		)
	}

	offset, err := rand.Int(rand.Reader, span)
	if err != nil {
		return nil, fmt.Errorf("MATH.RAND.NUMBER could not read random data: %w", err)
	}

	result := new(big.Int)
	result.Add(lowerCoefficient, offset)

	return normaliseCoefficientAndScale(result, MaxDecimalPlaces), nil
}

func stdMathRandWholeNumber(minimum *Number, maximum *Number) (*Number, error) {
	minimumInteger, ok := stdMathRandExactInteger(minimum)
	if !ok {
		return nil, fmt.Errorf("MATH.RAND.WHOLE_NUMBER minimum must be a whole number")
	}

	maximumInteger, ok := stdMathRandExactInteger(maximum)
	if !ok {
		return nil, fmt.Errorf("MATH.RAND.WHOLE_NUMBER maximum must be a whole number")
	}

	if minimumInteger.Cmp(maximumInteger) > 0 {
		return nil, fmt.Errorf("MATH.RAND.WHOLE_NUMBER minimum cannot be greater than maximum")
	}

	span := new(big.Int)
	span.Sub(maximumInteger, minimumInteger)
	span.Add(span, big.NewInt(1))

	offset, err := rand.Int(rand.Reader, span)
	if err != nil {
		return nil, fmt.Errorf("MATH.RAND.WHOLE_NUMBER could not read random data: %w", err)
	}

	result := new(big.Int)
	result.Add(minimumInteger, offset)

	return normaliseCoefficientAndScale(result, 0), nil
}

func stdMathRandScaledCoefficient(number *Number, targetScale int) *big.Int {
	coefficient, scale := numberCoefficientAndScale(number)

	out := new(big.Int)
	out.Set(coefficient)

	switch {
	case scale < targetScale:
		out.Mul(out, powerOfTen(targetScale-scale))

	case scale > targetScale:
		out, _ = roundCoefficientToScale(out, scale, targetScale)
	}

	return out
}

func stdMathRandExactInteger(number *Number) (*big.Int, bool) {
	coefficient, scale := numberCoefficientAndScale(number)

	if scale == 0 {
		return coefficient, true
	}

	divisor := powerOfTen(scale)

	quotient := new(big.Int)
	remainder := new(big.Int)

	quotient.QuoRem(coefficient, divisor, remainder)

	if remainder.Sign() != 0 {
		return nil, false
	}

	return quotient, true
}

func stdMathRandArrayLength(array *Array, name string) (int, error) {
	if array == nil {
		return 0, fmt.Errorf("%s cannot inspect invalid array", name)
	}

	length64, accuracy := array.Len().Int64()
	if accuracy != big.Exact {
		return 0, fmt.Errorf("%s array is too large", name)
	}

	length := int(length64)
	if int64(length) != length64 {
		return 0, fmt.Errorf("%s array is too large", name)
	}

	return length, nil
}

func stdMathRandIndex(length int) (int, error) {
	if length <= 0 {
		return 0, fmt.Errorf("cannot choose random index from empty range")
	}

	index, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return 0, fmt.Errorf("could not read random data: %w", err)
	}

	return int(index.Int64()), nil
}

func stdMathRandShuffleValues(values []Value) error {
	for index := len(values) - 1; index > 0; index-- {
		swapIndex, err := stdMathRandIndex(index + 1)
		if err != nil {
			return err
		}

		values[index], values[swapIndex] = values[swapIndex], values[index]
	}

	return nil
}

func stdMathRandShuffleRunes(runes []rune) error {
	for index := len(runes) - 1; index > 0; index-- {
		swapIndex, err := stdMathRandIndex(index + 1)
		if err != nil {
			return err
		}

		runes[index], runes[swapIndex] = runes[swapIndex], runes[index]
	}

	return nil
}
