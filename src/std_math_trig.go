package main

import (
	"fmt"
	"math/big"
)

const StdMathTrigGuardPrecisionBits uint = StdMathGuardPrecisionBits * 2

func NewStdMathTrigMap() Value {
	entries := map[string]Binding{
		"COS":     NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathCos)),
		"DEGREES": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathDegrees)),
		"RADIANS": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathRadians)),
		"SIN":     NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathSin)),
		"TAN":     NewImmutableBinding(NewBuiltinFunctionValue(builtinStdMathTan)),
	}

	return NewMapValue(entries, true)
}

func builtinStdMathSin(_ *Interpreter, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.TRIG.SIN", args)
	if err != nil {
		return Value{}, err
	}

	workPrecision := stdMathTrigWorkingPrecision(value.Number)

	angle, _, err := reduceTrigAngle(value.Number, workPrecision)
	if err != nil {
		return Value{}, err
	}

	out := sinNearZero(angle, workPrecision)

	return stdMathNumberValue(out), nil
}

func builtinStdMathCos(_ *Interpreter, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.TRIG.COS", args)
	if err != nil {
		return Value{}, err
	}

	workPrecision := stdMathTrigWorkingPrecision(value.Number)

	angle, cosSign, err := reduceTrigAngle(value.Number, workPrecision)
	if err != nil {
		return Value{}, err
	}

	out := cosNearZero(angle, workPrecision)

	if cosSign < 0 {
		out.Neg(out)
	}

	return stdMathNumberValue(out), nil
}

func builtinStdMathTan(_ *Interpreter, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.TRIG.TAN", args)
	if err != nil {
		return Value{}, err
	}

	workPrecision := stdMathTrigWorkingPrecision(value.Number)

	angle, cosSign, err := reduceTrigAngle(value.Number, workPrecision)
	if err != nil {
		return Value{}, err
	}

	sine := sinNearZero(angle, workPrecision)
	cosine := cosNearZero(angle, workPrecision)

	if cosSign < 0 {
		cosine.Neg(cosine)
	}

	if absFloat(cosine).Cmp(stdMathEpsilon(FloatPrecision)) <= 0 {
		return Value{}, fmt.Errorf("MATH.TRIG.TAN is undefined for this angle")
	}

	out := newFloatWithPrecision(workPrecision)
	out.Quo(sine, cosine)

	return stdMathNumberValue(out), nil
}

func builtinStdMathRadians(_ *Interpreter, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.TRIG.RADIANS", args)
	if err != nil {
		return Value{}, err
	}

	workPrecision := stdMathTrigWorkingPrecision(value.Number)

	degrees := newFloatWithPrecision(workPrecision)
	degrees.Set(value.Number)

	pi := calculatePiValue(workPrecision).Number

	out := newFloatWithPrecision(workPrecision)
	out.Mul(degrees, pi)
	out.Quo(out, newFloatWithPrecision(workPrecision).SetInt64(180))

	return stdMathNumberValue(out), nil
}

func builtinStdMathDegrees(_ *Interpreter, args []Value) (Value, error) {
	value, err := stdMathOneNumber("MATH.TRIG.DEGREES", args)
	if err != nil {
		return Value{}, err
	}

	workPrecision := stdMathTrigWorkingPrecision(value.Number)

	radians := newFloatWithPrecision(workPrecision)
	radians.Set(value.Number)

	pi := calculatePiValue(workPrecision).Number

	out := newFloatWithPrecision(workPrecision)
	out.Mul(radians, newFloatWithPrecision(workPrecision).SetInt64(180))
	out.Quo(out, pi)

	return stdMathNumberValue(out), nil
}

func stdMathTrigWorkingPrecision(values ...*big.Float) uint {
	precision := FloatPrecision + StdMathTrigGuardPrecisionBits

	for _, value := range values {
		if value == nil || value.Sign() == 0 {
			continue
		}

		mantissa := new(big.Float)
		exponent := value.MantExp(mantissa)

		if exponent > 0 {
			precision += uint(exponent)
		}
	}

	return precision
}

func reduceTrigAngle(value *big.Float, precision uint) (*big.Float, int, error) {
	if value == nil {
		return nil, 1, fmt.Errorf("cannot reduce invalid angle")
	}

	x := newFloatWithPrecision(precision)
	x.Set(value)

	pi := calculatePiValue(precision).Number

	tau := newFloatWithPrecision(precision)
	tau.Mul(pi, newFloatWithPrecision(precision).SetInt64(2))

	ratio := newFloatWithPrecision(precision)
	ratio.Quo(x, tau)

	quotient := nearestInteger(ratio)

	reduced := newFloatWithPrecision(precision)
	reduced.Sub(
		x,
		newFloatWithPrecision(precision).Mul(
			newFloatWithPrecision(precision).SetInt(quotient),
			tau,
		),
	)

	halfPi := newFloatWithPrecision(precision)
	halfPi.Quo(pi, newFloatWithPrecision(precision).SetInt64(2))

	negativeHalfPi := newFloatWithPrecision(precision)
	negativeHalfPi.Neg(halfPi)

	cosSign := 1

	if reduced.Cmp(halfPi) > 0 {
		// sin(x) = sin(pi - x)
		// cos(x) = -cos(pi - x)
		next := newFloatWithPrecision(precision)
		next.Sub(pi, reduced)

		reduced = next
		cosSign = -1
	} else if reduced.Cmp(negativeHalfPi) < 0 {
		// sin(x) = sin(-pi - x)
		// cos(x) = -cos(-pi - x)
		negativePi := newFloatWithPrecision(precision)
		negativePi.Neg(pi)

		next := newFloatWithPrecision(precision)
		next.Sub(negativePi, reduced)

		reduced = next
		cosSign = -1
	}

	return reduced, cosSign, nil
}

func sinNearZero(value *big.Float, precision uint) *big.Float {
	x := newFloatWithPrecision(precision)
	x.Set(value)

	xSquared := newFloatWithPrecision(precision)
	xSquared.Mul(x, x)

	sum := newFloatWithPrecision(precision)
	sum.Set(x)

	term := newFloatWithPrecision(precision)
	term.Set(x)

	epsilon := stdMathEpsilon(precision)

	for n := int64(1); ; n++ {
		denominator := newFloatWithPrecision(precision)
		denominator.SetInt64((2 * n) * (2*n + 1))

		term.Mul(term, xSquared)
		term.Quo(term, denominator)
		term.Neg(term)

		sum.Add(sum, term)

		if absFloat(term).Cmp(epsilon) < 0 {
			break
		}
	}

	return sum
}

func cosNearZero(value *big.Float, precision uint) *big.Float {
	x := newFloatWithPrecision(precision)
	x.Set(value)

	xSquared := newFloatWithPrecision(precision)
	xSquared.Mul(x, x)

	sum := newFloatWithPrecision(precision)
	sum.SetInt64(1)

	term := newFloatWithPrecision(precision)
	term.SetInt64(1)

	epsilon := stdMathEpsilon(precision)

	for n := int64(1); ; n++ {
		denominator := newFloatWithPrecision(precision)
		denominator.SetInt64((2*n - 1) * (2 * n))

		term.Mul(term, xSquared)
		term.Quo(term, denominator)
		term.Neg(term)

		sum.Add(sum, term)

		if absFloat(term).Cmp(epsilon) < 0 {
			break
		}
	}

	return sum
}

func stdMathNumberValue(value *big.Float) Value {
	rounded := newFloat()
	rounded.Set(value)

	return Value{
		Kind:   ValueNumber,
		Number: rounded,
	}
}
