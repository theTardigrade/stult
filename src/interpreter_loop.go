package main

import "fmt"

func (i *Interpreter) evalLoopStatement(stmt *LoopStatement) (Value, error) {
	if len(stmt.RangeParameters) > 0 {
		return i.evalCollectionRangeLoopStatement(stmt)
	}

	loopValue, err := i.evalExpression(stmt.Condition)
	if err != nil {
		return Value{}, err
	}

	loopValue = resolveSpecializedValue(loopValue)

	switch loopValue.Kind {
	case ValueBool:
		return i.evalWhileLoopStatementWithInitialCondition(stmt, loopValue)

	case ValueMap:
		return i.evalMapRangeLoopStatement(stmt, loopValue.Map)

	case ValueArray:
		return i.evalArrayRangeLoopStatement(stmt, loopValue.Array)

	case ValueString:
		return i.evalStringRangeLoopStatement(stmt, loopValue.Text)

	default:
		return Value{}, fmt.Errorf("loop expression must evaluate to a bool, map, array, or string")
	}
}

func (i *Interpreter) evalWhileLoopStatement(stmt *LoopStatement) (Value, error) {
	condition, err := i.evalExpression(stmt.Condition)
	if err != nil {
		return Value{}, err
	}

	return i.evalWhileLoopStatementWithInitialCondition(stmt, resolveSpecializedValue(condition))
}

func (i *Interpreter) evalWhileLoopStatementWithInitialCondition(stmt *LoopStatement, firstCondition Value) (Value, error) {
	condition := firstCondition
	hasCondition := true

	for {
		if !hasCondition {
			value, err := i.evalExpression(stmt.Condition)
			if err != nil {
				return Value{}, err
			}

			condition = resolveSpecializedValue(value)
		}

		hasCondition = false

		if condition.Kind != ValueBool {
			return Value{}, fmt.Errorf("loop condition must evaluate to a bool")
		}

		if !condition.Bool {
			break
		}

		if _, err := i.evalStatementBlock(stmt.Body); err != nil {
			if flow, ok := asControlFlow(err); ok {
				switch flow.Kind {
				case controlFlowBreak:
					return i.evalAfterLoopBody(stmt)
				case controlFlowReturn:
					return Value{}, flow
				}
			}

			return Value{}, err
		}
	}

	return i.evalAfterLoopBody(stmt)
}

func (i *Interpreter) evalCollectionRangeLoopStatement(stmt *LoopStatement) (Value, error) {
	if !isValidCollectionRangeParameterCount(len(stmt.RangeParameters)) {
		return Value{}, fmt.Errorf("collection range loop must have zero, one, two, three, or four parameters")
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

	default:
		return Value{}, fmt.Errorf("collection range loop expression must evaluate to a map, array, or string")
	}
}

func isValidCollectionRangeParameterCount(count int) bool {
	return count >= 0 && count <= 4
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
			if flow, ok := asControlFlow(err); ok {
				switch flow.Kind {
				case controlFlowBreak:
					return i.evalAfterLoopBody(stmt)
				case controlFlowReturn:
					return Value{}, flow
				}
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
			if flow, ok := asControlFlow(err); ok {
				switch flow.Kind {
				case controlFlowBreak:
					return i.evalAfterLoopBody(stmt)
				case controlFlowReturn:
					return Value{}, flow
				}
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
			if flow, ok := asControlFlow(err); ok {
				switch flow.Kind {
				case controlFlowBreak:
					return i.evalAfterLoopBody(stmt)
				case controlFlowReturn:
					return Value{}, flow
				}
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

	return NewVoidValue(), nil
}
