package main

import (
	"fmt"
	gomath "math"
	"math/big"
)

const StdMathGuardPrecisionBits uint = 64

func NewStdMathMap() Value {
	pi := calculatePiValue(FloatPrecision)

	entries := map[string]Binding{
		"ABS":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathAbs)),
		"CEIL":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathCeil)),
		"CLAMP":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathClamp)),
		"CUBE":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathCube)),
		"E":      NewImmutableBinding(calculateEValue(FloatPrecision)),
		"FLOOR":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathFloor)),
		"LERP":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathLerp)),
		"MAX":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathMax)),
		"MIN":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathMin)),
		"MOD":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathMod)),
		"PI":     NewImmutableBinding(pi),
		"POWER":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathPower)),
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

	return Value{
		Kind:   ValueNumber,
		Number: rounded,
	}
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

	return Value{
		Kind:   ValueNumber,
		Number: rounded,
	}
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

	out := newFloat()
	out.Mul(value.Number, newFloat().SetInt64(multiplier))

	return Value{Kind: ValueNumber, Number: out}
}

func builtinStdMathSquare(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.SQUARE", args)
	if err != nil {
		return Value{}, err
	}

	out := newFloat()
	out.Mul(value.Number, value.Number)

	return Value{Kind: ValueNumber, Number: out}, nil
}

func builtinStdMathCube(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.CUBE", args)
	if err != nil {
		return Value{}, err
	}

	out := newFloat()
	out.Mul(value.Number, value.Number)
	out.Mul(out, value.Number)

	return Value{Kind: ValueNumber, Number: out}, nil
}

func builtinStdMathAbs(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.ABS", args)
	if err != nil {
		return Value{}, err
	}

	out := CloneNumber(value.Number)

	if out.Sign() < 0 {
		out.Neg(out)
	}

	return Value{Kind: ValueNumber, Number: out}, nil
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
	startNumber.Set(start.Number)

	endNumber := newFloatWithPrecision(workPrecision)
	endNumber.Set(end.Number)

	amountNumber := newFloatWithPrecision(workPrecision)
	amountNumber.Set(amount.Number)

	difference := newFloatWithPrecision(workPrecision)
	difference.Sub(endNumber, startNumber)

	scaledDifference := newFloatWithPrecision(workPrecision)
	scaledDifference.Mul(difference, amountNumber)

	out := newFloatWithPrecision(workPrecision)
	out.Add(startNumber, scaledDifference)

	rounded := newFloat()
	rounded.Set(out)

	return Value{Kind: ValueNumber, Number: rounded}, nil
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

	return Value{Kind: ValueNumber, Number: CloneNumber(min.Number)}, nil
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

	return Value{Kind: ValueNumber, Number: CloneNumber(max.Number)}, nil
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
		return Value{Kind: ValueNumber, Number: CloneNumber(minimum.Number)}, nil
	}

	if value.Number.Cmp(maximum.Number) > 0 {
		return Value{Kind: ValueNumber, Number: CloneNumber(maximum.Number)}, nil
	}

	return Value{Kind: ValueNumber, Number: CloneNumber(value.Number)}, nil
}

func builtinStdMathFloor(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.FLOOR", args)
	if err != nil {
		return Value{}, err
	}

	return Value{Kind: ValueNumber, Number: floorFloat(value.Number)}, nil
}

func builtinStdMathCeil(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.CEIL", args)
	if err != nil {
		return Value{}, err
	}

	return Value{Kind: ValueNumber, Number: ceilFloat(value.Number)}, nil
}

func builtinStdMathRound(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.ROUND", args)
	if err != nil {
		return Value{}, err
	}

	adjusted := CloneNumber(value.Number)
	half := newFloat().SetFloat64(0.5)

	if adjusted.Sign() >= 0 {
		adjusted.Add(adjusted, half)
		return Value{Kind: ValueNumber, Number: floorFloat(adjusted)}, nil
	}

	adjusted.Sub(adjusted, half)
	return Value{Kind: ValueNumber, Number: ceilFloat(adjusted)}, nil
}

func builtinStdMathTrunc(_ *RuntimeContext, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.TRUNC", args)
	if err != nil {
		return Value{}, err
	}

	return Value{Kind: ValueNumber, Number: truncFloat(value.Number)}, nil
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
	out.Sqrt(value.Number)

	return Value{Kind: ValueNumber, Number: out}, nil
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

		out := powerFloatByInteger(base.Number, integerExponent)
		return Value{Kind: ValueNumber, Number: out}, nil
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

	lnBase, err := lnFloat(base.Number, workPrecision)
	if err != nil {
		return Value{}, err
	}

	exponentAtWorkPrecision := newFloatWithPrecision(workPrecision)
	exponentAtWorkPrecision.Set(exponent.Number)

	powerExponent := newFloatWithPrecision(workPrecision)
	powerExponent.Mul(exponentAtWorkPrecision, lnBase)

	out, err := expFloat(powerExponent, workPrecision)
	if err != nil {
		return Value{}, err
	}

	rounded := newFloat()
	rounded.Set(out)

	return Value{Kind: ValueNumber, Number: rounded}, nil
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

	if right.Number.Sign() == 0 {
		return Value{}, fmt.Errorf("MATH.MOD divisor cannot be zero")
	}

	workPrecision := stdMathWorkingPrecision(FloatPrecision)

	leftAtWorkPrecision := newFloatWithPrecision(workPrecision)
	leftAtWorkPrecision.Set(left.Number)

	rightAtWorkPrecision := newFloatWithPrecision(workPrecision)
	rightAtWorkPrecision.Set(right.Number)

	quotient := newFloatWithPrecision(workPrecision)
	quotient.Quo(leftAtWorkPrecision, rightAtWorkPrecision)

	flooredQuotient := floorFloatWithPrecision(quotient, workPrecision)

	multiple := newFloatWithPrecision(workPrecision)
	multiple.Mul(flooredQuotient, rightAtWorkPrecision)

	out := newFloatWithPrecision(workPrecision)
	out.Sub(leftAtWorkPrecision, multiple)

	rounded := newFloat()
	rounded.Set(out)

	return Value{Kind: ValueNumber, Number: rounded}, nil
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

func truncFloat(value *big.Float) *big.Float {
	return truncFloatWithPrecision(value, FloatPrecision)
}

func floorFloat(value *big.Float) *big.Float {
	return floorFloatWithPrecision(value, FloatPrecision)
}

func ceilFloat(value *big.Float) *big.Float {
	return ceilFloatWithPrecision(value, FloatPrecision)
}

func truncFloatWithPrecision(value *big.Float, precision uint) *big.Float {
	integer, _ := value.Int(nil)
	return newFloatWithPrecision(precision).SetInt(integer)
}

func floorFloatWithPrecision(value *big.Float, precision uint) *big.Float {
	integer, _ := value.Int(nil)
	truncated := newFloatWithPrecision(precision).SetInt(integer)

	if truncated.Cmp(value) > 0 {
		integer.Sub(integer, big.NewInt(1))
	}

	return newFloatWithPrecision(precision).SetInt(integer)
}

func ceilFloatWithPrecision(value *big.Float, precision uint) *big.Float {
	integer, _ := value.Int(nil)
	truncated := newFloatWithPrecision(precision).SetInt(integer)

	if truncated.Cmp(value) < 0 {
		integer.Add(integer, big.NewInt(1))
	}

	return newFloatWithPrecision(precision).SetInt(integer)
}

func powerFloatByInteger(base *big.Float, exponent int64) *big.Float {
	if exponent == 0 {
		return newFloat().SetInt64(1)
	}

	negativeExponent := exponent < 0
	magnitude := exponentMagnitude(exponent)

	result := newFloat().SetInt64(1)
	factor := CloneNumber(base)

	for magnitude > 0 {
		if magnitude%2 == 1 {
			result.Mul(result, factor)
		}

		magnitude /= 2

		if magnitude > 0 {
			factor.Mul(factor, factor)
		}
	}

	if negativeExponent {
		one := newFloat().SetInt64(1)
		result.Quo(one, result)
	}

	return result
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
