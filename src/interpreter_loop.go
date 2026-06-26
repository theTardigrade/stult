package main

import "fmt"

func (i *Interpreter) evalLoopStatement(stmt *LoopStatement) (Value, error) {
	if directRange, ok := directRangeLoopElement(stmt.Condition); ok && len(stmt.RangeParameters) <= 2 {
		return i.evalDirectRangeLoopStatement(stmt, directRange)
	}

	loopValue, err := i.evalExpression(stmt.Condition)
	if err != nil {
		return Value{}, err
	}

	loopValue = resolveSpecializedValue(loopValue)

	switch loopValue.Kind {
	case ValueBool:
		if len(stmt.RangeParameters) > 0 {
			return Value{}, fmt.Errorf(
				"loop with parameters must evaluate to a map, array, string, or function",
			)
		}

		return i.evalWhileLoopStatementWithInitialCondition(stmt, loopValue)

	case ValueMap:
		return i.evalMapRangeLoopStatement(stmt, loopValue.Map)

	case ValueArray:
		return i.evalArrayRangeLoopStatement(stmt, loopValue.Array)

	case ValueString:
		return i.evalStringRangeLoopStatement(stmt, loopValue.Text)

	case ValueFunction:
		return i.evalFunctionRangeLoopStatement(stmt, loopValue.Function)

	default:
		return Value{}, fmt.Errorf(
			"loop expression must evaluate to a bool, map, array, string, or function",
		)
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

func isValidCollectionRangeParameterCount(count int) bool {
	return count >= 0 && count <= 4
}

func isValidFunctionRangeParameterCount(count int) bool {
	return count >= 0 && count <= 2
}

func (i *Interpreter) evalMapRangeLoopStatement(stmt *LoopStatement, m *Map) (Value, error) {
	if !isValidCollectionRangeParameterCount(len(stmt.RangeParameters)) {
		return Value{}, fmt.Errorf("collection range loop must have zero, one, two, three, or four parameters")
	}

	position := 0
	lastKey := ""
	hasLastKey := false

	for {
		key, ok := nextMapRangeKey(m, lastKey, hasLastKey)
		if !ok {
			break
		}

		binding, ok := m.Get(key)
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
	keys := m.Keys()

	for _, key := range keys {
		if !hasLastKey || key > lastKey {
			return key, true
		}
	}

	return "", false
}

func (i *Interpreter) evalArrayRangeLoopStatement(stmt *LoopStatement, a *Array) (Value, error) {
	if !isValidCollectionRangeParameterCount(len(stmt.RangeParameters)) {
		return Value{}, fmt.Errorf("collection range loop must have zero, one, two, three, or four parameters")
	}

	if err := a.ForEach(func(index *Number, value Value) error {
		key := NewNumberValueFromNumber(index)

		loopBindings := collectionRangeBindings(
			stmt.RangeParameters,
			value,
			key,
			Value{Kind: ValueArray, Array: a},
			key,
		)

		_, err := i.evalStatementBlockWithBindings(stmt.Body, loopBindings)
		return err
	}); err != nil {
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

	return i.evalAfterLoopBody(stmt)
}

func (i *Interpreter) evalStringRangeLoopStatement(stmt *LoopStatement, s *String) (Value, error) {
	if !isValidCollectionRangeParameterCount(len(stmt.RangeParameters)) {
		return Value{}, fmt.Errorf("collection range loop must have zero, one, two, three, or four parameters")
	}

	if s == nil {
		return Value{}, fmt.Errorf("cannot range over invalid string")
	}

	if err := s.ForEach(func(index int, r rune) error {
		key := NewNumberValueFromInt(index)

		loopBindings := collectionRangeBindings(
			stmt.RangeParameters,
			NewStringValue(string(r)),
			key,
			Value{Kind: ValueString, Text: s},
			key,
		)

		if _, err := i.evalStatementBlockWithBindings(stmt.Body, loopBindings); err != nil {
			return err
		}

		return nil
	}); err != nil {
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

	return i.evalAfterLoopBody(stmt)
}

func directRangeLoopElement(expression Expression) (*RangeArrayElement, bool) {
	arrayLiteral, ok := expression.(*ArrayLiteral)
	if !ok || len(arrayLiteral.Elements) != 1 {
		return nil, false
	}

	rangeElement, ok := arrayLiteral.Elements[0].(*RangeArrayElement)
	return rangeElement, ok
}

func (i *Interpreter) evalDirectRangeLoopStatement(
	stmt *LoopStatement,
	rangeElement *RangeArrayElement,
) (Value, error) {
	if len(stmt.RangeParameters) > 2 {
		return Value{}, fmt.Errorf("direct range loop must have zero, one, or two parameters")
	}

	startValue, err := i.evalExpression(rangeElement.Start)
	if err != nil {
		return Value{}, err
	}

	endValue, err := i.evalExpression(rangeElement.End)
	if err != nil {
		return Value{}, err
	}

	stepValue := NewVoidValue()
	if rangeElement.Step != nil {
		stepValue, err = i.evalExpression(rangeElement.Step)
		if err != nil {
			return Value{}, err
		}
	}

	iterator, err := newStultRangeIterator(startValue, endValue, stepValue, rangeElement.IsInclusive)
	if err != nil {
		return Value{}, err
	}

	position := 0
	for {
		value, ok := iterator.nextValue()
		if !ok {
			break
		}

		positionValue := NewNumberValueFromInt(position)
		loopBindings := collectionRangeBindings(
			stmt.RangeParameters,
			value,
			positionValue,
			NewVoidValue(),
			positionValue,
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

		position++
	}

	return i.evalAfterLoopBody(stmt)
}

func (i *Interpreter) evalFunctionRangeLoopStatement(stmt *LoopStatement, fn *Function) (Value, error) {
	if !isValidFunctionRangeParameterCount(len(stmt.RangeParameters)) {
		return Value{}, fmt.Errorf("function range loop must have zero, one, or two parameters")
	}

	position := 0

	for {
		args, err := functionRangeLoopArguments(fn, position)
		if err != nil {
			return Value{}, err
		}

		value, err := i.callFunction(fn, args)
		if err != nil {
			return Value{}, err
		}

		value = resolveSpecializedValue(value)
		if value.Kind == ValueVoid {
			break
		}

		loopBindings := functionRangeBindings(
			stmt.RangeParameters,
			value,
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

		position++
	}

	return i.evalAfterLoopBody(stmt)
}

func functionRangeLoopArguments(fn *Function, position int) ([]Value, error) {
	if functionCanAcceptArgumentCount(fn, 1) {
		return []Value{NewNumberValueFromInt(position)}, nil
	}

	if functionCanAcceptArgumentCount(fn, 0) {
		return []Value{}, nil
	}

	return nil, fmt.Errorf("function loop source must accept zero or one argument")
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

func functionRangeBindings(parameters []Token, value Value, position Value) map[string]Binding {
	bindings := make(map[string]Binding)

	switch len(parameters) {
	case 1:
		addRangeBinding(bindings, parameters[0], value)

	case 2:
		addRangeBinding(bindings, parameters[0], value)
		addRangeBinding(bindings, parameters[1], position)
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
