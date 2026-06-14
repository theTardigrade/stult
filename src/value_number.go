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

func NewNumberValueFromNumber(number *Number) Value {
	return Value{
		Kind:   ValueNumber,
		Number: number,
	}
}

func NewNumberValueFromString(literal string) (Value, error) {
	if isIntegerNumberLiteral(literal) {
		small, err := strconv.ParseInt(literal, 10, 64)
		if err == nil {
			return NewNumberValueFromInt64(small), nil
		}
	}

	n, _, err := big.ParseFloat(literal, 10, FloatPrecision, big.ToNearestEven)
	if err != nil {
		return Value{}, fmt.Errorf("invalid number %q", literal)
	}

	return NewBigNumberValue(n), nil
}

func NewNumberValueFromInt(value int) Value {
	return NewNumberValueFromInt64(int64(value))
}

func NewNumberValueFromInt64(value int64) Value {
	return NewNumberValueFromNumber(NewSmallNumber(value))
}

func NewNumberValueFromBigInt(value *big.Int) Value {
	number := newFloat()
	number.SetInt(value)

	return NewBigNumberValue(number)
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
	if value != nil && value.Prec() > precision {
		precision = value.Prec()
	}

	out := &Number{
		useBig: true,
		small:  0,
		big:    newFloatWithPrecision(precision),
	}

	if value != nil {
		out.big.Set(value)
	}

	return out
}

func CloneNumber(x *Number) *Number {
	if x == nil {
		return nil
	}

	if !x.useBig {
		return NewSmallNumber(x.small)
	}

	return NewBigNumber(x.big)
}

func newFloat() *big.Float {
	return newFloatWithPrecision(FloatPrecision)
}

func newFloatWithPrecision(precision uint) *big.Float {
	return new(big.Float).
		SetPrec(precision).
		SetMode(big.ToNearestEven)
}

func isIntegerNumberLiteral(literal string) bool {
	return !strings.ContainsAny(literal, ".eE")
}

func (number *Number) IsBig() bool {
	return number != nil && number.useBig
}

func (number *Number) IsSmall() bool {
	return number != nil && !number.useBig
}

func (number *Number) BigFloat() *big.Float {
	return numberToBigFloat(number)
}

func (number *Number) Sign() int {
	if number == nil {
		return 0
	}

	if !number.useBig {
		switch {
		case number.small < 0:
			return -1

		case number.small > 0:
			return 1

		default:
			return 0
		}
	}

	if number.big == nil {
		return 0
	}

	return number.big.Sign()
}

func (number *Number) Cmp(other *Number) int {
	if number == nil && other == nil {
		return 0
	}

	if number == nil {
		return -other.Sign()
	}

	if other == nil {
		return number.Sign()
	}

	if !number.useBig && !other.useBig {
		switch {
		case number.small < other.small:
			return -1

		case number.small > other.small:
			return 1

		default:
			return 0
		}
	}

	left := numberToBigFloat(number)
	right := numberToBigFloat(other)

	return left.Cmp(right)
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
		out.SetInt64(0)
		return out, big.Exact
	}

	if !number.useBig {
		out.SetInt64(number.small)
		return out, big.Exact
	}

	if number.big == nil {
		out.SetInt64(0)
		return out, big.Exact
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

	text := number.big.Text('f', fractionDigits)

	return trimDecimalZeros(text)
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
		if number.small == math.MinInt64 {
			out := numberToBigFloat(number)
			out.Neg(out)

			return NewBigNumber(out)
		}

		return NewSmallNumber(-number.small)
	}

	out := numberToBigFloat(number)
	out.Neg(out)

	return NewBigNumber(out)
}

func numberAdd(left *Number, right *Number) *Number {
	if left != nil && right != nil && !left.useBig && !right.useBig {
		out, ok := addSmallNumbers(left.small, right.small)
		if ok {
			return NewSmallNumber(out)
		}
	}

	leftBig := numberToBigFloat(left)
	rightBig := numberToBigFloat(right)

	out := newFloat()
	out.Add(leftBig, rightBig)

	return NewBigNumber(out)
}

func numberSubtract(left *Number, right *Number) *Number {
	if left != nil && right != nil && !left.useBig && !right.useBig {
		out, ok := subtractSmallNumbers(left.small, right.small)
		if ok {
			return NewSmallNumber(out)
		}
	}

	leftBig := numberToBigFloat(left)
	rightBig := numberToBigFloat(right)

	out := newFloat()
	out.Sub(leftBig, rightBig)

	return NewBigNumber(out)
}

func numberMultiply(left *Number, right *Number) *Number {
	if left != nil && right != nil && !left.useBig && !right.useBig {
		out, ok := multiplySmallNumbers(left.small, right.small)
		if ok {
			return NewSmallNumber(out)
		}
	}

	leftBig := numberToBigFloat(left)
	rightBig := numberToBigFloat(right)

	out := newFloat()
	out.Mul(leftBig, rightBig)

	return NewBigNumber(out)
}

func numberDivide(left *Number, right *Number) (*Number, error) {
	if numberSign(right) == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	leftBig := numberToBigFloat(left)
	rightBig := numberToBigFloat(right)

	out := newFloat()
	out.Quo(leftBig, rightBig)

	return NewBigNumber(out), nil
}

func numberCompare(left *Number, right *Number) int {
	return left.Cmp(right)
}

func numberSign(number *Number) int {
	if number == nil {
		return 0
	}

	return number.Sign()
}

func numberToBigFloat(number *Number) *big.Float {
	return numberToBigFloatWithPrecision(number, FloatPrecision)
}

func numberToBigFloatWithPrecision(number *Number, precision uint) *big.Float {
	out := newFloatWithPrecision(precision)

	if number == nil {
		return out.SetInt64(0)
	}

	if !number.useBig {
		return out.SetInt64(number.small)
	}

	if number.big == nil {
		return out.SetInt64(0)
	}

	return out.Set(number.big)
}

func addSmallNumbers(left int64, right int64) (int64, bool) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, false
	}

	if right < 0 && left < math.MinInt64-right {
		return 0, false
	}

	return left + right, true
}

func subtractSmallNumbers(left int64, right int64) (int64, bool) {
	if right > 0 && left < math.MinInt64+right {
		return 0, false
	}

	if right < 0 && left > math.MaxInt64+right {
		return 0, false
	}

	return left - right, true
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
