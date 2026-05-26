package main

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

const FloatPrecision uint = 2048
const DefaultFractionDigits = 32

type ValueKind int

const (
	ValueEmpty ValueKind = iota
	ValueNumber
	ValueBool
	ValueString
	ValueMap
	ValueArray
	ValueEmptyCollection
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

type String struct {
	Runes       []rune
	IsImmutable bool
}

type Map struct {
	Entries     map[string]Binding
	IsImmutable bool
}

type Array struct {
	Elements    []Value
	IsImmutable bool
}

type EmptyCollection struct {
	Specialized *Value
	IsImmutable bool
}

type Value struct {
	Kind            ValueKind
	Number          *big.Float
	Bool            bool
	Text            *String
	Map             *Map
	Array           *Array
	EmptyCollection *EmptyCollection
	Function        *Function
	BuiltinFunction BuiltinFunction
}

type breakSignal struct {
	Token Token
}

func (s *breakSignal) Error() string {
	return fmt.Sprintf(
		"line %d, column %d: break used outside loop",
		s.Token.StartOfLine,
		s.Token.StartOfColumn,
	)
}

type returnSignal struct {
	Token Token
	Value Value
}

func (s *returnSignal) Error() string {
	return fmt.Sprintf(
		"line %d, column %d: return used outside function",
		s.Token.StartOfLine,
		s.Token.StartOfColumn,
	)
}

func asBreakSignal(err error) (*breakSignal, bool) {
	var signal *breakSignal
	if errors.As(err, &signal) {
		return signal, true
	}

	return nil, false
}

func asReturnSignal(err error) (*returnSignal, bool) {
	var signal *returnSignal
	if errors.As(err, &signal) {
		return signal, true
	}

	return nil, false
}

func NewEmptyValue() Value {
	return Value{Kind: ValueEmpty}
}

func NewEmptyCollectionValue() Value {
	return NewEmptyCollectionValueWithImmutability(false)
}

func NewEmptyCollectionValueWithImmutability(isImmutable bool) Value {
	return Value{
		Kind: ValueEmptyCollection,
		EmptyCollection: &EmptyCollection{
			IsImmutable: isImmutable,
		},
	}
}

func NewNumberValueFromString(literal string) (Value, error) {
	n, _, err := big.ParseFloat(literal, 10, FloatPrecision, big.ToNearestEven)
	if err != nil {
		return Value{}, fmt.Errorf("invalid number %q", literal)
	}

	return Value{Kind: ValueNumber, Number: n}, nil
}

func NewNumberValueFromInt(value int) Value {
	n := new(big.Float).
		SetPrec(FloatPrecision).
		SetMode(big.ToNearestEven).
		SetInt64(int64(value))

	return Value{Kind: ValueNumber, Number: n}
}

func NewBoolValue(value bool) Value {
	return Value{Kind: ValueBool, Bool: value}
}

func NewStringValue(value string) Value {
	return NewStringValueWithImmutability(value, false)
}

func NewStringValueWithImmutability(value string, isImmutable bool) Value {
	return Value{
		Kind: ValueString,
		Text: &String{
			Runes:       []rune(value),
			IsImmutable: isImmutable,
		},
	}
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

func NewArrayValue(elements []Value, isImmutable bool) Value {
	return Value{
		Kind: ValueArray,
		Array: &Array{
			Elements:    elements,
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

func NewFunctionValue(fn *Function) Value {
	return Value{
		Kind:     ValueFunction,
		Function: fn,
	}
}

func CloneNumber(x *big.Float) *big.Float {
	return new(big.Float).
		SetPrec(FloatPrecision).
		SetMode(big.ToNearestEven).
		Set(x)
}

func (s *String) String() string {
	if s == nil {
		return ""
	}

	return string(s.Runes)
}

func (v Value) String() string {
	return v.Format(DefaultFractionDigits)
}

func (v Value) PrintString() string {
	v = resolveSpecializedValue(v)

	switch v.Kind {
	case ValueString:
		return v.Text.String()
	default:
		return v.String()
	}
}

func (v Value) Format(fractionDigits int) string {
	v = resolveSpecializedValue(v)

	switch v.Kind {
	case ValueEmpty:
		return "_"

	case ValueEmptyCollection:
		return "{}"

	case ValueNumber:
		return formatNumber(v.Number, fractionDigits)

	case ValueBool:
		if v.Bool {
			return "true"
		}
		return "false"

	case ValueString:
		return strconv.Quote(v.Text.String())

	case ValueMap:
		return formatMap(v.Map, fractionDigits)

	case ValueArray:
		return formatArray(v.Array, fractionDigits)

	case ValueBuiltinFunction:
		return "<builtin function>"

	case ValueFunction:
		return "<function>"

	default:
		return "<unknown>"
	}
}

func (v Value) DebugString() string {
	v = resolveSpecializedValue(v)

	switch v.Kind {
	case ValueEmpty:
		return "_"

	case ValueEmptyCollection:
		return "{}"

	case ValueNumber:
		return formatNumber(v.Number, DefaultFractionDigits)

	case ValueBool:
		return v.String()

	case ValueString:
		return strconv.Quote(v.Text.String())

	case ValueMap:
		return formatMap(v.Map, DefaultFractionDigits)

	case ValueArray:
		return formatArray(v.Array, DefaultFractionDigits)

	case ValueFunction:
		return "<function>"

	case ValueBuiltinFunction:
		return "<builtin function>"

	default:
		return "<unknown>"
	}
}

func resolveSpecializedValue(value Value) Value {
	for value.Kind == ValueEmptyCollection &&
		value.EmptyCollection != nil &&
		value.EmptyCollection.Specialized != nil {
		value = *value.EmptyCollection.Specialized
	}

	return value
}

func formatNumber(x *big.Float, fractionDigits int) string {
	if fractionDigits < 0 {
		fractionDigits = DefaultFractionDigits
	}

	text := x.Text('f', fractionDigits)

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

func formatMap(m *Map, fractionDigits int) string {
	keys := sortedMapKeys(m)

	parts := make([]string, 0, len(keys))

	for _, key := range keys {
		binding := m.Entries[key]
		parts = append(parts, strconv.Quote(key)+": "+binding.Value.Format(fractionDigits))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func sortedMapKeys(m *Map) []string {
	keys := make([]string, 0, len(m.Entries))

	for key := range m.Entries {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func formatArray(a *Array, fractionDigits int) string {
	parts := make([]string, 0, len(a.Elements))

	for _, element := range a.Elements {
		parts = append(parts, element.Format(fractionDigits))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

type Binding struct {
	Value       Value
	IsImmutable bool
}

func NewImmutableBinding(value Value) Binding {
	return Binding{
		Value:       value,
		IsImmutable: true,
	}
}

type Environment struct {
	values map[string]Binding
	parent *Environment
}

func NewEnvironment() *Environment {
	return NewChildEnvironment(nil)
}

func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		values: make(map[string]Binding),
		parent: parent,
	}
}

func (e *Environment) Get(name string) (Binding, bool) {
	binding, ok := e.values[name]
	return binding, ok
}

func (e *Environment) GetOuter(name string) (Binding, bool) {
	for env := e.parent; env != nil; env = env.parent {
		if binding, ok := env.values[name]; ok {
			return binding, true
		}
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

	return nil
}

func (e *Environment) SetOuter(name string, value Value) error {
	for env := e.parent; env != nil; env = env.parent {
		existing, ok := env.values[name]
		if !ok {
			continue
		}

		if existing.IsImmutable {
			return fmt.Errorf("cannot reassign immutable outer constant %q", name)
		}

		env.values[name] = Binding{
			Value:       value,
			IsImmutable: existing.IsImmutable,
		}

		return nil
	}

	return fmt.Errorf("no outer binding named %q", name)
}

func (e *Environment) Dump() {
	keys := make([]string, 0, len(e.values))

	for name := range e.values {
		keys = append(keys, name)
	}

	sort.Strings(keys)

	for _, name := range keys {
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

		if s.IsOuter {
			if err := i.Env.SetOuter(s.Name.Literal, value); err != nil {
				return Value{}, fmt.Errorf("line %d, column %d: %w", s.Name.StartOfLine, s.Name.StartOfColumn, err)
			}
		} else {
			if err := i.Env.Set(s.Name.Literal, value, s.IsImmutable); err != nil {
				return Value{}, fmt.Errorf("line %d, column %d: %w", s.Name.StartOfLine, s.Name.StartOfColumn, err)
			}
		}

		return value, nil

	case *CompoundAssignmentStatement:
		return i.evalCompoundAssignmentStatement(s)

	case *BreakStatement:
		return Value{}, &breakSignal{Token: s.Token}

	case *ReturnStatement:
		value, err := i.evalExpression(s.Value)
		if err != nil {
			return Value{}, err
		}

		return Value{}, &returnSignal{
			Token: s.Token,
			Value: value,
		}

	case *IndexAssignmentStatement:
		return i.evalIndexAssignmentStatement(s)

	case *ExpressionStatement:
		return i.evalExpression(s.Expression)

	case *ConditionalStatement:
		return i.evalConditionalStatement(s)

	case *LoopStatement:
		return i.evalLoopStatement(s)

	default:
		return Value{}, fmt.Errorf("unknown statement type %T", stmt)
	}
}

func (i *Interpreter) evalConditionalStatement(stmt *ConditionalStatement) (Value, error) {
	for _, branch := range stmt.Branches {
		condition, err := i.evalExpression(branch.Condition)
		if err != nil {
			return Value{}, err
		}

		if condition.Kind != ValueBool {
			return Value{}, fmt.Errorf("conditional expression must evaluate to a bool")
		}

		if condition.Bool {
			return i.evalStatementBlock(branch.Body)
		}
	}

	if stmt.ElseBody != nil {
		return i.evalStatementBlock(stmt.ElseBody)
	}

	return NewEmptyValue(), nil
}

func (i *Interpreter) evalLoopStatement(stmt *LoopStatement) (Value, error) {
	if len(stmt.RangeParameters) > 0 {
		return i.evalCollectionRangeLoopStatement(stmt)
	}

	return i.evalWhileLoopStatement(stmt)
}

func (i *Interpreter) evalWhileLoopStatement(stmt *LoopStatement) (Value, error) {
	for {
		condition, err := i.evalExpression(stmt.Condition)
		if err != nil {
			return Value{}, err
		}

		if condition.Kind != ValueBool {
			return Value{}, fmt.Errorf("loop condition must evaluate to a bool")
		}

		if !condition.Bool {
			break
		}

		if _, err := i.evalStatementBlock(stmt.Body); err != nil {
			if _, ok := asBreakSignal(err); ok {
				break
			}

			return Value{}, err
		}
	}

	return i.evalAfterLoopBody(stmt)
}

func (i *Interpreter) evalCollectionRangeLoopStatement(stmt *LoopStatement) (Value, error) {
	if !isValidCollectionRangeParameterCount(len(stmt.RangeParameters)) {
		return Value{}, fmt.Errorf("collection range loop must have one, two, three, or four parameters")
	}

	iterable, err := i.evalExpression(stmt.Condition)
	if err != nil {
		return Value{}, err
	}

	iterable = resolveSpecializedValue(iterable)

	switch iterable.Kind {
	case ValueMap:
		return i.evalMapRangeLoopStatement(stmt, iterable.Map)

	case ValueArray:
		return i.evalArrayRangeLoopStatement(stmt, iterable.Array)

	case ValueString:
		return i.evalStringRangeLoopStatement(stmt, iterable.Text)

	case ValueEmptyCollection:
		return i.evalAfterLoopBody(stmt)

	default:
		return Value{}, fmt.Errorf("collection range loop expression must evaluate to a map, array, string, or empty collection")
	}
}

func isValidCollectionRangeParameterCount(count int) bool {
	return count >= 1 && count <= 4
}

func (i *Interpreter) evalMapRangeLoopStatement(stmt *LoopStatement, m *Map) (Value, error) {
	position := 0
	lastKey := ""
	hasLastKey := false

	for {
		key, ok := nextMapRangeKey(m, lastKey, hasLastKey)
		if !ok {
			break
		}

		binding, ok := m.Entries[key]
		if !ok {
			continue
		}

		loopBindings := collectionRangeBindings(
			stmt.RangeParameters,
			binding.Value,
			NewStringValue(key),
			Value{Kind: ValueMap, Map: m},
			NewNumberValueFromInt(position),
		)

		if _, err := i.evalStatementBlockWithBindings(stmt.Body, loopBindings); err != nil {
			if _, ok := asBreakSignal(err); ok {
				break
			}

			return Value{}, err
		}

		lastKey = key
		hasLastKey = true
		position++
	}

	return i.evalAfterLoopBody(stmt)
}

func nextMapRangeKey(m *Map, lastKey string, hasLastKey bool) (string, bool) {
	keys := sortedMapKeys(m)

	for _, key := range keys {
		if !hasLastKey || key > lastKey {
			return key, true
		}
	}

	return "", false
}

func (i *Interpreter) evalArrayRangeLoopStatement(stmt *LoopStatement, a *Array) (Value, error) {
	index := 0

	for index < len(a.Elements) {
		key := NewNumberValueFromInt(index)

		loopBindings := collectionRangeBindings(
			stmt.RangeParameters,
			a.Elements[index],
			key,
			Value{Kind: ValueArray, Array: a},
			key,
		)

		if _, err := i.evalStatementBlockWithBindings(stmt.Body, loopBindings); err != nil {
			if _, ok := asBreakSignal(err); ok {
				break
			}

			return Value{}, err
		}

		index++
	}

	return i.evalAfterLoopBody(stmt)
}

func (i *Interpreter) evalStringRangeLoopStatement(stmt *LoopStatement, s *String) (Value, error) {
	if s == nil {
		return Value{}, fmt.Errorf("cannot range over invalid string")
	}

	index := 0

	for index < len(s.Runes) {
		key := NewNumberValueFromInt(index)

		loopBindings := collectionRangeBindings(
			stmt.RangeParameters,
			NewStringValue(string(s.Runes[index])),
			key,
			Value{Kind: ValueString, Text: s},
			key,
		)

		if _, err := i.evalStatementBlockWithBindings(stmt.Body, loopBindings); err != nil {
			if _, ok := asBreakSignal(err); ok {
				break
			}

			return Value{}, err
		}

		index++
	}

	return i.evalAfterLoopBody(stmt)
}

func collectionRangeBindings(parameters []Token, value Value, key Value, collection Value, position Value) map[string]Binding {
	bindings := make(map[string]Binding)

	switch len(parameters) {
	case 1:
		addRangeBinding(bindings, parameters[0], value)

	case 2:
		addRangeBinding(bindings, parameters[0], value)
		addRangeBinding(bindings, parameters[1], key)

	case 3:
		addRangeBinding(bindings, parameters[0], value)
		addRangeBinding(bindings, parameters[1], key)
		addRangeBinding(bindings, parameters[2], collection)

	case 4:
		addRangeBinding(bindings, parameters[0], value)
		addRangeBinding(bindings, parameters[1], key)
		addRangeBinding(bindings, parameters[2], collection)
		addRangeBinding(bindings, parameters[3], position)
	}

	return bindings
}

func addRangeBinding(bindings map[string]Binding, parameter Token, value Value) {
	if parameter.Literal == "_" {
		return
	}

	bindings[parameter.Literal] = Binding{
		Value:       value,
		IsImmutable: parameter.IsImmutable,
	}
}

func (i *Interpreter) evalAfterLoopBody(stmt *LoopStatement) (Value, error) {
	if stmt.AfterLoopBody != nil {
		return i.evalStatementBlock(stmt.AfterLoopBody)
	}

	return NewEmptyValue(), nil
}

func (i *Interpreter) evalStatementBlock(statements []Statement) (Value, error) {
	return i.evalStatementBlockWithBindings(statements, nil)
}

func (i *Interpreter) evalStatementBlockWithBindings(statements []Statement, initialBindings map[string]Binding) (Value, error) {
	blockEnv := NewChildEnvironment(i.Env)

	for name, binding := range initialBindings {
		blockEnv.values[name] = binding
	}

	previousEnv := i.Env
	i.Env = blockEnv
	defer func() {
		i.Env = previousEnv
	}()

	result := NewEmptyValue()

	for _, stmt := range statements {
		value, err := i.evalStatement(stmt)
		if err != nil {
			return Value{}, err
		}

		result = value
	}

	return result, nil
}

func (i *Interpreter) evalExpression(expr Expression) (Value, error) {
	switch e := expr.(type) {
	case *EmptyLiteral:
		return NewEmptyValue(), nil

	case *EmptyCollectionLiteral:
		return NewEmptyCollectionValue(), nil

	case *NumberLiteral:
		return NewNumberValueFromString(e.Value)

	case *StringLiteral:
		return NewStringValue(e.Value), nil

	case *IdentifierExpression:
		var binding Binding
		var ok bool

		if e.IsOuter {
			binding, ok = i.Env.GetOuter(e.Name)
			if !ok {
				return Value{}, fmt.Errorf(
					"line %d, column %d: no outer binding named %q",
					e.Token.StartOfLine,
					e.Token.StartOfColumn,
					e.Name,
				)
			}
		} else {
			binding, ok = i.Env.Get(e.Name)
			if !ok {
				return Value{}, fmt.Errorf(
					"line %d, column %d: undefined identifier %q in current scope",
					e.Token.StartOfLine,
					e.Token.StartOfColumn,
					e.Name,
				)
			}
		}

		return binding.Value, nil

	case *PrefixExpression:
		value, err := i.evalExpression(e.Right)
		if err != nil {
			return Value{}, err
		}

		return evalPrefix(e.Operator, value)

	case *BinaryExpression:
		if e.Operator == "&" || e.Operator == "|" {
			return i.evalLogicalBinaryExpression(e)
		}

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

	case *ArrayLiteral:
		return i.evalArrayLiteral(e)

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

func (i *Interpreter) evalLogicalBinaryExpression(expr *BinaryExpression) (Value, error) {
	left, err := i.evalExpression(expr.Left)
	if err != nil {
		return Value{}, err
	}

	left = resolveSpecializedValue(left)

	if left.Kind != ValueBool {
		return Value{}, fmt.Errorf("operator %q requires bool operands", expr.Operator)
	}

	switch expr.Operator {
	case "&":
		if !left.Bool {
			return NewBoolValue(false), nil
		}

		right, err := i.evalExpression(expr.Right)
		if err != nil {
			return Value{}, err
		}

		right = resolveSpecializedValue(right)

		if right.Kind != ValueBool {
			return Value{}, fmt.Errorf("operator %q requires bool operands", expr.Operator)
		}

		return NewBoolValue(right.Bool), nil

	case "|":
		if left.Bool {
			return NewBoolValue(true), nil
		}

		right, err := i.evalExpression(expr.Right)
		if err != nil {
			return Value{}, err
		}

		right = resolveSpecializedValue(right)

		if right.Kind != ValueBool {
			return Value{}, fmt.Errorf("operator %q requires bool operands", expr.Operator)
		}

		return NewBoolValue(right.Bool), nil

	default:
		return Value{}, fmt.Errorf("unknown logical operator %q", expr.Operator)
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

func (i *Interpreter) evalArrayLiteral(lit *ArrayLiteral) (Value, error) {
	elements := make([]Value, 0, len(lit.Elements))

	for _, elementExpression := range lit.Elements {
		element, err := i.evalExpression(elementExpression)
		if err != nil {
			return Value{}, err
		}

		elements = append(elements, element)
	}

	return NewArrayValue(elements, false), nil
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

	object = resolveSpecializedValue(object)

	switch object.Kind {
	case ValueMap:
		if index.Kind != ValueString {
			return Value{}, fmt.Errorf("map index must be a string")
		}

		key := index.Text.String()

		binding, ok := object.Map.Entries[key]
		if !ok {
			return Value{}, fmt.Errorf("map has no key %q", key)
		}

		return binding.Value, nil

	case ValueArray:
		arrayIndex, err := numberToArrayIndex(index)
		if err != nil {
			return Value{}, err
		}

		if arrayIndex < 0 || arrayIndex >= len(object.Array.Elements) {
			return Value{}, fmt.Errorf("array index %d out of bounds", arrayIndex)
		}

		return object.Array.Elements[arrayIndex], nil

	case ValueString:
		stringIndex, err := numberToArrayIndex(index)
		if err != nil {
			return Value{}, err
		}

		if object.Text == nil {
			return Value{}, fmt.Errorf("invalid string")
		}

		if stringIndex < 0 || stringIndex >= len(object.Text.Runes) {
			return Value{}, fmt.Errorf("string index %d out of bounds", stringIndex)
		}

		return NewStringValue(string(object.Text.Runes[stringIndex])), nil

	case ValueEmptyCollection:
		return Value{}, fmt.Errorf("cannot read from empty collection before it becomes a map or array")

	default:
		return Value{}, fmt.Errorf("cannot index non-collection value")
	}
}

func (i *Interpreter) evalIndexAssignmentStatement(stmt *IndexAssignmentStatement) (Value, error) {
	value, err := i.evalExpression(stmt.Value)
	if err != nil {
		return Value{}, err
	}

	return i.assignIndexExpression(stmt.Target, value)
}

func (i *Interpreter) evalCompoundAssignmentStatement(stmt *CompoundAssignmentStatement) (Value, error) {
	currentValue, err := i.evalExpression(stmt.Target)
	if err != nil {
		return Value{}, err
	}

	rightValue, err := i.evalExpression(stmt.Value)
	if err != nil {
		return Value{}, err
	}

	operator, err := compoundAssignmentBinaryOperator(stmt.Operator)
	if err != nil {
		return Value{}, err
	}

	newValue, err := evalBinary(operator, currentValue, rightValue)
	if err != nil {
		return Value{}, err
	}

	return i.assignExpressionTarget(stmt.Target, newValue)
}

func compoundAssignmentBinaryOperator(operator Token) (string, error) {
	switch operator.Type {
	case TokenPlusAssign:
		return "+", nil

	case TokenMinusAssign:
		return "-", nil

	default:
		return "", fmt.Errorf("unknown compound assignment operator %q", operator.Literal)
	}
}

func (i *Interpreter) assignExpressionTarget(target Expression, value Value) (Value, error) {
	switch t := target.(type) {
	case *IdentifierExpression:
		if t.IsOuter {
			if err := i.Env.SetOuter(t.Name, value); err != nil {
				return Value{}, fmt.Errorf(
					"line %d, column %d: %w",
					t.Token.StartOfLine,
					t.Token.StartOfColumn,
					err,
				)
			}
		} else {
			if err := i.Env.Set(t.Name, value, t.IsImmutable); err != nil {
				return Value{}, fmt.Errorf(
					"line %d, column %d: %w",
					t.Token.StartOfLine,
					t.Token.StartOfColumn,
					err,
				)
			}
		}

		return value, nil

	case *IndexExpression:
		return i.assignIndexExpression(t, value)

	default:
		return Value{}, fmt.Errorf("invalid assignment target")
	}
}

func (i *Interpreter) assignIndexExpression(target *IndexExpression, value Value) (Value, error) {
	object, err := i.evalExpression(target.Object)
	if err != nil {
		return Value{}, err
	}

	index, err := i.evalExpression(target.Index)
	if err != nil {
		return Value{}, err
	}

	object = resolveSpecializedValue(object)

	switch object.Kind {
	case ValueMap:
		return assignMapIndex(object.Map, index, value)

	case ValueArray:
		return assignArrayIndex(object.Array, index, value)

	case ValueString:
		return assignStringIndex(object.Text, index, value)

	case ValueEmptyCollection:
		return specializeEmptyCollection(object, index, value)

	default:
		return Value{}, fmt.Errorf("cannot assign into non-collection value")
	}
}

func assignMapIndex(m *Map, index Value, value Value) (Value, error) {
	if m.IsImmutable {
		return Value{}, fmt.Errorf("cannot modify immutable map")
	}

	if index.Kind != ValueString {
		return Value{}, fmt.Errorf("map index must be a string")
	}

	key := index.Text.String()

	binding, exists := m.Entries[key]
	if exists && binding.IsImmutable {
		return Value{}, fmt.Errorf("cannot reassign immutable map entry %q", key)
	}

	if exists {
		binding.Value = value
		m.Entries[key] = binding
		return value, nil
	}

	m.Entries[key] = Binding{
		Value:       value,
		IsImmutable: isImmutableIdentifier(key),
	}

	return value, nil
}

func assignArrayIndex(a *Array, index Value, value Value) (Value, error) {
	if a.IsImmutable {
		return Value{}, fmt.Errorf("cannot modify immutable array")
	}

	arrayIndex, err := numberToArrayIndex(index)
	if err != nil {
		return Value{}, err
	}

	if arrayIndex < 0 {
		return Value{}, fmt.Errorf("array index cannot be negative")
	}

	if arrayIndex > len(a.Elements) {
		return Value{}, fmt.Errorf(
			"array index %d is past the next append position %d",
			arrayIndex,
			len(a.Elements),
		)
	}

	if arrayIndex == len(a.Elements) {
		a.Elements = append(a.Elements, value)
		return value, nil
	}

	a.Elements[arrayIndex] = value
	return value, nil
}

func assignStringIndex(s *String, index Value, value Value) (Value, error) {
	if s == nil {
		return Value{}, fmt.Errorf("invalid string")
	}

	if s.IsImmutable {
		return Value{}, fmt.Errorf("cannot modify immutable string")
	}

	stringIndex, err := numberToArrayIndex(index)
	if err != nil {
		return Value{}, err
	}

	if stringIndex < 0 {
		return Value{}, fmt.Errorf("string index cannot be negative")
	}

	if stringIndex > len(s.Runes) {
		return Value{}, fmt.Errorf(
			"string index %d is past the next append position %d",
			stringIndex,
			len(s.Runes),
		)
	}

	replacement, err := stringAssignmentRune(value)
	if err != nil {
		return Value{}, err
	}

	if stringIndex == len(s.Runes) {
		s.Runes = append(s.Runes, replacement)
		return value, nil
	}

	s.Runes[stringIndex] = replacement
	return value, nil
}

func stringAssignmentRune(value Value) (rune, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueString {
		return 0, fmt.Errorf("string index assignment requires a string value")
	}

	if value.Text == nil {
		return 0, fmt.Errorf("string index assignment requires a valid string value")
	}

	if len(value.Text.Runes) != 1 {
		return 0, fmt.Errorf("string index assignment value must contain exactly one rune")
	}

	return value.Text.Runes[0], nil
}

func specializeEmptyCollection(collection Value, index Value, value Value) (Value, error) {
	if collection.EmptyCollection == nil {
		return Value{}, fmt.Errorf("invalid empty collection")
	}

	if collection.EmptyCollection.IsImmutable {
		return Value{}, fmt.Errorf("cannot modify immutable empty collection")
	}

	if collection.EmptyCollection.Specialized != nil {
		specialized := resolveSpecializedValue(*collection.EmptyCollection.Specialized)

		switch specialized.Kind {
		case ValueMap:
			return assignMapIndex(specialized.Map, index, value)

		case ValueArray:
			return assignArrayIndex(specialized.Array, index, value)

		default:
			return Value{}, fmt.Errorf("invalid specialized empty collection")
		}
	}

	switch index.Kind {
	case ValueString:
		specialized := NewMapValue(make(map[string]Binding), false)
		collection.EmptyCollection.Specialized = &specialized
		return assignMapIndex(specialized.Map, index, value)

	case ValueNumber:
		specialized := NewArrayValue([]Value{}, false)
		collection.EmptyCollection.Specialized = &specialized
		return assignArrayIndex(specialized.Array, index, value)

	default:
		return Value{}, fmt.Errorf("empty collection can only become a map or array through string or numeric indexing")
	}
}

func numberToArrayIndex(index Value) (int, error) {
	if index.Kind != ValueNumber {
		return 0, fmt.Errorf("array index must be a number")
	}

	i, accuracy := index.Number.Int64()
	if accuracy != big.Exact {
		return 0, fmt.Errorf("array index must be an integer")
	}

	return int(i), nil
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
		if parameter.Literal == "_" {
			continue
		}

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
			if signal, ok := asReturnSignal(err); ok {
				return signal.Value, nil
			}

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

func evalBinary(operator string, left Value, right Value) (Value, error) {
	left = resolveSpecializedValue(left)
	right = resolveSpecializedValue(right)

	if operator == "=" || operator == "!" {
		equal, err := valuesEqual(left, right)
		if err != nil {
			return Value{}, err
		}

		if operator == "!" {
			equal = !equal
		}

		return NewBoolValue(equal), nil
	}

	if operator == "+" && left.Kind == ValueString && right.Kind == ValueString {
		return NewStringValue(left.Text.String() + right.Text.String()), nil
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

func valuesEqual(left Value, right Value) (bool, error) {
	left = resolveSpecializedValue(left)
	right = resolveSpecializedValue(right)

	if left.Kind != right.Kind {
		return false, nil
	}

	switch left.Kind {
	case ValueEmpty:
		return true, nil

	case ValueNumber:
		return left.Number.Cmp(right.Number) == 0, nil

	case ValueBool:
		return left.Bool == right.Bool, nil

	case ValueString:
		return left.Text.String() == right.Text.String(), nil

	case ValueMap:
		return false, fmt.Errorf("cannot compare maps")

	case ValueArray:
		return false, fmt.Errorf("cannot compare arrays")

	case ValueEmptyCollection:
		return false, fmt.Errorf("cannot compare empty collections")

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
