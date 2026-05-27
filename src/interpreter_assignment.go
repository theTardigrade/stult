package main

import (
	"fmt"
)

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
