package main

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

const FloatPrecision uint = 2048
const DefaultFractionDigits = 32

type Number struct {
	useBig bool
	small  int64
	big    *big.Float
}

func NewNumberValueFromString(literal string) (Value, error) {
	if isIntegerNumberLiteral(literal) {
		small, err := strconv.ParseInt(literal, 10, 64)
		if err == nil {
			return NewNumberValueFromInt64(small), nil
		}
	}

	number, _, err := big.ParseFloat(literal, 10, FloatPrecision, big.ToNearestEven)
	if err != nil {
		return Value{}, fmt.Errorf("invalid number %q", literal)
	}

	return NewBigNumberValue(number), nil
}

func NewNumberValueFromInt(value int) Value {
	return NewNumberValueFromInt64(int64(value))
}

func NewNumberValueFromInt64(value int64) Value {
	return NewNumberValueFromNumber(NewSmallNumber(value))
}

func NewNumberValueFromBigInt(value *big.Int) Value {
	if value == nil {
		return NewNumberValueFromInt64(0)
	}

	if value.IsInt64() {
		return NewNumberValueFromInt64(value.Int64())
	}

	precision := FloatPrecision
	if bitLen := uint(value.BitLen()); bitLen > precision {
		precision = bitLen
	}

	number := newFloatWithPrecision(precision)
	number.SetInt(value)

	return newBigNumberValueFromOwnedFloat(number)
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
		useBig: false,
		small:  value,
		big:    nil,
	}
}

func NewBigNumber(value *big.Float) *Number {
	precision := FloatPrecision
	if value != nil {
		if p := value.Prec(); p > precision {
			precision = p
		}
	}

	out := newFloatWithPrecision(precision)
	if value != nil {
		out.Set(value)
	}

	return newBigNumberFromOwnedFloat(out)
}

func CloneNumber(number *Number) *Number {
	if number == nil {
		return NewSmallNumber(0)
	}

	if !number.useBig {
		return NewSmallNumber(number.small)
	}

	return NewBigNumber(number.big)
}

func newBigNumberValueFromOwnedFloat(value *big.Float) Value {
	return NewNumberValueFromNumber(newBigNumberFromOwnedFloat(value))
}

func newBigNumberFromOwnedFloat(value *big.Float) *Number {
	if value == nil {
		value = newFloat()
	} else if value.Prec() < FloatPrecision {
		out := newFloat()
		out.Set(value)
		value = out
	} else {
		value.SetMode(big.ToNearestEven)
	}

	return &Number{
		useBig: true,
		small:  0,
		big:    value,
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
	return numberIsBig(number)
}

func (number *Number) IsSmall() bool {
	return numberIsSmall(number)
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
	if number == nil {
		return 0, big.Exact
	}

	if !number.useBig {
		return number.small, big.Exact
	}

	if number.big == nil {
		return 0, big.Exact
	}

	return number.big.Int64()
}

func (number *Number) Int(out *big.Int) (*big.Int, big.Accuracy) {
	if out == nil {
		out = new(big.Int)
	}

	if number == nil {
		return out.SetInt64(0), big.Exact
	}

	if !number.useBig {
		return out.SetInt64(number.small), big.Exact
	}

	if number.big == nil {
		return out.SetInt64(0), big.Exact
	}

	return number.big.Int(out)
}

func (number *Number) String() string {
	return number.Format(DefaultFractionDigits)
}

func (number *Number) Format(fractionDigits int) string {
	if number == nil {
		return "0"
	}

	if !number.useBig {
		return strconv.FormatInt(number.small, 10)
	}

	if number.big == nil {
		return "0"
	}

	if fractionDigits < 0 {
		fractionDigits = DefaultFractionDigits
	}

	return trimDecimalZeros(number.big.Text('f', fractionDigits))
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
	if number == nil {
		return NewSmallNumber(0)
	}

	if !number.useBig {
		if number.small != math.MinInt64 {
			return NewSmallNumber(-number.small)
		}

		out := newFloatFromInt64(number.small, FloatPrecision)
		out.Neg(out)

		return newBigNumberFromOwnedFloat(out)
	}

	out := newFloatWithPrecision(numberPrecision(number))
	if number.big != nil {
		out.Neg(number.big)
	}

	return newBigNumberFromOwnedFloat(out)
}

func numberAdd(left *Number, right *Number) *Number {
	if numberIsSmall(left) && numberIsSmall(right) {
		out, ok := addSmallNumbers(numberSmallValue(left), numberSmallValue(right))
		if ok {
			return NewSmallNumber(out)
		}
	}

	leftFloat, rightFloat, precision := promoteNumbersToBig(left, right)

	out := newFloatWithPrecision(precision)
	out.Add(leftFloat, rightFloat)

	return newBigNumberFromOwnedFloat(out)
}

func numberSubtract(left *Number, right *Number) *Number {
	if numberIsSmall(left) && numberIsSmall(right) {
		out, ok := subtractSmallNumbers(numberSmallValue(left), numberSmallValue(right))
		if ok {
			return NewSmallNumber(out)
		}
	}

	leftFloat, rightFloat, precision := promoteNumbersToBig(left, right)

	out := newFloatWithPrecision(precision)
	out.Sub(leftFloat, rightFloat)

	return newBigNumberFromOwnedFloat(out)
}

func numberMultiply(left *Number, right *Number) *Number {
	if numberIsSmall(left) && numberIsSmall(right) {
		out, ok := multiplySmallNumbers(numberSmallValue(left), numberSmallValue(right))
		if ok {
			return NewSmallNumber(out)
		}
	}

	leftFloat, rightFloat, precision := promoteNumbersToBig(left, right)

	out := newFloatWithPrecision(precision)
	out.Mul(leftFloat, rightFloat)

	return newBigNumberFromOwnedFloat(out)
}

func numberDivide(left *Number, right *Number) (*Number, error) {
	if numberSign(right) == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	if numberIsSmall(left) && numberIsSmall(right) {
		leftFloat := newFloatFromInt64(numberSmallValue(left), FloatPrecision)
		rightFloat := newFloatFromInt64(numberSmallValue(right), FloatPrecision)

		out := newFloat()
		out.Quo(leftFloat, rightFloat)

		return newBigNumberFromOwnedFloat(out), nil
	}

	leftFloat, rightFloat, precision := promoteNumbersToBig(left, right)

	out := newFloatWithPrecision(precision)
	out.Quo(leftFloat, rightFloat)

	return newBigNumberFromOwnedFloat(out), nil
}

func numberCompare(left *Number, right *Number) int {
	if numberIsSmall(left) && numberIsSmall(right) {
		return compareInt64(numberSmallValue(left), numberSmallValue(right))
	}

	if numberIsBig(left) && numberIsBig(right) {
		return numberBigFloatReference(left).Cmp(numberBigFloatReference(right))
	}

	leftFloat, rightFloat, _ := promoteNumbersToBig(left, right)

	return leftFloat.Cmp(rightFloat)
}

func numberSign(number *Number) int {
	if number == nil {
		return 0
	}

	if !number.useBig {
		return compareInt64(number.small, 0)
	}

	if number.big == nil {
		return 0
	}

	return number.big.Sign()
}

func numberToBigFloat(number *Number) *big.Float {
	return numberToBigFloatWithPrecision(number, numberPrecision(number))
}

func numberToBigFloatWithPrecision(number *Number, precision uint) *big.Float {
	if precision < FloatPrecision {
		precision = FloatPrecision
	}

	out := newFloatWithPrecision(precision)

	if number == nil {
		return out
	}

	if !number.useBig {
		out.SetInt64(number.small)
		return out
	}

	if number.big != nil {
		out.Set(number.big)
	}

	return out
}

func promoteNumbersToBig(left *Number, right *Number) (*big.Float, *big.Float, uint) {
	precision := numberPairPrecision(left, right)

	leftFloat := promoteNumberToBigWithPrecision(left, precision)
	rightFloat := promoteNumberToBigWithPrecision(right, precision)

	return leftFloat, rightFloat, precision
}

func promoteNumberToBigWithPrecision(number *Number, precision uint) *big.Float {
	if precision < FloatPrecision {
		precision = FloatPrecision
	}

	if number == nil {
		return newFloatWithPrecision(precision)
	}

	if number.useBig {
		if number.big == nil {
			number.big = newFloatWithPrecision(precision)
		}

		return number.big
	}

	number.useBig = true
	number.big = newFloatFromInt64(number.small, precision)
	number.small = 0

	return number.big
}

func numberBigFloatReference(number *Number) *big.Float {
	if number == nil {
		return newFloat()
	}

	if !number.useBig {
		return promoteNumberToBigWithPrecision(number, FloatPrecision)
	}

	if number.big == nil {
		number.big = newFloatWithPrecision(numberPrecision(number))
	}

	return number.big
}

func numberPairPrecision(left *Number, right *Number) uint {
	leftPrecision := numberPrecision(left)
	rightPrecision := numberPrecision(right)

	if leftPrecision > rightPrecision {
		return leftPrecision
	}

	return rightPrecision
}

func numberPrecision(number *Number) uint {
	if number == nil {
		return FloatPrecision
	}

	if number.useBig && number.big != nil {
		if precision := number.big.Prec(); precision > FloatPrecision {
			return precision
		}
	}

	return FloatPrecision
}

func numberIsBig(number *Number) bool {
	return number != nil && number.useBig
}

func numberIsSmall(number *Number) bool {
	return number == nil || !number.useBig
}

func numberSmallValue(number *Number) int64 {
	if number == nil {
		return 0
	}

	return number.small
}

func newFloatFromInt64(value int64, precision uint) *big.Float {
	if precision < FloatPrecision {
		precision = FloatPrecision
	}

	out := newFloatWithPrecision(precision)
	out.SetInt64(value)

	return out
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
