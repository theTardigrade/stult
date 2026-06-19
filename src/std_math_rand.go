package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func NewStdMathRandMap() Value {
	entries := map[string]Binding{
		"CHOICE":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathRandChoice)),
		"INTEGER": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathRandInteger)),
		"NUMBER":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathRandNumber)),
		"SHUFFLE": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathRandShuffle)),
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

func builtinStdMathRandInteger(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("MATH.RAND.INTEGER expected 2 arguments, got %d", len(args))
	}

	minimum, err := stdMathNumberArg("MATH.RAND.INTEGER", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	maximum, err := stdMathNumberArg("MATH.RAND.INTEGER", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	out, err := stdMathRandInteger(minimum.Number, maximum.Number)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(out), nil
}

func builtinStdMathRandChoice(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("MATH.RAND.CHOICE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("MATH.RAND.CHOICE cannot choose from invalid array")
		}

		if value.Array.Len().Sign() == 0 {
			return Value{}, fmt.Errorf("MATH.RAND.CHOICE cannot choose from empty array")
		}

		length, err := stdMathRandArrayLength(value.Array, "MATH.RAND.CHOICE")
		if err != nil {
			return Value{}, err
		}

		index, err := stdMathRandIndex(length)
		if err != nil {
			return Value{}, err
		}

		chosen, ok, err := value.Array.Get(NewSmallNumber(int64(index)))
		if err != nil {
			return Value{}, err
		}

		if !ok {
			return Value{}, fmt.Errorf("MATH.RAND.CHOICE generated array index out of bounds")
		}

		return chosen, nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("MATH.RAND.CHOICE cannot choose from invalid string")
		}

		if len(value.Text.Runes) == 0 {
			return Value{}, fmt.Errorf("MATH.RAND.CHOICE cannot choose from empty string")
		}

		index, err := stdMathRandIndex(len(value.Text.Runes))
		if err != nil {
			return Value{}, err
		}

		return NewStringValue(string(value.Text.Runes[index])), nil

	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("MATH.RAND.CHOICE cannot choose from invalid map")
		}

		if len(value.Map.Entries) == 0 {
			return Value{}, fmt.Errorf("MATH.RAND.CHOICE cannot choose from empty map")
		}

		keys := sortedMapKeys(value.Map)

		index, err := stdMathRandIndex(len(keys))
		if err != nil {
			return Value{}, err
		}

		return value.Map.Entries[keys[index]].Value, nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return Value{}, fmt.Errorf("MATH.RAND.CHOICE expected an array, string or map")

	default:
		return Value{}, fmt.Errorf("MATH.RAND.CHOICE cannot choose from unknown value kind")
	}
}

func builtinStdMathRandShuffle(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("MATH.RAND.SHUFFLE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("MATH.RAND.SHUFFLE cannot shuffle invalid array")
		}

		if _, err := stdMathRandArrayLength(value.Array, "MATH.RAND.SHUFFLE"); err != nil {
			return Value{}, err
		}

		elements := make([]Value, 0, len(value.Array.Elements))
		if err := value.Array.ForEach(func(_ *Number, element Value) error {
			elements = append(elements, element)
			return nil
		}); err != nil {
			return Value{}, err
		}

		if err := stdMathRandShuffleValues(elements); err != nil {
			return Value{}, err
		}

		return NewArrayValue(elements, false), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("MATH.RAND.SHUFFLE cannot shuffle invalid string")
		}

		runes := append([]rune{}, value.Text.Runes...)

		if err := stdMathRandShuffleRunes(runes); err != nil {
			return Value{}, err
		}

		return NewStringValue(string(runes)), nil

	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("MATH.RAND.SHUFFLE cannot shuffle invalid map")
		}

		keys := sortedMapKeys(value.Map)
		values := make([]Value, 0, len(keys))

		for _, key := range keys {
			values = append(values, value.Map.Entries[key].Value)
		}

		if err := stdMathRandShuffleValues(values); err != nil {
			return Value{}, err
		}

		entries := make(map[string]Binding, len(keys))

		for index, key := range keys {
			originalBinding := value.Map.Entries[key]

			entries[key] = Binding{
				Value:       values[index],
				IsImmutable: originalBinding.IsImmutable,
			}
		}

		return NewMapValue(entries, false), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return Value{}, fmt.Errorf("MATH.RAND.SHUFFLE expected an array or string")

	default:
		return Value{}, fmt.Errorf("MATH.RAND.SHUFFLE cannot shuffle unknown value kind")
	}
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

func stdMathRandInteger(minimum *Number, maximum *Number) (*Number, error) {
	minimumInteger, ok := stdMathRandExactInteger(minimum)
	if !ok {
		return nil, fmt.Errorf("MATH.RAND.INTEGER minimum must be an integer")
	}

	maximumInteger, ok := stdMathRandExactInteger(maximum)
	if !ok {
		return nil, fmt.Errorf("MATH.RAND.INTEGER maximum must be an integer")
	}

	if minimumInteger.Cmp(maximumInteger) > 0 {
		return nil, fmt.Errorf("MATH.RAND.INTEGER minimum cannot be greater than maximum")
	}

	span := new(big.Int)
	span.Sub(maximumInteger, minimumInteger)
	span.Add(span, big.NewInt(1))

	offset, err := rand.Int(rand.Reader, span)
	if err != nil {
		return nil, fmt.Errorf("MATH.RAND.INTEGER could not read random data: %w", err)
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
