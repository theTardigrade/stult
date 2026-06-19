package main

import "fmt"

type bytecodeVMIterator struct {
	Source              Value
	ParameterCount      int
	Position            int
	CurrentKey          Value
	CurrentValue        Value
	HasCurrent          bool
	LastMapKey          string
	HasLastMapKey       bool
	DirectRangeIterator *stultRangeIterator
}

func (vm *BytecodeVM) iteratorInit(parameterCount int) error {
	source, err := vm.popValue()
	if err != nil {
		return err
	}

	iterator, err := newBytecodeVMIterator(source, parameterCount)
	if err != nil {
		return err
	}

	vm.iterators = append(vm.iterators, iterator)

	return nil
}

func (vm *BytecodeVM) iteratorRangeInit(operand int) error {
	parameterCount, isInclusive := decodeIteratorRangeInitOperand(operand)

	step, err := vm.popValue()
	if err != nil {
		return err
	}

	end, err := vm.popValue()
	if err != nil {
		return err
	}

	start, err := vm.popValue()
	if err != nil {
		return err
	}

	iterator, err := newBytecodeVMDirectRangeIterator(start, end, step, isInclusive, parameterCount)
	if err != nil {
		return err
	}

	vm.iterators = append(vm.iterators, iterator)

	return nil
}

func newBytecodeVMDirectRangeIterator(
	startValue Value,
	endValue Value,
	stepValue Value,
	isInclusive bool,
	parameterCount int,
) (bytecodeVMIterator, error) {
	iterator := bytecodeVMIterator{
		Source:         NewVoidValue(),
		ParameterCount: parameterCount,
		Position:       -1,
		CurrentKey:     NewVoidValue(),
		CurrentValue:   NewVoidValue(),
		HasCurrent:     false,
	}

	if !isValidFunctionRangeParameterCount(parameterCount) {
		return iterator, fmt.Errorf("direct range loop must have zero, one, or two parameters")
	}

	rangeIterator, err := newStultRangeIterator(startValue, endValue, stepValue, isInclusive)
	if err != nil {
		return iterator, err
	}

	iterator.DirectRangeIterator = rangeIterator

	return iterator, nil
}

func newBytecodeVMIterator(source Value, parameterCount int) (bytecodeVMIterator, error) {
	source = resolveSpecializedValue(source)

	iterator := bytecodeVMIterator{
		Source:         source,
		ParameterCount: parameterCount,
		Position:       -1,
		CurrentKey:     NewVoidValue(),
		CurrentValue:   NewVoidValue(),
		HasCurrent:     false,
		LastMapKey:     "",
		HasLastMapKey:  false,
	}

	switch source.Kind {
	case ValueArray:
		if source.Array == nil {
			return iterator, fmt.Errorf("invalid array")
		}

		if !isValidCollectionRangeParameterCount(parameterCount) {
			return iterator, fmt.Errorf("collection range loop must have zero, one, two, three, or four parameters")
		}

	case ValueMap:
		if source.Map == nil {
			return iterator, fmt.Errorf("invalid map")
		}

		if !isValidCollectionRangeParameterCount(parameterCount) {
			return iterator, fmt.Errorf("collection range loop must have zero, one, two, three, or four parameters")
		}

	case ValueString:
		if source.Text == nil {
			return iterator, fmt.Errorf("invalid string")
		}

		if !isValidCollectionRangeParameterCount(parameterCount) {
			return iterator, fmt.Errorf("collection range loop must have zero, one, two, three, or four parameters")
		}

	case ValueFunction:
		if source.Function == nil {
			return iterator, fmt.Errorf("invalid function")
		}

		if !isValidFunctionRangeParameterCount(parameterCount) {
			return iterator, fmt.Errorf("function range loop must have zero, one, or two parameters")
		}

	default:
		return iterator, fmt.Errorf("loop expression must evaluate to a map, array, string, or function")
	}

	return iterator, nil
}

func (vm *BytecodeVM) iteratorNext(target int) error {
	iterator, err := vm.currentIterator()
	if err != nil {
		return err
	}

	if iterator.DirectRangeIterator != nil {
		return vm.iteratorNextDirectRange(iterator, target)
	}

	switch iterator.Source.Kind {
	case ValueArray:
		return vm.iteratorNextArray(iterator, target)

	case ValueMap:
		return vm.iteratorNextMap(iterator, target)

	case ValueString:
		return vm.iteratorNextString(iterator, target)

	case ValueFunction:
		return vm.iteratorNextFunction(iterator, target)

	default:
		return fmt.Errorf("loop expression must evaluate to a map, array, string, or function")
	}
}

func (vm *BytecodeVM) iteratorNextDirectRange(iterator *bytecodeVMIterator, target int) error {
	value, ok := iterator.DirectRangeIterator.nextValue()
	if !ok {
		iterator.HasCurrent = false
		return vm.jump(target)
	}

	position := iterator.Position + 1
	positionValue := NewNumberValueFromInt(position)

	iterator.Position = position
	iterator.CurrentKey = positionValue
	iterator.CurrentValue = value
	iterator.HasCurrent = true

	return nil
}

func (vm *BytecodeVM) iteratorNextArray(iterator *bytecodeVMIterator, target int) error {
	array := iterator.Source.Array
	if array == nil {
		return fmt.Errorf("invalid array")
	}

	position := iterator.Position + 1
	key := NewNumberValueFromInt(position)
	value, ok, err := array.Get(key.Number)
	if err != nil {
		return err
	}

	if !ok {
		iterator.HasCurrent = false
		return vm.jump(target)
	}

	iterator.Position = position
	iterator.CurrentKey = key
	iterator.CurrentValue = value
	iterator.HasCurrent = true

	return nil
}

func (vm *BytecodeVM) iteratorNextMap(iterator *bytecodeVMIterator, target int) error {
	m := iterator.Source.Map
	if m == nil {
		return fmt.Errorf("invalid map")
	}

	for {
		key, ok := nextMapRangeKey(m, iterator.LastMapKey, iterator.HasLastMapKey)
		if !ok {
			iterator.HasCurrent = false
			return vm.jump(target)
		}

		binding, ok := m.Entries[key]

		iterator.LastMapKey = key
		iterator.HasLastMapKey = true

		if !ok {
			continue
		}

		iterator.Position++
		iterator.CurrentKey = NewStringValue(key)
		iterator.CurrentValue = binding.Value
		iterator.HasCurrent = true

		return nil
	}
}

func (vm *BytecodeVM) iteratorNextString(iterator *bytecodeVMIterator, target int) error {
	text := iterator.Source.Text
	if text == nil {
		return fmt.Errorf("invalid string")
	}

	position := iterator.Position + 1
	if position >= len(text.Runes) {
		iterator.HasCurrent = false
		return vm.jump(target)
	}

	key := NewNumberValueFromInt(position)

	iterator.Position = position
	iterator.CurrentKey = key
	iterator.CurrentValue = NewStringValue(string(text.Runes[position]))
	iterator.HasCurrent = true

	return nil
}

func (vm *BytecodeVM) iteratorNextFunction(iterator *bytecodeVMIterator, target int) error {
	fn := iterator.Source.Function
	if fn == nil {
		return fmt.Errorf("invalid function")
	}

	position := iterator.Position + 1
	args, err := functionRangeLoopArguments(fn, position)
	if err != nil {
		return err
	}

	value, err := vm.callFunction(fn, args)
	if err != nil {
		return err
	}

	value = resolveSpecializedValue(value)
	if value.Kind == ValueVoid {
		iterator.HasCurrent = false
		return vm.jump(target)
	}

	positionValue := NewNumberValueFromInt(position)

	iterator.Position = position
	iterator.CurrentKey = positionValue
	iterator.CurrentValue = value
	iterator.HasCurrent = true

	return nil
}

func (vm *BytecodeVM) storeIteratorValue(instructionIndex int, localIndex int) error {
	iterator, err := vm.currentIterator()
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	if !iterator.HasCurrent {
		return vm.runtimeError(instructionIndex, "iterator has no current value")
	}

	if err := vm.storeLocal(localIndex, iterator.CurrentValue, false, true); err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	return nil
}

func (vm *BytecodeVM) storeIteratorKey(instructionIndex int, localIndex int) error {
	iterator, err := vm.currentIterator()
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	if !iterator.HasCurrent {
		return vm.runtimeError(instructionIndex, "iterator has no current key")
	}

	if err := vm.storeLocal(localIndex, iterator.CurrentKey, false, true); err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	return nil
}

func (vm *BytecodeVM) storeIteratorCollection(instructionIndex int, localIndex int) error {
	iterator, err := vm.currentIterator()
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	if iterator.Source.Kind == ValueFunction || iterator.DirectRangeIterator != nil {
		return vm.runtimeError(
			instructionIndex,
			"loop source has no collection parameter",
		)
	}

	if err := vm.storeLocal(localIndex, iterator.Source, false, true); err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	return nil
}

func (vm *BytecodeVM) storeIteratorPosition(instructionIndex int, localIndex int) error {
	iterator, err := vm.currentIterator()
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	if !iterator.HasCurrent {
		return vm.runtimeError(instructionIndex, "iterator has no current position")
	}

	if err := vm.storeLocal(
		localIndex,
		NewNumberValueFromInt(iterator.Position),
		false,
		true,
	); err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	return nil
}

func (vm *BytecodeVM) currentIterator() (*bytecodeVMIterator, error) {
	if len(vm.iterators) == 0 {
		return nil, fmt.Errorf("no active iterator")
	}

	return &vm.iterators[len(vm.iterators)-1], nil
}
