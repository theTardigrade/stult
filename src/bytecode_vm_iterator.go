package main

import "fmt"

type bytecodeVMIterator struct {
	Collection    Value
	Position      int
	CurrentKey    Value
	CurrentValue  Value
	HasCurrent    bool
	LastMapKey    string
	HasLastMapKey bool
}

func (vm *BytecodeVM) iteratorInit() error {
	collection, err := vm.popValue()
	if err != nil {
		return err
	}

	iterator, err := newBytecodeVMIterator(collection)
	if err != nil {
		return err
	}

	vm.iterators = append(vm.iterators, iterator)

	return nil
}

func newBytecodeVMIterator(collection Value) (bytecodeVMIterator, error) {
	collection = resolveSpecializedValue(collection)

	iterator := bytecodeVMIterator{
		Collection:    collection,
		Position:      -1,
		CurrentKey:    NewVoidValue(),
		CurrentValue:  NewVoidValue(),
		HasCurrent:    false,
		LastMapKey:    "",
		HasLastMapKey: false,
	}

	switch collection.Kind {
	case ValueArray:
		if collection.Array == nil {
			return iterator, fmt.Errorf("invalid array")
		}

	case ValueMap:
		if collection.Map == nil {
			return iterator, fmt.Errorf("invalid map")
		}

	case ValueString:
		if collection.Text == nil {
			return iterator, fmt.Errorf("invalid string")
		}

	default:
		return iterator, fmt.Errorf("loop expression must evaluate to a map, array or string")
	}

	return iterator, nil
}

func (vm *BytecodeVM) iteratorNext(target int) error {
	iterator, err := vm.currentIterator()
	if err != nil {
		return err
	}

	switch iterator.Collection.Kind {
	case ValueArray:
		return vm.iteratorNextArray(iterator, target)

	case ValueMap:
		return vm.iteratorNextMap(iterator, target)

	case ValueString:
		return vm.iteratorNextString(iterator, target)

	default:
		return fmt.Errorf("loop expression must evaluate to a map, array or string")
	}
}

func (vm *BytecodeVM) iteratorNextArray(iterator *bytecodeVMIterator, target int) error {
	array := iterator.Collection.Array
	if array == nil {
		return fmt.Errorf("invalid array")
	}

	position := iterator.Position + 1
	if position >= len(array.Elements) {
		iterator.HasCurrent = false
		return vm.jump(target)
	}

	key := NewNumberValueFromInt(position)

	iterator.Position = position
	iterator.CurrentKey = key
	iterator.CurrentValue = array.Elements[position]
	iterator.HasCurrent = true

	return nil
}

func (vm *BytecodeVM) iteratorNextMap(iterator *bytecodeVMIterator, target int) error {
	m := iterator.Collection.Map
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
	text := iterator.Collection.Text
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

	if err := vm.storeLocal(localIndex, iterator.Collection, false, true); err != nil {
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
