package main

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

const FloatPrecision uint = 1024
const DefaultDisplayDigits = 20

type ValueKind int

const (
	ValueNumber ValueKind = iota
	ValueBool
	ValueString
	ValueMap
	ValueFunction
	ValueBuiltinFunction
)

type Function struct {
	Parameters []Token
	Body       []Statement
	Returns    []Expression
	Env        *Environment
}

type BuiltinFunction func(interpreter *Interpreter, args []Value) (Value, error)

type Map struct {
	Entries     map[string]Binding
	Order       []string
	IsImmutable bool
}

type Value struct {
	Kind            ValueKind
	Number          *big.Float
	Bool            bool
	Text            string
	Map             *Map
	Function        *Function
	BuiltinFunction BuiltinFunction
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

func NewStringValue(value string) Value {
	return Value{Kind: ValueString, Text: value}
}

func NewMapValue(entries map[string]Binding, isImmutable bool) Value {
	return Value{
		Kind: ValueMap,
		Map: &Map{
			Entries:     entries,
			IsImmutable: isImmutable,
		},
	}
}

func NewBuiltinFunctionValue(fn BuiltinFunction) Value {
	return Value{
		Kind:            ValueBuiltinFunction,
		BuiltinFunction: fn,
	}
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

func (v Value) PrintString() string {
	switch v.Kind {
	case ValueString:
		return v.Text
	default:
		return v.String()
	}
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

	case ValueString:
		return strconv.Quote(v.Text)

	case ValueMap:
		return formatMap(v.Map, digits)

	case ValueBuiltinFunction:
		return "<builtin function>"

	case ValueFunction:
		return "<function>"

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

	case ValueString:
		return strconv.Quote(v.Text)

	case ValueMap:
		return formatMap(v.Map, DefaultDisplayDigits)

	case ValueFunction:
		return "<function>"

	case ValueBuiltinFunction:
		return "<builtin function>"

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

func formatMap(m *Map, digits int) string {
	keys := make([]string, 0, len(m.Entries))

	for key := range m.Entries {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	parts := make([]string, 0, len(keys))

	for _, key := range keys {
		binding := m.Entries[key]
		parts = append(parts, strconv.Quote(key)+": "+binding.Value.Format(digits))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

type Binding struct {
	Value       Value
	IsImmutable bool
}

type Environment struct {
	values map[string]Binding
	order  []string
	parent *Environment
}

func NewEnvironment() *Environment {
	return NewChildEnvironment(nil)
}

func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		values: make(map[string]Binding),
		order:  []string{},
		parent: parent,
	}
}

func (e *Environment) Get(name string) (Binding, bool) {
	binding, ok := e.values[name]
	if ok {
		return binding, true
	}

	if e.parent != nil {
		return e.parent.Get(name)
	}

	return Binding{}, false
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
	e.order = append(e.order, name)

	return nil
}

func (e *Environment) Dump() {
	for _, name := range e.order {
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
	env := NewEnvironment()

	if err := env.Set("STD", NewStdMap(), true); err != nil {
		panic(err)
	}

	return &Interpreter{Env: env}
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
			return Value{}, fmt.Errorf("line %d, column %d: %w", s.Name.StartOfLine, s.Name.StartOfColumn, err)
		}

		return value, nil

	case *IndexAssignmentStatement:
		return i.evalIndexAssignmentStatement(s)

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

	case *StringLiteral:
		return NewStringValue(e.Value), nil

	case *IdentifierExpression:
		binding, ok := i.Env.Get(e.Name)
		if !ok {
			return Value{}, fmt.Errorf("line %d, column %d: undefined identifier %q", e.Token.StartOfLine, e.Token.StartOfColumn, e.Name)
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

	case *MapLiteral:
		return i.evalMapLiteral(e)

	case *IndexExpression:
		return i.evalIndexExpression(e)

	case *FunctionLiteral:
		return NewFunctionValue(&Function{
			Parameters: e.Parameters,
			Body:       e.Body,
			Returns:    e.Returns,
			Env:        i.Env,
		}), nil

	case *CallExpression:
		return i.evalCallExpression(e)

	default:
		return Value{}, fmt.Errorf("unknown expression type %T", expr)
	}
}

func (i *Interpreter) evalMapLiteral(lit *MapLiteral) (Value, error) {
	entries := make(map[string]Binding)

	for _, entry := range lit.Entries {
		key := entry.Key.Literal

		if _, exists := entries[key]; exists {
			return Value{}, fmt.Errorf(
				"line %d, column %d: duplicate map key %q",
				entry.Key.StartOfLine,
				entry.Key.StartOfColumn,
				key,
			)
		}

		value, err := i.evalExpression(entry.Value)
		if err != nil {
			return Value{}, err
		}

		entries[key] = Binding{
			Value:       value,
			IsImmutable: isImmutableIdentifier(key),
		}
	}

	return NewMapValue(entries, false), nil
}

func (i *Interpreter) evalIndexExpression(expr *IndexExpression) (Value, error) {
	object, err := i.evalExpression(expr.Object)
	if err != nil {
		return Value{}, err
	}

	index, err := i.evalExpression(expr.Index)
	if err != nil {
		return Value{}, err
	}

	if object.Kind != ValueMap {
		return Value{}, fmt.Errorf("cannot index non-map value")
	}

	if index.Kind != ValueString {
		return Value{}, fmt.Errorf("map index must be a string")
	}

	binding, ok := object.Map.Entries[index.Text]
	if !ok {
		return Value{}, fmt.Errorf("map has no key %q", index.Text)
	}

	return binding.Value, nil
}

func (i *Interpreter) evalIndexAssignmentStatement(stmt *IndexAssignmentStatement) (Value, error) {
	object, err := i.evalExpression(stmt.Target.Object)
	if err != nil {
		return Value{}, err
	}

	index, err := i.evalExpression(stmt.Target.Index)
	if err != nil {
		return Value{}, err
	}

	if object.Kind != ValueMap {
		return Value{}, fmt.Errorf("cannot assign into non-map value")
	}

	if object.Map.IsImmutable {
		return Value{}, fmt.Errorf("cannot modify immutable map")
	}

	if index.Kind != ValueString {
		return Value{}, fmt.Errorf("map index must be a string")
	}

	binding, exists := object.Map.Entries[index.Text]
	if exists && binding.IsImmutable {
		return Value{}, fmt.Errorf("cannot reassign immutable map entry %q", index.Text)
	}

	value, err := i.evalExpression(stmt.Value)
	if err != nil {
		return Value{}, err
	}

	if exists {
		binding.Value = value
		object.Map.Entries[index.Text] = binding
		return value, nil
	}

	object.Map.Entries[index.Text] = Binding{
		Value:       value,
		IsImmutable: isImmutableIdentifier(index.Text),
	}

	return value, nil
}

func (i *Interpreter) evalCallExpression(call *CallExpression) (Value, error) {
	callee, err := i.evalExpression(call.Callee)
	if err != nil {
		return Value{}, err
	}

	args := make([]Value, 0, len(call.Arguments))

	for _, argExpr := range call.Arguments {
		arg, err := i.evalExpression(argExpr)
		if err != nil {
			return Value{}, err
		}

		args = append(args, arg)
	}

	switch callee.Kind {
	case ValueBuiltinFunction:
		return callee.BuiltinFunction(i, args)

	case ValueFunction:
		return i.callFunction(callee.Function, args)

	default:
		return Value{}, fmt.Errorf("value is not callable")
	}
}

func (i *Interpreter) callFunction(fn *Function, args []Value) (Value, error) {
	if len(args) != len(fn.Parameters) {
		return Value{}, fmt.Errorf(
			"function expected %d argument(s), got %d",
			len(fn.Parameters),
			len(args),
		)
	}

	callEnv := NewChildEnvironment(fn.Env)

	for index, parameter := range fn.Parameters {
		if err := callEnv.Set(parameter.Literal, args[index], parameter.IsImmutable); err != nil {
			return Value{}, fmt.Errorf(
				"line %d, column %d: %w",
				parameter.StartOfLine,
				parameter.StartOfColumn,
				err,
			)
		}
	}

	previousEnv := i.Env
	i.Env = callEnv
	defer func() {
		i.Env = previousEnv
	}()

	for _, stmt := range fn.Body {
		if _, err := i.evalStatement(stmt); err != nil {
			return Value{}, err
		}
	}

	if len(fn.Returns) != 1 {
		return Value{}, fmt.Errorf("functions must return exactly one value for now")
	}

	return i.evalExpression(fn.Returns[0])
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

	if operator == "+" && left.Kind == ValueString && right.Kind == ValueString {
		return NewStringValue(left.Text + right.Text), nil
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

	case ValueString:
		return left.Text == right.Text, nil

	case ValueMap:
		return false, fmt.Errorf("cannot compare maps")

	case ValueFunction, ValueBuiltinFunction:
		return false, fmt.Errorf("cannot compare functions")

	default:
		return false, fmt.Errorf("cannot compare unknown value kind")
	}
}

func newFloat() *big.Float {
	return newFloatWithPrecision(FloatPrecision)
}

func newFloatWithPrecision(precision uint) *big.Float {
	return new(big.Float).
		SetPrec(precision).
		SetMode(big.ToNearestEven)
}

func NewFunctionValue(fn *Function) Value {
	return Value{
		Kind:     ValueFunction,
		Function: fn,
	}
}
