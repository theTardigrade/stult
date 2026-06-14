package main

import "fmt"

func (vm *BytecodeVM) storeGlobalFromStack(
	instructionIndex int,
	operand int,
	isImmutable bool,
) error {
	name, err := vm.constantName(operand)
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	value, err := vm.popValue()
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	if err := vm.storeGlobal(name, value, isImmutable); err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	return nil
}

func (vm *BytecodeVM) storeLocalFromStack(
	instructionIndex int,
	index int,
	isImmutable bool,
) error {
	value, err := vm.popValue()
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	if err := vm.storeLocal(index, value, isImmutable, false); err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	return nil
}

func (vm *BytecodeVM) storeUpvalueFromStack(
	instructionIndex int,
	index int,
	isImmutable bool,
) error {
	value, err := vm.popValue()
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	if err := vm.storeUpvalue(index, value, isImmutable, false); err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	return nil
}

func (vm *BytecodeVM) constantValue(index int) (Value, error) {
	if vm.chunk == nil {
		return Value{}, fmt.Errorf("no active bytecode chunk")
	}

	if index < 0 || index >= len(vm.chunk.Constants) {
		return Value{}, fmt.Errorf("constant index %d out of bounds", index)
	}

	return vm.chunk.Constants[index], nil
}

func (vm *BytecodeVM) constantName(index int) (string, error) {
	value, err := vm.constantValue(index)
	if err != nil {
		return "", err
	}

	value = resolveSpecializedValue(value)

	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("constant %d is not a string name", index)
	}

	return value.Text.String(), nil
}

func (vm *BytecodeVM) loadGlobal(name string) (Value, error) {
	binding, ok := vm.globals[name]
	if !ok {
		return Value{}, fmt.Errorf("undefined global %q", name)
	}

	return binding.Value, nil
}

func (vm *BytecodeVM) storeGlobal(name string, value Value, isImmutable bool) error {
	existing, exists := vm.globals[name]

	if exists && existing.IsImmutable {
		return fmt.Errorf("cannot reassign immutable global %q", name)
	}

	if exists {
		vm.globals[name] = Binding{
			Value:       value,
			IsImmutable: existing.IsImmutable,
		}

		return nil
	}

	vm.globals[name] = Binding{
		Value:       value,
		IsImmutable: isImmutable,
	}

	return nil
}

func (vm *BytecodeVM) loadLocal(index int) (Value, error) {
	cell, err := vm.localCell(index)
	if err != nil {
		return Value{}, err
	}

	if !cell.Initialized {
		return Value{}, fmt.Errorf("local %d is not initialized", index)
	}

	return cell.Value, nil
}

func (vm *BytecodeVM) storeLocal(
	index int,
	value Value,
	isImmutable bool,
	allowImmutableRebind bool,
) error {
	cell, err := vm.localCell(index)
	if err != nil {
		return err
	}

	if cell.Initialized && cell.IsImmutable && !allowImmutableRebind {
		return fmt.Errorf("cannot reassign immutable local %d", index)
	}

	cell.Value = value
	cell.Initialized = true
	cell.IsImmutable = cell.IsImmutable || isImmutable

	return nil
}

func (vm *BytecodeVM) localCell(index int) (*bytecodeVMCell, error) {
	if index < 0 || index >= len(vm.locals) {
		return nil, fmt.Errorf("local index %d out of bounds", index)
	}

	return &vm.locals[index], nil
}

func (vm *BytecodeVM) resetLocalsFromDepth(scopeDepth int) error {
	if vm.chunk == nil {
		return fmt.Errorf("no active bytecode chunk")
	}

	if scopeDepth < 0 {
		return fmt.Errorf("local scope depth cannot be negative")
	}

	if len(vm.locals) < len(vm.chunk.Locals) {
		return fmt.Errorf("local storage has not been initialized")
	}

	for _, index := range vm.resetLocalIndexesFromDepth(vm.chunk, scopeDepth) {
		local := vm.chunk.Locals[index]

		vm.locals[index] = bytecodeVMCell{
			Value:       NewVoidValue(),
			Initialized: false,
			IsImmutable: local.IsImmutable,
		}
	}

	return nil
}

func (vm *BytecodeVM) resetLocalIndexesFromDepth(
	chunk *BytecodeChunk,
	scopeDepth int,
) []int {
	if vm.resetLocalIndexesByDepth == nil {
		vm.resetLocalIndexesByDepth = map[*BytecodeChunk]map[int][]int{}
	}

	indexesByDepth, ok := vm.resetLocalIndexesByDepth[chunk]
	if !ok {
		indexesByDepth = map[int][]int{}
		vm.resetLocalIndexesByDepth[chunk] = indexesByDepth
	}

	indexes, ok := indexesByDepth[scopeDepth]
	if ok {
		return indexes
	}

	indexes = make([]int, 0)

	for index, local := range chunk.Locals {
		if local.ScopeDepth >= scopeDepth {
			indexes = append(indexes, index)
		}
	}

	indexesByDepth[scopeDepth] = indexes

	return indexes
}

func (vm *BytecodeVM) loadUpvalue(index int) (Value, error) {
	cell, err := vm.upvalueCell(index)
	if err != nil {
		return Value{}, err
	}

	if !cell.Initialized {
		return Value{}, fmt.Errorf("upvalue %d is not initialized", index)
	}

	return cell.Value, nil
}

func (vm *BytecodeVM) storeUpvalue(
	index int,
	value Value,
	isImmutable bool,
	allowImmutableRebind bool,
) error {
	cell, err := vm.upvalueCell(index)
	if err != nil {
		return err
	}

	if cell.Initialized && cell.IsImmutable && !allowImmutableRebind {
		return fmt.Errorf("cannot reassign immutable upvalue %d", index)
	}

	cell.Value = value
	cell.Initialized = true
	cell.IsImmutable = cell.IsImmutable || isImmutable

	return nil
}

func (vm *BytecodeVM) upvalueCell(index int) (*bytecodeVMCell, error) {
	if index < 0 || index >= len(vm.upvalues) {
		return nil, fmt.Errorf("upvalue index %d out of bounds", index)
	}

	return vm.upvalues[index], nil
}

func (vm *BytecodeVM) buildArray(count int) error {
	if count < 0 {
		return fmt.Errorf("array element count cannot be negative")
	}

	if len(vm.stack) < count {
		return fmt.Errorf("BUILD_ARRAY expected %d stack value(s)", count)
	}

	entries := make([]bytecodeVMStackEntry, count)

	for index := count - 1; index >= 0; index-- {
		entry, err := vm.popEntry()
		if err != nil {
			return err
		}

		entries[index] = entry
	}

	values := make([]Value, 0, count)

	for _, entry := range entries {
		if entry.IsRangeSegment {
			if entry.Value.Kind != ValueArray || entry.Value.Array == nil {
				return fmt.Errorf("range segment did not produce an array")
			}

			values = append(values, entry.Value.Array.Elements...)
			continue
		}

		values = append(values, entry.Value)
	}

	vm.pushValue(NewArrayValue(values, false))

	return nil
}

func (vm *BytecodeVM) buildMap(entryCount int) error {
	if entryCount < 0 {
		return fmt.Errorf("map entry count cannot be negative")
	}

	if len(vm.stack) < entryCount*2 {
		return fmt.Errorf("BUILD_MAP expected %d stack value(s)", entryCount*2)
	}

	entries := make(map[string]Binding, entryCount)

	for index := entryCount - 1; index >= 0; index-- {
		value, err := vm.popValue()
		if err != nil {
			return err
		}

		key, err := vm.popValue()
		if err != nil {
			return err
		}

		key = resolveSpecializedValue(key)

		if key.Kind != ValueString || key.Text == nil {
			return fmt.Errorf("map key must be a string")
		}

		keyText := key.Text.String()

		entries[keyText] = Binding{
			Value:       value,
			IsImmutable: isImmutableIdentifier(keyText),
		}
	}

	vm.pushValue(NewMapValue(entries, false))

	return nil
}

func (vm *BytecodeVM) buildRange(isInclusive bool) error {
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

	values, err := bytecodeRangeValues(start, end, step, isInclusive)
	if err != nil {
		return err
	}

	vm.pushRangeSegment(NewArrayValue(values, false))

	return nil
}

func bytecodeRangeValues(
	startValue Value,
	endValue Value,
	stepValue Value,
	isInclusive bool,
) ([]Value, error) {
	start, err := numberToInt64(startValue, "range start")
	if err != nil {
		return nil, err
	}

	end, err := numberToInt64(endValue, "range end")
	if err != nil {
		return nil, err
	}

	step := int64(0)
	stepValue = resolveSpecializedValue(stepValue)

	if stepValue.Kind == ValueVoid {
		if start <= end {
			step = 1
		} else {
			step = -1
		}
	} else {
		step, err = numberToInt64(stepValue, "range step")
		if err != nil {
			return nil, err
		}

		if step == 0 {
			return nil, fmt.Errorf("range step cannot be zero")
		}
	}

	values := []Value{}

	if step > 0 {
		limit := end
		if !isInclusive {
			limit = end - 1
		}

		for current := start; current <= limit; current += step {
			values = append(values, NewNumberValueFromInt64(current))
		}
	} else {
		limit := end
		if !isInclusive {
			limit = end + 1
		}

		for current := start; current >= limit; current += step {
			values = append(values, NewNumberValueFromInt64(current))
		}
	}

	return values, nil
}

func (vm *BytecodeVM) jump(target int) error {
	if vm.chunk == nil {
		return fmt.Errorf("no active bytecode chunk")
	}

	if target < 0 || target > len(vm.chunk.Instructions) {
		return fmt.Errorf("jump target %d out of bounds", target)
	}

	vm.ip = target

	return nil
}

func (vm *BytecodeVM) pushValue(value Value) {
	vm.stack = append(vm.stack, bytecodeVMStackEntry{
		Value:          value,
		IsRangeSegment: false,
	})
}

func (vm *BytecodeVM) pushRangeSegment(value Value) {
	vm.stack = append(vm.stack, bytecodeVMStackEntry{
		Value:          value,
		IsRangeSegment: true,
	})
}

func (vm *BytecodeVM) popEntry() (bytecodeVMStackEntry, error) {
	if len(vm.stack) == 0 {
		return bytecodeVMStackEntry{}, fmt.Errorf("stack underflow")
	}

	entry := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]

	return entry, nil
}

func (vm *BytecodeVM) popValue() (Value, error) {
	entry, err := vm.popEntry()
	if err != nil {
		return Value{}, err
	}

	if entry.IsRangeSegment {
		return Value{}, fmt.Errorf("range segment cannot be used as a plain value")
	}

	return entry.Value, nil
}

func (vm *BytecodeVM) peekValue() (Value, error) {
	if len(vm.stack) == 0 {
		return Value{}, fmt.Errorf("stack underflow")
	}

	entry := vm.stack[len(vm.stack)-1]

	if entry.IsRangeSegment {
		return Value{}, fmt.Errorf("range segment cannot be used as a plain value")
	}

	return entry.Value, nil
}

func bytecodeValueIsCollection(value Value) bool {
	value = resolveSpecializedValue(value)

	return value.Kind == ValueMap ||
		value.Kind == ValueArray ||
		value.Kind == ValueString
}

func bytecodeIndexValue(object Value, index Value) (Value, error) {
	object = resolveSpecializedValue(object)
	index = resolveSpecializedValue(index)

	switch object.Kind {
	case ValueMap:
		if object.Map == nil {
			return Value{}, fmt.Errorf("invalid map")
		}

		if index.Kind != ValueString || index.Text == nil {
			return Value{}, fmt.Errorf("map index must be a string")
		}

		key := index.Text.String()

		binding, ok := object.Map.Entries[key]
		if !ok {
			return Value{}, fmt.Errorf("map has no key %q", key)
		}

		return binding.Value, nil

	case ValueArray:
		if object.Array == nil {
			return Value{}, fmt.Errorf("invalid array")
		}

		arrayIndex, err := numberToArrayIndex(index)
		if err != nil {
			return Value{}, err
		}

		if arrayIndex < 0 || arrayIndex >= len(object.Array.Elements) {
			return Value{}, fmt.Errorf("array index %d out of bounds", arrayIndex)
		}

		return object.Array.Elements[arrayIndex], nil

	case ValueString:
		if object.Text == nil {
			return Value{}, fmt.Errorf("invalid string")
		}

		stringIndex, err := numberToArrayIndex(index)
		if err != nil {
			return Value{}, err
		}

		if stringIndex < 0 || stringIndex >= len(object.Text.Runes) {
			return Value{}, fmt.Errorf("string index %d out of bounds", stringIndex)
		}

		return NewStringValue(string(object.Text.Runes[stringIndex])), nil

	default:
		return Value{}, fmt.Errorf("cannot index non-collection value")
	}
}

func bytecodeAssignIndex(object Value, index Value, value Value) (Value, error) {
	object = resolveSpecializedValue(object)
	index = resolveSpecializedValue(index)

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

func bytecodeBoolValue(value Value) (bool, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueBool {
		return false, fmt.Errorf("condition must be a bool")
	}

	return value.Bool, nil
}

func bytecodeCloneValueForLoad(value Value) Value {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueNumber:
		if value.Number == nil {
			return value
		}

		return Value{
			Kind:   ValueNumber,
			Number: CloneNumber(value.Number),
		}

	case ValueString:
		if value.Text == nil {
			return value
		}

		return NewStringValueWithImmutability(value.Text.String(), value.Text.IsImmutable)

	default:
		return value
	}
}

func (vm *BytecodeVM) runtimeError(
	instructionIndex int,
	format string,
	args ...interface{},
) error {
	message := fmt.Sprintf(format, args...)
	sourceSpan := EmptyBytecodeSourceSpan()

	if vm.chunk != nil &&
		instructionIndex >= 0 &&
		instructionIndex < len(vm.chunk.SourceSpans) {
		sourceSpan = vm.chunk.SourceSpans[instructionIndex]
	}

	location := formatBytecodeSourceSpan(sourceSpan)

	if location == "<unknown>" {
		if vm.chunk == nil {
			return fmt.Errorf("bytecode runtime error: %s", message)
		}

		return fmt.Errorf(
			"bytecode runtime error in %q at instruction %04d: %s",
			vm.chunk.Name,
			instructionIndex,
			message,
		)
	}

	return fmt.Errorf(
		"bytecode runtime error at %s instruction %04d: %s",
		location,
		instructionIndex,
		message,
	)
}
