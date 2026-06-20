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

	case TokenStarAssign:
		return "*", nil

	case TokenSlashAssign:
		return "/", nil

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
	if err := m.Set(index, value); err != nil {
		return Value{}, err
	}

	return value, nil
}

func assignArrayIndex(a *Array, index Value, value Value) (Value, error) {
	index = resolveSpecializedValue(index)
	if index.Kind != ValueNumber {
		return Value{}, fmt.Errorf("array index must be a number")
	}

	if err := a.Set(index.Number, value); err != nil {
		return Value{}, err
	}

	return value, nil
}

func assignStringIndex(s *String, index Value, value Value) (Value, error) {
	index = resolveSpecializedValue(index)
	if index.Kind != ValueNumber {
		return Value{}, fmt.Errorf("string index must be a number")
	}

	if err := s.Set(index.Number, value); err != nil {
		return Value{}, err
	}

	return value, nil
}
