package main

import "fmt"

func (vm *BytecodeVM) makeFunction(index int) error {
	if vm.chunk == nil {
		return fmt.Errorf("no active bytecode chunk")
	}

	if index < 0 || index >= len(vm.chunk.Functions) {
		return fmt.Errorf("function index %d out of bounds", index)
	}

	function := vm.chunk.Functions[index]
	upvalues, err := vm.captureUpvalueCells(function.Upvalues)
	if err != nil {
		return err
	}

	value := NewBuiltinFunctionValue(func(_ *RuntimeContext, args []Value) (Value, error) {
		return vm.runFunction(function, upvalues, args)
	})

	vm.pushValue(value)

	return nil
}

func (vm *BytecodeVM) captureUpvalueCells(upvalues []BytecodeUpvalue) ([]*bytecodeVMCell, error) {
	cells := make([]*bytecodeVMCell, 0, len(upvalues))

	for _, upvalue := range upvalues {
		if upvalue.IsLocal {
			cell, err := vm.localCell(upvalue.Index)
			if err != nil {
				return nil, err
			}

			cells = append(cells, cell)
			continue
		}

		cell, err := vm.upvalueCell(upvalue.Index)
		if err != nil {
			return nil, err
		}

		cells = append(cells, cell)
	}

	return cells, nil
}

func (vm *BytecodeVM) runFunction(
	function BytecodeFunction,
	upvalues []*bytecodeVMCell,
	args []Value,
) (Value, error) {
	if function.Chunk == nil {
		return Value{}, fmt.Errorf("function %s has no chunk", function.Name)
	}

	previousState := vm.saveExecutionState()

	vm.chunk = function.Chunk
	vm.ip = 0
	vm.stack = make([]bytecodeVMStackEntry, 0, 8)
	vm.upvalues = upvalues
	vm.iterators = []bytecodeVMIterator{}

	vm.initializeLocals(function.Chunk)

	if err := vm.bindFunctionArguments(function, args); err != nil {
		vm.restoreExecutionState(previousState)
		return Value{}, err
	}

	result, err := vm.runActiveChunk()

	vm.restoreExecutionState(previousState)

	return result, err
}

func (vm *BytecodeVM) saveExecutionState() bytecodeVMExecutionState {
	return bytecodeVMExecutionState{
		Chunk:     vm.chunk,
		IP:        vm.ip,
		Stack:     vm.stack,
		Locals:    vm.locals,
		Upvalues:  vm.upvalues,
		Iterators: vm.iterators,
	}
}

func (vm *BytecodeVM) restoreExecutionState(state bytecodeVMExecutionState) {
	vm.chunk = state.Chunk
	vm.ip = state.IP
	vm.stack = state.Stack
	vm.locals = state.Locals
	vm.upvalues = state.Upvalues
	vm.iterators = state.Iterators
}

func (vm *BytecodeVM) bindFunctionArguments(function BytecodeFunction, args []Value) error {
	requiredCount := requiredBytecodeParameterCount(function.Parameters)
	maxCount := len(function.Parameters)

	if len(args) < requiredCount {
		if function.VariadicParameter == nil && requiredCount == maxCount {
			return fmt.Errorf(
				"function expected %d argument(s), got %d",
				requiredCount,
				len(args),
			)
		}

		return fmt.Errorf(
			"function expected at least %d argument(s), got %d",
			requiredCount,
			len(args),
		)
	}

	if function.VariadicParameter == nil && len(args) > maxCount {
		if requiredCount == maxCount {
			return fmt.Errorf(
				"function expected %d argument(s), got %d",
				maxCount,
				len(args),
			)
		}

		return fmt.Errorf(
			"function expected at most %d argument(s), got %d",
			maxCount,
			len(args),
		)
	}

	for index, parameter := range function.Parameters {
		if parameter.Name == "_" {
			continue
		}

		localIndex, ok := vm.localIndexByName(parameter.Name)
		if !ok {
			return fmt.Errorf("function local for parameter %q was not found", parameter.Name)
		}

		value := NewVoidValue()
		if index < len(args) {
			value = args[index]
		}

		if err := vm.storeLocal(localIndex, value, parameter.IsImmutable, true); err != nil {
			return err
		}
	}

	if function.VariadicParameter != nil && function.VariadicParameter.Name != "_" {
		localIndex, ok := vm.localIndexByName(function.VariadicParameter.Name)
		if !ok {
			return fmt.Errorf(
				"function local for variadic parameter %q was not found",
				function.VariadicParameter.Name,
			)
		}

		variadicStart := len(function.Parameters)
		if len(args) < variadicStart {
			variadicStart = len(args)
		}

		extra := append([]Value{}, args[variadicStart:]...)

		if err := vm.storeLocal(
			localIndex,
			NewArrayValue(extra, false),
			function.VariadicParameter.IsImmutable,
			true,
		); err != nil {
			return err
		}
	}

	return nil
}

func requiredBytecodeParameterCount(parameters []BytecodeParameter) int {
	count := 0

	for _, parameter := range parameters {
		if parameter.IsOptional {
			continue
		}

		count++
	}

	return count
}

func (vm *BytecodeVM) localIndexByName(name string) (int, bool) {
	if vm.chunk == nil {
		return 0, false
	}

	if vm.localIndexCache == nil {
		vm.localIndexCache = map[*BytecodeChunk]map[string]int{}
	}

	indices, ok := vm.localIndexCache[vm.chunk]
	if !ok {
		indices = make(map[string]int, len(vm.chunk.Locals))

		for index, local := range vm.chunk.Locals {
			if _, exists := indices[local.Name]; !exists {
				indices[local.Name] = index
			}
		}

		vm.localIndexCache[vm.chunk] = indices
	}

	index, ok := indices[name]

	return index, ok
}

func (vm *BytecodeVM) callValue(argCount int) error {
	if argCount < 0 {
		return fmt.Errorf("CALL argument count cannot be negative")
	}

	if len(vm.stack) < argCount+1 {
		return fmt.Errorf("CALL expected callee and %d argument(s)", argCount)
	}

	args := make([]Value, argCount)

	for index := argCount - 1; index >= 0; index-- {
		arg, err := vm.popValue()
		if err != nil {
			return err
		}

		args[index] = arg
	}

	callee, err := vm.popValue()
	if err != nil {
		return err
	}

	callee = resolveSpecializedValue(callee)

	switch callee.Kind {
	case ValueBuiltinFunction:
		result, err := callee.BuiltinFunction(vm.runtime, args)
		if err != nil {
			return err
		}

		vm.pushValue(result)
		return nil

	default:
		return fmt.Errorf("value is not callable")
	}
}
