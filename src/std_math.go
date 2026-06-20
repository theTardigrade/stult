package main

import (
	"fmt"
	gomath "math"
	"math/big"
	"sync"
)

const StdMathGuardPrecisionBits uint = 64

var (
	stdMathPiOnce   sync.Once
	stdMathPiNumber *Number

	stdMathEOnce   sync.Once
	stdMathENumber *Number
)

func stdMathPiValue() Value {
	stdMathPiOnce.Do(func() {
		value := calculatePiValue(FloatPrecision)
		stdMathPiNumber = value.Number.Clone()
	})

	return NewNumberValueFromNumber(stdMathPiNumber.Clone())
}

func stdMathEValue() Value {
	stdMathEOnce.Do(func() {
		value := calculateEValue(FloatPrecision)
		stdMathENumber = value.Number.Clone()
	})

	return NewNumberValueFromNumber(stdMathENumber.Clone())
}

func NewStdMathMap() Value {
	pi := stdMathPiValue()

	entries := map[string]Binding{
		"ABS":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathAbs)),
		"CEIL":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathCeil)),
		"CLAMP":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathClamp)),
		"CUBE":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathCube)),
		"E":      NewImmutableBinding(stdMathEValue()),
		"FLOOR":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathFloor)),
		"LERP":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathLerp)),
		"MAX":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathMax)),
		"MIN":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathMin)),
		"MOD":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathMod)),
		"PI":     NewImmutableBinding(pi),
		"POWER":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathPower)),
		"RAND":   NewImmutableBinding(NewStdMathRandMap()),
		"REM":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathRem)),
		"ROUND":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathRound)),
		"SIGN":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathSign)),
		"SQRT":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathSqrt)),
		"SQUARE": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathSquare)),
		"TAU":    NewImmutableBinding(multiplyNumberByInt(pi, 2)),
		"TRIG":   NewImmutableBinding(NewStdMathTrigMap()),
		"TRUNC":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathTrunc)),
	}

	return NewMapValue(entries, true)
}

func stdMathWorkingPrecision(precision uint) uint {
	return precision + StdMathGuardPrecisionBits
}

func calculatePiValue(precision uint) Value {
	workPrecision := stdMathWorkingPrecision(precision)

	sum := newFloatWithPrecision(workPrecision)

	terms := chudnovskyTermsForPrecision(precision)

	for k := 0; k < terms; k++ {
		term := chudnovskyTerm(k, workPrecision)
		sum.Add(sum, term)
	}

	sqrt10005 := newFloatWithPrecision(workPrecision).SetInt64(10005)
	sqrt10005.Sqrt(sqrt10005)

	constant := newFloatWithPrecision(workPrecision).SetInt64(426880)
	constant.Mul(constant, sqrt10005)

	pi := newFloatWithPrecision(workPrecision)
	pi.Quo(constant, sum)

	rounded := newFloatWithPrecision(precision)
	rounded.Set(pi)

	return NewBigNumberValue(rounded)
}

func chudnovskyTermsForPrecision(precision uint) int {
	decimalDigits := int(float64(precision) * 0.30103)

	return decimalDigits/14 + 2
}

func chudnovskyTerm(k int, precision uint) *big.Float {
	sixKFactorial := factorial(6 * k)
	threeKFactorial := factorial(3 * k)
	kFactorial := factorial(k)

	multiplier := big.NewInt(int64(13591409 + 545140134*k))

	numerator := new(big.Int).Set(sixKFactorial)
	numerator.Mul(numerator, multiplier)

	denominator := new(big.Int).Set(threeKFactorial)
	denominator.Mul(denominator, kFactorial)
	denominator.Mul(denominator, kFactorial)
	denominator.Mul(denominator, kFactorial)

	power := new(big.Int).Exp(
		big.NewInt(640320),
		big.NewInt(int64(3*k)),
		nil,
	)

	denominator.Mul(denominator, power)

	term := newFloatWithPrecision(precision).SetInt(numerator)
	denominatorFloat := newFloatWithPrecision(precision).SetInt(denominator)

	term.Quo(term, denominatorFloat)

	if k%2 == 1 {
		term.Neg(term)
	}

	return term
}

func factorial(n int) *big.Int {
	result := big.NewInt(1)

	for i := 2; i <= n; i++ {
		result.Mul(result, big.NewInt(int64(i)))
	}

	return result
}

func calculateEValue(precision uint) Value {
	workPrecision := stdMathWorkingPrecision(precision)

	sum := newFloatWithPrecision(workPrecision).SetInt64(1)
	term := newFloatWithPrecision(workPrecision).SetInt64(1)

	for n := 1; n <= eTermsForPrecision(precision); n++ {
		divisor := newFloatWithPrecision(workPrecision).SetInt64(int64(n))
		term.Quo(term, divisor)
		sum.Add(sum, term)
	}

	rounded := newFloatWithPrecision(precision)
	rounded.Set(sum)

	return NewBigNumberValue(rounded)
}

func eTermsForPrecision(precision uint) int {
	decimalDigits := int(float64(precision)*0.30103) + 10
	logFactorial := 0.0

	for n := 1; ; n++ {
		logFactorial += gomath.Log10(float64(n))

		if logFactorial > float64(decimalDigits) {
			return n + 2
		}
	}
}

func multiplyNumberByInt(value Value, multiplier int64) Value {
	value = resolveSpecializedValue(value)

	return NewNumberValueFromNumber(numberMultiply(value.Number, NewSmallNumber(multiplier)))
}

func builtinStdMathSquare(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.SQUARE", args)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(numberMultiply(value.Number, value.Number)), nil
}

func builtinStdMathCube(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.CUBE", args)
	if err != nil {
		return Value{}, err
	}

	square := numberMultiply(value.Number, value.Number)

	return NewNumberValueFromNumber(numberMultiply(square, value.Number)), nil
}

func builtinStdMathAbs(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.ABS", args)
	if err != nil {
		return Value{}, err
	}

	if value.Number.Sign() < 0 {
		return NewNumberValueFromNumber(numberNegate(value.Number)), nil
	}

	return NewNumberValueFromNumber(value.Number.Clone()), nil
}

func builtinStdMathSign(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.SIGN", args)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromInt(value.Number.Sign()), nil
}

func builtinStdMathLerp(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 3 {
		return Value{}, fmt.Errorf("MATH.LERP expected 3 arguments, got %d", len(args))
	}

	start, err := stdMathNumberArg("MATH.LERP", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	end, err := stdMathNumberArg("MATH.LERP", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	amount, err := stdMathNumberArg("MATH.LERP", args[2], 3)
	if err != nil {
		return Value{}, err
	}

	workPrecision := stdMathWorkingPrecision(FloatPrecision)

	startNumber := newFloatWithPrecision(workPrecision)
	startNumber.Set(start.Number.BigFloat())

	endNumber := newFloatWithPrecision(workPrecision)
	endNumber.Set(end.Number.BigFloat())

	amountNumber := newFloatWithPrecision(workPrecision)
	amountNumber.Set(amount.Number.BigFloat())

	difference := newFloatWithPrecision(workPrecision)
	difference.Sub(endNumber, startNumber)

	scaledDifference := newFloatWithPrecision(workPrecision)
	scaledDifference.Mul(difference, amountNumber)

	out := newFloatWithPrecision(workPrecision)
	out.Add(startNumber, scaledDifference)

	rounded := newFloat()
	rounded.Set(out)

	return NewBigNumberValue(rounded), nil
}

func builtinStdMathMin(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return Value{}, fmt.Errorf("MATH.MIN expected at least 1 argument, got 0")
	}

	min, err := stdMathNumberArg("MATH.MIN", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	for index, arg := range args[1:] {
		value, err := stdMathNumberArg("MATH.MIN", arg, index+2)
		if err != nil {
			return Value{}, err
		}

		if value.Number.Cmp(min.Number) < 0 {
			min = value
		}
	}

	return NewNumberValueFromNumber(min.Number.Clone()), nil
}

func builtinStdMathMax(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return Value{}, fmt.Errorf("MATH.MAX expected at least 1 argument, got 0")
	}

	max, err := stdMathNumberArg("MATH.MAX", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	for index, arg := range args[1:] {
		value, err := stdMathNumberArg("MATH.MAX", arg, index+2)
		if err != nil {
			return Value{}, err
		}

		if value.Number.Cmp(max.Number) > 0 {
			max = value
		}
	}

	return NewNumberValueFromNumber(max.Number.Clone()), nil
}

func builtinStdMathClamp(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 3 {
		return Value{}, fmt.Errorf("MATH.CLAMP expected 3 arguments, got %d", len(args))
	}

	value, err := stdMathNumberArg("MATH.CLAMP", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	minimum, err := stdMathNumberArg("MATH.CLAMP", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	maximum, err := stdMathNumberArg("MATH.CLAMP", args[2], 3)
	if err != nil {
		return Value{}, err
	}

	if minimum.Number.Cmp(maximum.Number) > 0 {
		return Value{}, fmt.Errorf("MATH.CLAMP minimum cannot be greater than maximum")
	}

	if value.Number.Cmp(minimum.Number) < 0 {
		return NewNumberValueFromNumber(minimum.Number.Clone()), nil
	}

	if value.Number.Cmp(maximum.Number) > 0 {
		return NewNumberValueFromNumber(maximum.Number.Clone()), nil
	}

	return NewNumberValueFromNumber(value.Number.Clone()), nil
}

func builtinStdMathFloor(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.FLOOR", args)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(stdMathFloorNumber(value.Number)), nil
}

func builtinStdMathCeil(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.CEIL", args)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(stdMathCeilNumber(value.Number)), nil
}

func builtinStdMathRound(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.ROUND", args)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(stdMathRoundNumber(value.Number)), nil
}

func builtinStdMathTrunc(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.TRUNC", args)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(stdMathTruncNumber(value.Number)), nil
}

func builtinStdMathSqrt(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.SQRT", args)
	if err != nil {
		return Value{}, err
	}

	if value.Number.Sign() < 0 {
		return Value{}, fmt.Errorf("MATH.SQRT expected a non-negative number")
	}

	out := newFloat()
	out.Sqrt(value.Number.BigFloat())

	return NewBigNumberValue(out), nil
}

func builtinStdMathPower(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("MATH.POWER expected 2 arguments, got %d", len(args))
	}

	base, err := stdMathNumberArg("MATH.POWER", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	exponent, err := stdMathNumberArg("MATH.POWER", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	integerExponent, isIntegerExponent, err := stdMathExactInt64("MATH.POWER", exponent, 2)
	if err != nil {
		return Value{}, err
	}

	if isIntegerExponent {
		if base.Number.Sign() == 0 && integerExponent < 0 {
			return Value{}, fmt.Errorf("MATH.POWER cannot raise zero to a negative exponent")
		}

		out, err := stdMathPowerByInteger(base.Number, integerExponent)
		if err != nil {
			return Value{}, err
		}

		return NewNumberValueFromNumber(out), nil
	}

	if base.Number.Sign() < 0 {
		return Value{}, fmt.Errorf("MATH.POWER cannot raise a negative base to a non-integer exponent")
	}

	if base.Number.Sign() == 0 {
		if exponent.Number.Sign() < 0 {
			return Value{}, fmt.Errorf("MATH.POWER cannot raise zero to a negative exponent")
		}

		return NewNumberValueFromInt(0), nil
	}

	workPrecision := stdMathWorkingPrecision(FloatPrecision)

	lnBase, err := lnFloat(base.Number.BigFloat(), workPrecision)
	if err != nil {
		return Value{}, err
	}

	exponentAtWorkPrecision := newFloatWithPrecision(workPrecision)
	exponentAtWorkPrecision.Set(exponent.Number.BigFloat())

	powerExponent := newFloatWithPrecision(workPrecision)
	powerExponent.Mul(exponentAtWorkPrecision, lnBase)

	out, err := expFloat(powerExponent, workPrecision)
	if err != nil {
		return Value{}, err
	}

	rounded := newFloat()
	rounded.Set(out)

	return NewBigNumberValue(rounded), nil
}

func builtinStdMathMod(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("MATH.MOD expected 2 arguments, got %d", len(args))
	}

	left, err := stdMathNumberArg("MATH.MOD", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	right, err := stdMathNumberArg("MATH.MOD", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	out, err := stdMathModNumbers(left.Number, right.Number)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(out), nil
}

func builtinStdMathRem(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("MATH.REM expected 2 arguments, got %d", len(args))
	}

	left, err := stdMathNumberArg("MATH.REM", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	right, err := stdMathNumberArg("MATH.REM", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	out, err := stdMathRemNumbers(left.Number, right.Number)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(out), nil
}

func stdMathRemNumbers(left *Number, right *Number) (*Number, error) {
	if right.Sign() == 0 {
		return nil, fmt.Errorf("MATH.REM divisor cannot be zero")
	}

	leftCoefficient, leftScale := left.CoefficientAndScale()
	rightCoefficient, rightScale := right.CoefficientAndScale()

	leftCoefficient, rightCoefficient, scale := alignCoefficients(leftCoefficient, leftScale, rightCoefficient, rightScale)

	quotient := new(big.Int)
	quotient.Quo(leftCoefficient, rightCoefficient)

	multiple := new(big.Int)
	multiple.Mul(quotient, rightCoefficient)

	remainder := new(big.Int)
	remainder.Sub(leftCoefficient, multiple)

	return normaliseCoefficientAndScale(remainder, scale), nil
}

func stdMathOneNumber(name string, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("%s expected 1 argument, got %d", name, len(args))
	}

	return stdMathNumberArg(name, args[0], 1)
}

func stdMathNumberArg(name string, arg Value, position int) (Value, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueNumber {
		return Value{}, fmt.Errorf("%s argument %d expected a number", name, position)
	}

	return value, nil
}

func stdMathExactInt64(name string, value Value, position int) (int64, bool, error) {
	integer, accuracy := value.Number.Int(nil)
	if accuracy != big.Exact {
		return 0, false, nil
	}

	if !integer.IsInt64() {
		return 0, true, fmt.Errorf("%s argument %d is too large", name, position)
	}

	return integer.Int64(), true, nil
}

func stdMathFloorNumber(number *Number) *Number {
	coefficient, scale := number.CoefficientAndScale()
	if scale == 0 {
		return normaliseCoefficientAndScale(coefficient, 0)
	}

	quotient := new(big.Int)
	remainder := new(big.Int)

	quotient.QuoRem(coefficient, powerOfTen(scale), remainder)

	if coefficient.Sign() < 0 && remainder.Sign() != 0 {
		quotient.Sub(quotient, big.NewInt(1))
	}

	return normaliseCoefficientAndScale(quotient, 0)
}

func stdMathCeilNumber(number *Number) *Number {
	coefficient, scale := number.CoefficientAndScale()
	if scale == 0 {
		return normaliseCoefficientAndScale(coefficient, 0)
	}

	quotient := new(big.Int)
	remainder := new(big.Int)

	quotient.QuoRem(coefficient, powerOfTen(scale), remainder)

	if coefficient.Sign() > 0 && remainder.Sign() != 0 {
		quotient.Add(quotient, big.NewInt(1))
	}

	return normaliseCoefficientAndScale(quotient, 0)
}

func stdMathTruncNumber(number *Number) *Number {
	coefficient, scale := number.CoefficientAndScale()
	if scale == 0 {
		return normaliseCoefficientAndScale(coefficient, 0)
	}

	quotient := new(big.Int)
	quotient.Quo(coefficient, powerOfTen(scale))

	return normaliseCoefficientAndScale(quotient, 0)
}

func stdMathRoundNumber(number *Number) *Number {
	coefficient, scale := number.CoefficientAndScale()
	if scale == 0 {
		return normaliseCoefficientAndScale(coefficient, 0)
	}

	rounded := roundedQuotient(coefficient, powerOfTen(scale))

	return normaliseCoefficientAndScale(rounded, 0)
}

func stdMathModNumbers(left *Number, right *Number) (*Number, error) {
	if right.Sign() == 0 {
		return nil, fmt.Errorf("MATH.MOD divisor cannot be zero")
	}

	leftCoefficient, leftScale := left.CoefficientAndScale()
	rightCoefficient, rightScale := right.CoefficientAndScale()

	leftCoefficient, rightCoefficient, scale := alignCoefficients(leftCoefficient, leftScale, rightCoefficient, rightScale)

	quotient := stdMathFloorQuotient(leftCoefficient, rightCoefficient)

	multiple := new(big.Int)
	multiple.Mul(quotient, rightCoefficient)

	remainder := new(big.Int)
	remainder.Sub(leftCoefficient, multiple)

	return normaliseCoefficientAndScale(remainder, scale), nil
}

func stdMathFloorQuotient(numerator *big.Int, denominator *big.Int) *big.Int {
	quotient := new(big.Int)
	remainder := new(big.Int)

	quotient.QuoRem(numerator, denominator, remainder)

	if remainder.Sign() != 0 && remainder.Sign() != denominator.Sign() {
		quotient.Sub(quotient, big.NewInt(1))
	}

	return quotient
}

func stdMathPowerByInteger(base *Number, exponent int64) (*Number, error) {
	if exponent == 0 {
		return NewSmallNumber(1), nil
	}

	negativeExponent := exponent < 0
	magnitude := exponentMagnitude(exponent)

	result := NewSmallNumber(1)
	factor := base.Clone()

	for magnitude > 0 {
		if magnitude%2 == 1 {
			result = numberMultiply(result, factor)
		}

		magnitude /= 2

		if magnitude > 0 {
			factor = numberMultiply(factor, factor)
		}
	}

	if !negativeExponent {
		return result, nil
	}

	return numberDivide(NewSmallNumber(1), result)
}

func exponentMagnitude(exponent int64) uint64 {
	if exponent >= 0 {
		return uint64(exponent)
	}

	return uint64(-(exponent + 1)) + 1
}

func lnFloat(value *big.Float, precision uint) (*big.Float, error) {
	if value.Sign() <= 0 {
		return nil, fmt.Errorf("cannot calculate logarithm of non-positive number")
	}

	workPrecision := stdMathWorkingPrecision(precision)

	x := newFloatWithPrecision(workPrecision)
	x.Set(value)

	mantissa := newFloatWithPrecision(workPrecision)
	exponent := x.MantExp(mantissa)

	two := newFloatWithPrecision(workPrecision).SetInt64(2)
	mantissa.Mul(mantissa, two)
	powerOfTwo := exponent - 1

	sqrtTwo := newFloatWithPrecision(workPrecision).SetInt64(2)
	sqrtTwo.Sqrt(sqrtTwo)

	if mantissa.Cmp(sqrtTwo) > 0 {
		mantissa.Quo(mantissa, two)
		powerOfTwo++
	}

	lnMantissa := lnNearOne(mantissa, workPrecision)
	lnTwo := lnTwoFloat(workPrecision)

	scaledLnTwo := newFloatWithPrecision(workPrecision)
	scaledLnTwo.Mul(lnTwo, newFloatWithPrecision(workPrecision).SetInt64(int64(powerOfTwo)))

	out := newFloatWithPrecision(workPrecision)
	out.Add(lnMantissa, scaledLnTwo)

	rounded := newFloatWithPrecision(precision)
	rounded.Set(out)

	return rounded, nil
}

func lnTwoFloat(precision uint) *big.Float {
	two := newFloatWithPrecision(precision).SetInt64(2)
	return lnNearOne(two, precision)
}

func lnNearOne(value *big.Float, precision uint) *big.Float {
	one := newFloatWithPrecision(precision).SetInt64(1)

	numerator := newFloatWithPrecision(precision)
	numerator.Sub(value, one)

	denominator := newFloatWithPrecision(precision)
	denominator.Add(value, one)

	z := newFloatWithPrecision(precision)
	z.Quo(numerator, denominator)

	zSquared := newFloatWithPrecision(precision)
	zSquared.Mul(z, z)

	sum := newFloatWithPrecision(precision)
	sum.Set(z)

	term := newFloatWithPrecision(precision)
	term.Set(z)

	epsilon := stdMathEpsilon(precision)

	for divisor := int64(3); ; divisor += 2 {
		term.Mul(term, zSquared)

		addend := newFloatWithPrecision(precision)
		addend.Quo(term, newFloatWithPrecision(precision).SetInt64(divisor))

		sum.Add(sum, addend)

		if absFloat(addend).Cmp(epsilon) < 0 {
			break
		}
	}

	sum.Mul(sum, newFloatWithPrecision(precision).SetInt64(2))

	return sum
}

func expFloat(value *big.Float, precision uint) (*big.Float, error) {
	workPrecision := stdMathWorkingPrecision(precision)

	x := newFloatWithPrecision(workPrecision)
	x.Set(value)

	lnTwo := lnTwoFloat(workPrecision)

	ratio := newFloatWithPrecision(workPrecision)
	ratio.Quo(x, lnTwo)

	powerOfTwoBigInt := nearestInteger(ratio)
	if !powerOfTwoBigInt.IsInt64() {
		return nil, fmt.Errorf("exponent is too large")
	}

	powerOfTwoInt64 := powerOfTwoBigInt.Int64()
	if int64(int(powerOfTwoInt64)) != powerOfTwoInt64 {
		return nil, fmt.Errorf("exponent is too large")
	}

	powerOfTwo := int(powerOfTwoInt64)

	reduced := newFloatWithPrecision(workPrecision)
	reduced.Sub(
		x,
		newFloatWithPrecision(workPrecision).Mul(
			newFloatWithPrecision(workPrecision).SetInt(powerOfTwoBigInt),
			lnTwo,
		),
	)

	expReduced := expNearZero(reduced, workPrecision)

	out := newFloatWithPrecision(workPrecision)
	out.SetMantExp(expReduced, powerOfTwo)

	rounded := newFloatWithPrecision(precision)
	rounded.Set(out)

	return rounded, nil
}

func expNearZero(value *big.Float, precision uint) *big.Float {
	one := newFloatWithPrecision(precision).SetInt64(1)

	sum := newFloatWithPrecision(precision)
	sum.Set(one)

	term := newFloatWithPrecision(precision)
	term.Set(one)

	epsilon := stdMathEpsilon(precision)

	for n := int64(1); ; n++ {
		term.Mul(term, value)
		term.Quo(term, newFloatWithPrecision(precision).SetInt64(n))

		sum.Add(sum, term)

		if absFloat(term).Cmp(epsilon) < 0 {
			break
		}
	}

	return sum
}

func nearestInteger(value *big.Float) *big.Int {
	adjusted := newFloatWithPrecision(value.Prec())
	adjusted.Set(value)

	half := newFloatWithPrecision(value.Prec()).SetFloat64(0.5)

	if adjusted.Sign() >= 0 {
		adjusted.Add(adjusted, half)
	} else {
		adjusted.Sub(adjusted, half)
	}

	integer, _ := adjusted.Int(nil)

	return integer
}

func stdMathEpsilon(precision uint) *big.Float {
	one := newFloatWithPrecision(precision).SetInt64(1)

	epsilon := newFloatWithPrecision(precision)
	epsilon.SetMantExp(one, -int(precision))

	return epsilon
}

func absFloat(value *big.Float) *big.Float {
	out := newFloatWithPrecision(value.Prec())
	out.Set(value)

	if out.Sign() < 0 {
		out.Neg(out)
	}

	return out
}
