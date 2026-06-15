package main

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

const FloatPrecision uint = 1024
const MaxDecimalScale = 256
const DefaultFractionDigits = 32

type NumberKind int

const (
	NumberSmallInt NumberKind = iota
	NumberBigInt
	NumberDecimal
)

// Number stores one user-visible Stult number.
//
// For NumberSmallInt, smallInt stores the whole signed integer value.
//
// For NumberBigInt, bigInt stores the whole signed integer value.
//
// For NumberDecimal, bigInt stores the signed scaled decimal coefficient and
// scale stores the number of decimal places:
//
//	value = bigInt / 10^scale
//
// Decimal values are normalised so that:
//
//	0 < scale <= MaxDecimalScale
//
// Integer decimal values are collapsed to NumberSmallInt or NumberBigInt.
// Decimal operations that would produce more than MaxDecimalScale decimal
// places are rounded to MaxDecimalScale.
type Number struct {
	kind     NumberKind
	smallInt int64
	bigInt   *big.Int
	scale    int
}

var (
	bigTen             = big.NewInt(10)
	bigTwo             = big.NewInt(2)
	decimalPowersOfTen = buildDecimalPowersOfTen(MaxDecimalScale * 2)
)

func NewNumberValueFromString(literal string) (Value, error) {
	number, err := NewNumberFromString(literal)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromNumber(number), nil
}

func NewNumberFromString(literal string) (*Number, error) {
	literal = strings.TrimSpace(literal)
	if literal == "" {
		return nil, fmt.Errorf("invalid number %q", literal)
	}

	if isIntegerNumberLiteral(literal) {
		if small, err := strconv.ParseInt(literal, 10, 64); err == nil {
			return NewSmallNumber(small), nil
		}

		integer, ok := new(big.Int).SetString(literal, 10)
		if !ok {
			return nil, fmt.Errorf("invalid number %q", literal)
		}

		return NewBigIntNumber(integer), nil
	}

	return NewDecimalNumberFromLiteral(literal)
}

func NewDecimalNumberFromLiteral(literal string) (*Number, error) {
	mantissa := literal
	exponent := 0

	if index := strings.IndexAny(literal, "eE"); index >= 0 {
		mantissa = literal[:index]

		parsedExponent, err := strconv.Atoi(literal[index+1:])
		if err != nil {
			return nil, fmt.Errorf("invalid number %q", literal)
		}

		exponent = parsedExponent
	}

	negative := false

	if strings.HasPrefix(mantissa, "-") {
		negative = true
		mantissa = mantissa[1:]
	} else if strings.HasPrefix(mantissa, "+") {
		mantissa = mantissa[1:]
	}

	if mantissa == "" {
		return nil, fmt.Errorf("invalid number %q", literal)
	}

	scale := 0
	digits := mantissa

	if index := strings.Index(mantissa, "."); index >= 0 {
		whole := mantissa[:index]
		fraction := mantissa[index+1:]

		scale = len(fraction)
		digits = whole + fraction
	}

	if digits == "" {
		return nil, fmt.Errorf("invalid number %q", literal)
	}

	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return nil, fmt.Errorf("invalid number %q", literal)
		}
	}

	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		return NewSmallNumber(0), nil
	}

	coefficient, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, fmt.Errorf("invalid number %q", literal)
	}

	if negative {
		coefficient.Neg(coefficient)
	}

	scale -= exponent

	if scale < 0 {
		coefficient.Mul(coefficient, powerOfTen(-scale))
		scale = 0
	}

	return normaliseCoefficientAndScale(coefficient, scale), nil
}

func NewNumberValueFromInt(value int) Value {
	return NewNumberValueFromInt64(int64(value))
}

func NewNumberValueFromInt64(value int64) Value {
	return NewNumberValueFromNumber(NewSmallNumber(value))
}

func NewNumberValueFromBigInt(value *big.Int) Value {
	return NewNumberValueFromNumber(NewBigIntNumber(value))
}

func NewNumberValueFromNumber(number *Number) Value {
	if number == nil {
		number = NewSmallNumber(0)
	}

	return Value{
		Kind:   ValueNumber,
		Number: number,
	}
}

func NewBigNumberValue(value *big.Float) Value {
	return NewNumberValueFromNumber(NewBigNumber(value))
}

func NewSmallNumber(value int64) *Number {
	return &Number{
		kind:     NumberSmallInt,
		smallInt: value,
	}
}

func NewBigIntNumber(value *big.Int) *Number {
	if value == nil || value.Sign() == 0 {
		return NewSmallNumber(0)
	}

	if value.IsInt64() {
		return NewSmallNumber(value.Int64())
	}

	return &Number{
		kind:   NumberBigInt,
		bigInt: new(big.Int).Set(value),
	}
}

func NewBigNumber(value *big.Float) *Number {
	return newDecimalNumberFromBigFloat(value, MaxDecimalScale)
}

func CloneNumber(number *Number) *Number {
	if number == nil {
		return NewSmallNumber(0)
	}

	switch number.kind {
	case NumberSmallInt:
		return NewSmallNumber(number.smallInt)

	case NumberBigInt:
		return NewBigIntNumber(number.bigInt)

	case NumberDecimal:
		return &Number{
			kind:   NumberDecimal,
			bigInt: new(big.Int).Set(numberBigInt(number)),
			scale:  number.scale,
		}

	default:
		return NewSmallNumber(0)
	}
}

func newFloat() *big.Float {
	return newFloatWithPrecision(FloatPrecision)
}

func newFloatWithPrecision(precision uint) *big.Float {
	return new(big.Float).SetPrec(precision).SetMode(big.ToNearestEven)
}

func isIntegerNumberLiteral(literal string) bool {
	return !strings.ContainsAny(literal, ".eE")
}

func (number *Number) IsBig() bool {
	return number != nil && number.kind != NumberSmallInt
}

func (number *Number) IsSmall() bool {
	return number == nil || number.kind == NumberSmallInt
}

func (number *Number) BigFloat() *big.Float {
	return numberToBigFloat(number)
}

func (number *Number) Sign() int {
	return numberSign(number)
}

func (number *Number) Cmp(other *Number) int {
	return numberCompare(number, other)
}

func (number *Number) Int64() (int64, big.Accuracy) {
	integer, accuracy := number.Int(nil)
	if !integer.IsInt64() {
		if integer.Sign() < 0 {
			return math.MinInt64, big.Below
		}

		return math.MaxInt64, big.Above
	}

	return integer.Int64(), accuracy
}

func (number *Number) Int(out *big.Int) (*big.Int, big.Accuracy) {
	if out == nil {
		out = new(big.Int)
	}

	switch {
	case number == nil:
		return out.SetInt64(0), big.Exact

	case number.kind == NumberSmallInt:
		return out.SetInt64(number.smallInt), big.Exact

	case number.kind == NumberBigInt:
		return out.Set(numberBigInt(number)), big.Exact

	case number.kind == NumberDecimal:
		remainder := new(big.Int)

		out.QuoRem(numberBigInt(number), powerOfTen(number.scale), remainder)

		if remainder.Sign() == 0 {
			return out, big.Exact
		}

		if numberSign(number) < 0 {
			return out, big.Above
		}

		return out, big.Below

	default:
		return out.SetInt64(0), big.Exact
	}
}

func (number *Number) String() string {
	return number.Format(DefaultFractionDigits)
}

func (number *Number) Format(fractionDigits int) string {
	if fractionDigits < 0 {
		fractionDigits = DefaultFractionDigits
	}

	if fractionDigits > MaxDecimalScale {
		fractionDigits = MaxDecimalScale
	}

	switch {
	case number == nil:
		return "0"

	case number.kind == NumberSmallInt:
		return strconv.FormatInt(number.smallInt, 10)

	case number.kind == NumberBigInt:
		return numberBigInt(number).String()

	case number.kind == NumberDecimal:
		coefficient := numberBigInt(number)
		scale := number.scale

		if scale > fractionDigits {
			coefficient, scale = roundCoefficientToScale(coefficient, scale, fractionDigits)
		}

		return trimDecimalZeros(formatCoefficientAndScale(coefficient, scale))

	default:
		return "0"
	}
}

func formatNumber(number *Number, fractionDigits int) string {
	if number == nil {
		return "0"
	}

	return number.Format(fractionDigits)
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

func numberNegate(number *Number) *Number {
	switch {
	case number == nil:
		return NewSmallNumber(0)

	case number.kind == NumberSmallInt:
		if number.smallInt != math.MinInt64 {
			return NewSmallNumber(-number.smallInt)
		}

		out := big.NewInt(number.smallInt)
		out.Neg(out)

		return NewBigIntNumber(out)

	case number.kind == NumberBigInt:
		out := new(big.Int)
		out.Neg(numberBigInt(number))

		return NewBigIntNumber(out)

	case number.kind == NumberDecimal:
		out := new(big.Int)
		out.Neg(numberBigInt(number))

		return normaliseCoefficientAndScale(out, number.scale)

	default:
		return NewSmallNumber(0)
	}
}

func numberAdd(left *Number, right *Number) *Number {
	if isSmallPair(left, right) {
		out, ok := addSmallNumbers(numberSmallValue(left), numberSmallValue(right))
		if ok {
			return NewSmallNumber(out)
		}
	}

	leftCoefficient, leftScale := numberCoefficientAndScale(left)
	rightCoefficient, rightScale := numberCoefficientAndScale(right)

	leftCoefficient, rightCoefficient, scale := alignCoefficients(leftCoefficient, leftScale, rightCoefficient, rightScale)

	out := new(big.Int)
	out.Add(leftCoefficient, rightCoefficient)

	return normaliseCoefficientAndScale(out, scale)
}

func numberSubtract(left *Number, right *Number) *Number {
	if isSmallPair(left, right) {
		out, ok := subtractSmallNumbers(numberSmallValue(left), numberSmallValue(right))
		if ok {
			return NewSmallNumber(out)
		}
	}

	leftCoefficient, leftScale := numberCoefficientAndScale(left)
	rightCoefficient, rightScale := numberCoefficientAndScale(right)

	leftCoefficient, rightCoefficient, scale := alignCoefficients(leftCoefficient, leftScale, rightCoefficient, rightScale)

	out := new(big.Int)
	out.Sub(leftCoefficient, rightCoefficient)

	return normaliseCoefficientAndScale(out, scale)
}

func numberMultiply(left *Number, right *Number) *Number {
	if isSmallPair(left, right) {
		out, ok := multiplySmallNumbers(numberSmallValue(left), numberSmallValue(right))
		if ok {
			return NewSmallNumber(out)
		}
	}

	leftCoefficient, leftScale := numberCoefficientAndScale(left)
	rightCoefficient, rightScale := numberCoefficientAndScale(right)

	out := new(big.Int)
	out.Mul(leftCoefficient, rightCoefficient)

	return normaliseCoefficientAndScale(out, leftScale+rightScale)
}

func numberDivide(left *Number, right *Number) (*Number, error) {
	if numberSign(right) == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	leftCoefficient, leftScale := numberCoefficientAndScale(left)
	rightCoefficient, rightScale := numberCoefficientAndScale(right)

	numerator := leftCoefficient
	numerator.Mul(numerator, powerOfTen(rightScale+MaxDecimalScale))

	denominator := rightCoefficient
	denominator.Mul(denominator, powerOfTen(leftScale))

	coefficient := roundedQuotient(numerator, denominator)

	return normaliseCoefficientAndScale(coefficient, MaxDecimalScale), nil
}

func numberCompare(left *Number, right *Number) int {
	if isSmallPair(left, right) {
		return compareInt64(numberSmallValue(left), numberSmallValue(right))
	}

	leftCoefficient, leftScale := numberCoefficientAndScale(left)
	rightCoefficient, rightScale := numberCoefficientAndScale(right)

	leftCoefficient, rightCoefficient, _ = alignCoefficients(leftCoefficient, leftScale, rightCoefficient, rightScale)

	return leftCoefficient.Cmp(rightCoefficient)
}

func numberSign(number *Number) int {
	switch {
	case number == nil:
		return 0

	case number.kind == NumberSmallInt:
		return compareInt64(number.smallInt, 0)

	case number.kind == NumberBigInt,
		number.kind == NumberDecimal:
		return numberBigInt(number).Sign()

	default:
		return 0
	}
}

func numberToBigFloat(number *Number) *big.Float {
	return numberToBigFloatWithPrecision(number, FloatPrecision)
}

func numberToBigFloatWithPrecision(number *Number, precision uint) *big.Float {
	if precision < FloatPrecision {
		precision = FloatPrecision
	}

	out := newFloatWithPrecision(precision)

	switch {
	case number == nil:
		return out

	case number.kind == NumberSmallInt:
		out.SetInt64(number.smallInt)
		return out

	case number.kind == NumberBigInt:
		out.SetInt(numberBigInt(number))
		return out

	case number.kind == NumberDecimal:
		out.SetInt(numberBigInt(number))

		divisor := newFloatWithPrecision(precision)
		divisor.SetInt(powerOfTen(number.scale))

		out.Quo(out, divisor)

		return out

	default:
		return out
	}
}

func newDecimalNumberFromBigFloat(value *big.Float, scale int) *Number {
	if value == nil || value.Sign() == 0 {
		return NewSmallNumber(0)
	}

	if scale < 0 {
		scale = 0
	}

	if scale > MaxDecimalScale {
		scale = MaxDecimalScale
	}

	precision := FloatPrecision
	if valuePrecision := value.Prec(); valuePrecision > precision {
		precision = valuePrecision
	}

	precision += uint(scale*4 + 16)

	scaled := newFloatWithPrecision(precision)
	scaled.Set(value)

	multiplier := newFloatWithPrecision(precision)
	multiplier.SetInt(powerOfTen(scale))

	scaled.Mul(scaled, multiplier)

	half := newFloatWithPrecision(precision)
	half.SetFloat64(0.5)

	if scaled.Sign() >= 0 {
		scaled.Add(scaled, half)
	} else {
		scaled.Sub(scaled, half)
	}

	coefficient, _ := scaled.Int(nil)

	return normaliseCoefficientAndScale(coefficient, scale)
}

func normaliseCoefficientAndScale(coefficient *big.Int, scale int) *Number {
	if coefficient == nil || coefficient.Sign() == 0 {
		return NewSmallNumber(0)
	}

	if scale < 0 {
		coefficient.Mul(coefficient, powerOfTen(-scale))
		scale = 0
	}

	if scale > MaxDecimalScale {
		coefficient, scale = roundCoefficientToScale(coefficient, scale, MaxDecimalScale)
	}

	coefficient, scale = trimDecimalTrailingZeros(coefficient, scale)

	if coefficient.Sign() == 0 {
		return NewSmallNumber(0)
	}

	if scale == 0 {
		return NewBigIntNumber(coefficient)
	}

	return &Number{
		kind:   NumberDecimal,
		bigInt: coefficient,
		scale:  scale,
	}
}

func numberCoefficientAndScale(number *Number) (*big.Int, int) {
	switch {
	case number == nil:
		return new(big.Int), 0

	case number.kind == NumberSmallInt:
		return big.NewInt(number.smallInt), 0

	case number.kind == NumberBigInt:
		return new(big.Int).Set(numberBigInt(number)), 0

	case number.kind == NumberDecimal:
		return new(big.Int).Set(numberBigInt(number)), number.scale

	default:
		return new(big.Int), 0
	}
}

func alignCoefficients(left *big.Int, leftScale int, right *big.Int, rightScale int) (*big.Int, *big.Int, int) {
	if leftScale == rightScale {
		return left, right, leftScale
	}

	if leftScale < rightScale {
		left.Mul(left, powerOfTen(rightScale-leftScale))
		return left, right, rightScale
	}

	right.Mul(right, powerOfTen(leftScale-rightScale))

	return left, right, leftScale
}

func roundCoefficientToScale(coefficient *big.Int, scale int, targetScale int) (*big.Int, int) {
	if targetScale < 0 {
		targetScale = 0
	}

	if scale <= targetScale {
		return coefficient, scale
	}

	divisor := powerOfTen(scale - targetScale)

	return roundedQuotient(coefficient, divisor), targetScale
}

func roundedQuotient(numerator *big.Int, denominator *big.Int) *big.Int {
	if denominator.Sign() == 0 {
		return new(big.Int)
	}

	sign := numerator.Sign() * denominator.Sign()

	absoluteNumerator := new(big.Int)
	absoluteNumerator.Abs(numerator)

	absoluteDenominator := new(big.Int)
	absoluteDenominator.Abs(denominator)

	quotient := new(big.Int)
	remainder := new(big.Int)

	quotient.QuoRem(absoluteNumerator, absoluteDenominator, remainder)

	remainder.Mul(remainder, bigTwo)

	if remainder.Cmp(absoluteDenominator) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}

	if sign < 0 {
		quotient.Neg(quotient)
	}

	return quotient
}

func trimDecimalTrailingZeros(coefficient *big.Int, scale int) (*big.Int, int) {
	if coefficient == nil || coefficient.Sign() == 0 {
		return new(big.Int), 0
	}

	quotient := new(big.Int)
	remainder := new(big.Int)

	for scale > 0 {
		quotient.QuoRem(coefficient, bigTen, remainder)
		if remainder.Sign() != 0 {
			break
		}

		coefficient.Set(quotient)
		scale--
	}

	return coefficient, scale
}

func formatCoefficientAndScale(coefficient *big.Int, scale int) string {
	if coefficient == nil || coefficient.Sign() == 0 {
		return "0"
	}

	if scale <= 0 {
		return coefficient.String()
	}

	negative := coefficient.Sign() < 0

	digits := new(big.Int)
	digits.Abs(coefficient)

	text := digits.String()

	if len(text) <= scale {
		text = strings.Repeat("0", scale-len(text)+1) + text
	}

	pointIndex := len(text) - scale

	out := text[:pointIndex] + "." + text[pointIndex:]
	if negative {
		out = "-" + out
	}

	return out
}

func numberBigInt(number *Number) *big.Int {
	if number == nil || number.bigInt == nil {
		return new(big.Int)
	}

	return number.bigInt
}

func powerOfTen(exponent int) *big.Int {
	if exponent <= 0 {
		return decimalPowersOfTen[0]
	}

	if exponent < len(decimalPowersOfTen) {
		return decimalPowersOfTen[exponent]
	}

	return new(big.Int).Exp(bigTen, big.NewInt(int64(exponent)), nil)
}

func buildDecimalPowersOfTen(maxExponent int) []*big.Int {
	powers := make([]*big.Int, maxExponent+1)

	powers[0] = big.NewInt(1)

	for exponent := 1; exponent <= maxExponent; exponent++ {
		powers[exponent] = new(big.Int)
		powers[exponent].Mul(powers[exponent-1], bigTen)
	}

	return powers
}

func isSmallPair(left *Number, right *Number) bool {
	return numberIsSmall(left) && numberIsSmall(right)
}

func numberIsSmall(number *Number) bool {
	return number == nil || number.kind == NumberSmallInt
}

func numberSmallValue(number *Number) int64 {
	if number == nil {
		return 0
	}

	return number.smallInt
}

func compareInt64(left int64, right int64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func addSmallNumbers(left int64, right int64) (int64, bool) {
	out := left + right

	if right > 0 && out < left {
		return 0, false
	}

	if right < 0 && out > left {
		return 0, false
	}

	return out, true
}

func subtractSmallNumbers(left int64, right int64) (int64, bool) {
	out := left - right

	if right > 0 && out > left {
		return 0, false
	}

	if right < 0 && out < left {
		return 0, false
	}

	return out, true
}

func multiplySmallNumbers(left int64, right int64) (int64, bool) {
	if left == 0 || right == 0 {
		return 0, true
	}

	if left == math.MinInt64 && right == -1 {
		return 0, false
	}

	if right == math.MinInt64 && left == -1 {
		return 0, false
	}

	out := left * right

	if out/right != left {
		return 0, false
	}

	return out, true
}

func numberToArrayIndex(index Value) (int, error) {
	index = resolveSpecializedValue(index)

	if index.Kind != ValueNumber {
		return 0, fmt.Errorf("array index must be a number")
	}

	i, accuracy := index.Number.Int64()
	if accuracy != big.Exact {
		return 0, fmt.Errorf("array index must be an integer")
	}

	if int64(int(i)) != i {
		return 0, fmt.Errorf("array index %d out of bounds", i)
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
