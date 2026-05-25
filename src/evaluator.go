package main

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

const FloatPrecision uint = 1024

const DefaultDisplayDigits = 20

// ValueKind identifies the kind of runtime value.
// For now, the language only has Number and Bool.
type ValueKind int

const (
	ValueNumber ValueKind = iota
	ValueBool
)

type Value struct {
	Kind   ValueKind
	Number *big.Float
	Bool   bool
}

func NewNumberValueFromString(literal string) (Value, error) {
	n, _, err := big.ParseFloat(literal, 10, FloatPrecision, big.ToNearestEven)
	if err != nil {
		return Value{}, fmt.Errorf("invalid number %q", literal)
	}

	return Value{Kind: ValueNumber, Number: n}, nil
}

func NewBoolValue(value bool) Value {
	return Value{Kind: ValueBool, Bool: value}
}

func CloneNumber(x *big.Float) *big.Float {
	return new(big.Float).
		SetPrec(FloatPrecision).
		SetMode(big.ToNearestEven).
		Set(x)
}

func (v Value) String() string {
	return v.Format(DefaultDisplayDigits)
}

func (v Value) Format(digits int) string {
	switch v.Kind {
	case ValueNumber:
		return formatNumber(v.Number, digits)

	case ValueBool:
		if v.Bool {
			return "true"
		}
		return "false"

	default:
		return "<unknown>"
	}
}

func (v Value) DebugString() string {
	switch v.Kind {
	case ValueNumber:
		return v.Number.Text('g', -1)

	case ValueBool:
		return v.String()

	default:
		return "<unknown>"
	}
}

func formatNumber(x *big.Float, digits int) string {
	if digits <= 0 {
		digits = DefaultDisplayDigits
	}

	text := x.Text('g', digits)

	if idx := strings.IndexAny(text, "eE"); idx >= 0 {
		mantissa := trimDecimalZeros(text[:idx])
		return mantissa + text[idx:]
	}

	return trimDecimalZeros(text)
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

type Binding struct {
	Value       Value
	IsImmutable bool
}

type Environment struct {
	values map[string]Binding
}

func NewEnvironment() *Environment {
	return &Environment{values: make(map[string]Binding)}
}

func (e *Environment) Get(name string) (Binding, bool) {
	binding, ok := e.values[name]
	return binding, ok
}

func (e *Environment) Set(name string, value Value, isImmutable bool) error {
	if existing, ok := e.values[name]; ok {
		if existing.IsImmutable {
			return fmt.Errorf("cannot reassign immutable constant %q", name)
		}

		e.values[name] = Binding{
			Value:       value,
			IsImmutable: existing.IsImmutable,
		}
		return nil
	}

	e.values[name] = Binding{
		Value:       value,
		IsImmutable: isImmutable,
	}
	return nil
}

func (e *Environment) Dump() {
	names := make([]string, 0, len(e.values))
	for name := range e.values {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		binding := e.values[name]

		mutability := "mutable"
		if binding.IsImmutable {
			mutability = "immutable"
		}

		fmt.Printf("%s = %s (%s)\n", name, binding.Value.String(), mutability)
	}
}

type Interpreter struct {
	Env *Environment
}

func NewInterpreter() *Interpreter {
	return &Interpreter{Env: NewEnvironment()}
}

func (i *Interpreter) EvalProgram(program *Program) error {
	for _, stmt := range program.Statements {
		if _, err := i.evalStatement(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) evalStatement(stmt Statement) (Value, error) {
	switch s := stmt.(type) {
	case *AssignmentStatement:
		value, err := i.evalExpression(s.Value)
		if err != nil {
			return Value{}, err
		}

		if err := i.Env.Set(s.Name.Literal, value, s.IsImmutable); err != nil {
			return Value{}, fmt.Errorf("line %d, column %d: %w", s.Name.Line, s.Name.Column, err)
		}

		return value, nil

	case *ExpressionStatement:
		return i.evalExpression(s.Expression)

	default:
		return Value{}, fmt.Errorf("unknown statement type %T", stmt)
	}
}

func (i *Interpreter) evalExpression(expr Expression) (Value, error) {
	switch e := expr.(type) {
	case *NumberLiteral:
		return NewNumberValueFromString(e.Value)

	case *IdentifierExpression:
		binding, ok := i.Env.Get(e.Name)
		if !ok {
			return Value{}, fmt.Errorf("line %d, column %d: undefined identifier %q", e.Token.Line, e.Token.Column, e.Name)
		}
		return binding.Value, nil

	case *PrefixExpression:
		value, err := i.evalExpression(e.Right)
		if err != nil {
			return Value{}, err
		}
		return evalPrefix(e.Operator, value)

	case *BinaryExpression:
		left, err := i.evalExpression(e.Left)
		if err != nil {
			return Value{}, err
		}

		right, err := i.evalExpression(e.Right)
		if err != nil {
			return Value{}, err
		}

		return evalBinary(e.Operator, left, right)

	default:
		return Value{}, fmt.Errorf("unknown expression type %T", expr)
	}
}

func evalPrefix(operator string, right Value) (Value, error) {
	switch operator {
	case "-":
		if right.Kind != ValueNumber {
			return Value{}, fmt.Errorf("unary '-' requires a number")
		}

		out := CloneNumber(right.Number)
		out.Neg(out)

		return Value{Kind: ValueNumber, Number: out}, nil

	default:
		return Value{}, fmt.Errorf("unknown prefix operator %q", operator)
	}
}

func evalBinary(operator string, left, right Value) (Value, error) {
	if operator == "==" || operator == "!=" {
		equal, err := valuesEqual(left, right)
		if err != nil {
			return Value{}, err
		}

		if operator == "!=" {
			equal = !equal
		}

		return NewBoolValue(equal), nil
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return Value{}, fmt.Errorf("operator %q requires numbers", operator)
	}

	switch operator {
	case "+":
		out := newFloat()
		out.Add(left.Number, right.Number)
		return Value{Kind: ValueNumber, Number: out}, nil

	case "-":
		out := newFloat()
		out.Sub(left.Number, right.Number)
		return Value{Kind: ValueNumber, Number: out}, nil

	case "*":
		out := newFloat()
		out.Mul(left.Number, right.Number)
		return Value{Kind: ValueNumber, Number: out}, nil

	case "/":
		if right.Number.Sign() == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}

		out := newFloat()
		out.Quo(left.Number, right.Number)
		return Value{Kind: ValueNumber, Number: out}, nil

	case "<":
		return NewBoolValue(left.Number.Cmp(right.Number) < 0), nil

	case "<=":
		return NewBoolValue(left.Number.Cmp(right.Number) <= 0), nil

	case ">":
		return NewBoolValue(left.Number.Cmp(right.Number) > 0), nil

	case ">=":
		return NewBoolValue(left.Number.Cmp(right.Number) >= 0), nil

	default:
		return Value{}, fmt.Errorf("unknown binary operator %q", operator)
	}
}

func valuesEqual(left, right Value) (bool, error) {
	if left.Kind != right.Kind {
		return false, nil
	}

	switch left.Kind {
	case ValueNumber:
		return left.Number.Cmp(right.Number) == 0, nil

	case ValueBool:
		return left.Bool == right.Bool, nil

	default:
		return false, fmt.Errorf("cannot compare unknown value kind")
	}
}

func newFloat() *big.Float {
	return new(big.Float).
		SetPrec(FloatPrecision).
		SetMode(big.ToNearestEven)
}
