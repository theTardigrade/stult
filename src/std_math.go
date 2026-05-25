package main

import "math/big"

func NewStdMathMap() Value {
	entries := map[string]Binding{
		"PI": {
			Value:       calculatePiValue(FloatPrecision),
			IsImmutable: true,
		},
	}

	order := []string{"PI"}

	return NewMapValue(entries, order, true)
}

func calculatePiValue(precision uint) Value {
	// Use some extra internal precision to reduce rounding error during calculation.
	workPrecision := precision + 64

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
	// log10(2) is about 0.30103, so this converts bits to decimal digits.
	// The Chudnovsky algorithm gives about 14 decimal digits per term.
	decimalDigits := int(float64(precision) * 0.30103)

	return decimalDigits/14 + 2
}

func chudnovskyTerm(k int, precision uint) *big.Float {
	// term =
	//   (-1)^k * (6k)! * (13591409 + 545140134k)
	//   ------------------------------------------------
	//   (3k)! * (k!)^3 * 640320^(3k)

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
