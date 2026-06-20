package main

import "fmt"

func (i *Interpreter) evalExpression(expr Expression) (Value, error) {
	switch e := expr.(type) {
	case *VoidLiteral:
		return NewVoidValue(), nil

	case *BoolLiteral:
		return NewBoolValue(e.Value), nil

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
					"line %d, column %d: undefined identifier %q",
					e.Token.StartOfLine,
					e.Token.StartOfColumn,
					e.Name,
				)
			}
		}

		return binding.Value, nil

	case *LeadingDotMapExpression:
		return i.evalLeadingDotMapExpression(e)

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

	case *ConditionalExpression:
		return i.evalConditionalExpression(e)

	case *MatchExpression:
		return i.evalMatchExpression(e)

	case *MapLiteral:
		return i.evalMapLiteral(e)

	case *ArrayLiteral:
		return i.evalArrayLiteral(e)

	case *IndexExpression:
		return i.evalIndexExpression(e)

	case *FunctionLiteral:
		return NewFunctionValue(&Function{
			Parameters:        e.Parameters,
			VariadicParameter: e.VariadicParameter,
			Body:              e.Body,
			Returns:           e.Returns,
			Env:               i.Env,
			DotMap:            i.currentDotMap(),
		}), nil

	case *CallExpression:
		return i.evalCallExpression(e)

	default:
		return Value{}, fmt.Errorf("unknown expression type %T", expr)
	}
}

func (i *Interpreter) evalConditionalExpression(expr *ConditionalExpression) (Value, error) {
	condition, err := i.evalExpression(expr.Condition)
	if err != nil {
		return Value{}, err
	}

	condition = resolveSpecializedValue(condition)

	if condition.Kind != ValueBool {
		return Value{}, fmt.Errorf("conditional expression condition must be bool")
	}

	if condition.Bool {
		return i.evalExpression(expr.WhenTrue)
	}

	return i.evalExpression(expr.WhenFalse)
}

func (i *Interpreter) evalMatchExpression(expr *MatchExpression) (Value, error) {
	target, err := i.evalExpression(expr.Target)
	if err != nil {
		return Value{}, err
	}

	for _, arm := range expr.Arms {
		pattern, err := matchPatternValue(arm.Pattern)
		if err != nil {
			return Value{}, err
		}

		equal, err := evalBinary("=", target, pattern)
		if err != nil {
			return Value{}, err
		}

		equal = resolveSpecializedValue(equal)
		if equal.Kind != ValueBool {
			return Value{}, fmt.Errorf("match expression equality did not return bool")
		}

		if equal.Bool {
			return i.evalExpression(arm.Value)
		}
	}

	if expr.Default != nil {
		return i.evalExpression(expr.Default)
	}

	return NewVoidValue(), nil
}

func matchPatternValue(pattern MatchPattern) (Value, error) {
	switch pattern.Kind {
	case MatchPatternString:
		return NewStringValue(pattern.Token.Literal), nil

	case MatchPatternNumber:
		return NewNumberValueFromString(pattern.Token.Literal)

	case MatchPatternBool:
		return NewBoolValue(pattern.Token.Literal == "\\/"), nil

	default:
		return Value{}, fmt.Errorf("unknown match pattern kind %d", pattern.Kind)
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

func (i *Interpreter) evalLeadingDotMapExpression(expr *LeadingDotMapExpression) (Value, error) {
	currentMap := i.currentDotMap()
	if currentMap == nil {
		return Value{}, fmt.Errorf(
			"line %d, column %d: leading dot access has no surrounding map",
			expr.Token.StartOfLine,
			expr.Token.StartOfColumn,
		)
	}

	return Value{Kind: ValueMap, Map: currentMap}, nil
}

func (i *Interpreter) currentDotMap() *Map {
	if len(i.dotMaps) == 0 {
		return nil
	}

	return i.dotMaps[len(i.dotMaps)-1]
}

func (i *Interpreter) evalMapLiteral(lit *MapLiteral) (Value, error) {
	value := NewMapValue(map[string]Binding{}, false)
	if value.Map == nil {
		return Value{}, fmt.Errorf("invalid map")
	}

	i.dotMaps = append(i.dotMaps, value.Map)
	defer func() {
		i.dotMaps = i.dotMaps[:len(i.dotMaps)-1]
	}()

	for _, entry := range lit.Entries {
		key := entry.Key.Literal

		if _, exists := value.Map.Entries[key]; exists {
			return Value{}, fmt.Errorf(
				"line %d, column %d: duplicate map key %q",
				entry.Key.StartOfLine,
				entry.Key.StartOfColumn,
				key,
			)
		}

		entryValue, err := i.evalExpression(entry.Value)
		if err != nil {
			return Value{}, err
		}

		value.Map.Entries[key] = Binding{
			Value:       entryValue,
			IsImmutable: isImmutableIdentifier(key),
		}
	}

	return value, nil
}

func (i *Interpreter) evalArrayLiteral(lit *ArrayLiteral) (Value, error) {
	elements := make([]Value, 0, len(lit.Elements))

	for _, arrayElement := range lit.Elements {
		values, err := i.evalArrayElement(arrayElement)
		if err != nil {
			return Value{}, err
		}

		elements = append(elements, values...)
	}

	return NewArrayValue(elements, false), nil
}

func (i *Interpreter) evalArrayElement(element ArrayElement) ([]Value, error) {
	switch e := element.(type) {
	case *ExpressionArrayElement:
		value, err := i.evalExpression(e.Expression)
		if err != nil {
			return nil, err
		}

		return []Value{value}, nil

	case *RangeArrayElement:
		return i.evalRangeArrayElement(e)

	default:
		return nil, fmt.Errorf("unknown array element type %T", element)
	}
}

func (i *Interpreter) evalRangeArrayElement(element *RangeArrayElement) ([]Value, error) {
	startValue, err := i.evalExpression(element.Start)
	if err != nil {
		return nil, err
	}

	endValue, err := i.evalExpression(element.End)
	if err != nil {
		return nil, err
	}

	stepValue := NewVoidValue()
	if element.Step != nil {
		stepValue, err = i.evalExpression(element.Step)
		if err != nil {
			return nil, err
		}
	}

	return stultRangeValues(startValue, endValue, stepValue, element.IsInclusive)
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
		index = resolveSpecializedValue(index)
		if index.Kind != ValueNumber {
			return Value{}, fmt.Errorf("array index must be a number")
		}

		value, ok, err := object.Array.Get(index.Number)
		if err != nil {
			return Value{}, err
		}

		if !ok {
			return Value{}, fmt.Errorf("array index %s out of bounds", formatArrayIndex(index.Number))
		}

		return value, nil

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

	default:
		return Value{}, fmt.Errorf("cannot index non-collection value")
	}
}
