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
	resolvedFunction, err := vm.bytecodeFunctionWithResolvedParameterContracts(function)
	if err != nil {
		return err
	}

	upvalues, err := vm.captureUpvalueCells(resolvedFunction.Upvalues)
	if err != nil {
		return err
	}

	value := NewFunctionValue(&Function{
		BytecodeFunction: &resolvedFunction,
		BytecodeUpvalues: upvalues,
		ReturnContract:   cloneBindingContractPointer(resolvedFunction.ReturnContract),
		DotMap:           vm.currentDotMap,
	})

	vm.pushValue(value)

	return nil
}

func (vm *BytecodeVM) bytecodeFunctionWithResolvedParameterContracts(function BytecodeFunction) (BytecodeFunction, error) {
	resolvedFunction := function
	resolvedFunction.Parameters = make([]BytecodeParameter, len(function.Parameters))

	for index, parameter := range function.Parameters {
		resolvedContract, err := vm.resolveBindingContractAliases(parameter.Contract)
		if err != nil {
			return BytecodeFunction{}, fmt.Errorf("function parameter %q: %w", parameter.Name, err)
		}

		resolvedFunction.Parameters[index] = parameter
		resolvedFunction.Parameters[index].Contract = resolvedContract
	}

	if function.VariadicParameter != nil {
		resolvedContract, err := vm.resolveBindingContractAliases(function.VariadicParameter.Contract)
		if err != nil {
			return BytecodeFunction{}, fmt.Errorf("variadic function parameter %q: %w", function.VariadicParameter.Name, err)
		}

		resolvedVariadicParameter := *function.VariadicParameter
		resolvedVariadicParameter.Contract = resolvedContract
		resolvedFunction.VariadicParameter = &resolvedVariadicParameter
	}

	if function.ReturnContract != nil {
		resolvedContract, err := vm.resolveBindingContractAliases(*function.ReturnContract)
		if err != nil {
			return BytecodeFunction{}, fmt.Errorf("function return contract: %w", err)
		}

		resolvedFunction.ReturnContract = resolvedContract.ClonePointer()
	}

	return resolvedFunction, nil
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
	dotMap *Map,
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
	vm.errorHandlers = []bytecodeVMErrorHandler{}
	vm.currentDotMap = dotMap
	vm.dotMapStack = []*Map{}

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
		Chunk:         vm.chunk,
		IP:            vm.ip,
		Stack:         vm.stack,
		Locals:        vm.locals,
		Upvalues:      vm.upvalues,
		Iterators:     vm.iterators,
		ErrorHandlers: vm.errorHandlers,
		CurrentDotMap: vm.currentDotMap,
		DotMapStack:   vm.dotMapStack,
	}
}

func (vm *BytecodeVM) restoreExecutionState(state bytecodeVMExecutionState) {
	vm.chunk = state.Chunk
	vm.ip = state.IP
	vm.stack = state.Stack
	vm.locals = state.Locals
	vm.upvalues = state.Upvalues
	vm.iterators = state.Iterators
	vm.errorHandlers = state.ErrorHandlers
	vm.currentDotMap = state.CurrentDotMap
	vm.dotMapStack = state.DotMapStack
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
		value := NewVoidValue()
		if index < len(args) {
			value = args[index]
		}

		if parameter.Name == "_" {
			if err := parameter.Contract.CheckAndLearn("function parameter _", value); err != nil {
				return err
			}
			continue
		}

		localIndex, ok := vm.localIndexByName(parameter.Name)
		if !ok {
			return fmt.Errorf("function local for parameter %q was not found", parameter.Name)
		}

		if err := vm.storeLocal(localIndex, value, parameter.IsImmutable, true, true, parameter.Contract); err != nil {
			return err
		}
	}

	if function.VariadicParameter != nil {
		variadicStart := len(function.Parameters)
		if len(args) < variadicStart {
			variadicStart = len(args)
		}

		extra := append([]Value{}, args[variadicStart:]...)
		variadicValue := NewArrayValue(extra, false)

		if function.VariadicParameter.Name == "_" {
			return function.VariadicParameter.Contract.CheckAndLearn("variadic function parameter _", variadicValue)
		}

		localIndex, ok := vm.localIndexByName(function.VariadicParameter.Name)
		if !ok {
			return fmt.Errorf(
				"function local for variadic parameter %q was not found",
				function.VariadicParameter.Name,
			)
		}

		if err := vm.storeLocal(
			localIndex,
			variadicValue,
			function.VariadicParameter.IsImmutable,
			true,
			true,
			function.VariadicParameter.Contract,
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

func bytecodeFunctionCanAcceptArgumentCount(function BytecodeFunction, count int) bool {
	requiredCount := requiredBytecodeParameterCount(function.Parameters)
	maxCount := len(function.Parameters)

	if count < requiredCount {
		return false
	}

	if function.VariadicParameter != nil {
		return true
	}

	return count <= maxCount
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

	entries := make([]bytecodeVMStackEntry, argCount)

	for index := argCount - 1; index >= 0; index-- {
		entry, err := vm.popEntry()
		if err != nil {
			return err
		}

		if entry.IsRangeSegment {
			return fmt.Errorf("range segment cannot be used as a function argument")
		}

		entries[index] = entry
	}

	callee, err := vm.popValue()
	if err != nil {
		return err
	}

	args := make([]Value, 0, argCount)

	for _, entry := range entries {
		if entry.IsSpreadArgument {
			spreadValues, err := callSpreadArgumentValues(entry.Value)
			if err != nil {
				return err
			}

			args = append(args, spreadValues...)
			continue
		}

		args = append(args, entry.Value)
	}

	result, err := vm.callResolvedValue(callee, args)
	if err != nil {
		return err
	}

	vm.pushValue(result)
	return nil
}

func (vm *BytecodeVM) callResolvedValue(callee Value, args []Value) (Value, error) {
	callee = resolveSpecializedValue(callee)

	switch callee.Kind {
	case ValueBuiltinFunction:
		return callee.BuiltinFunction(vm.runtime, args)

	case ValueFunction:
		return vm.callFunction(callee.Function, args)

	default:
		return Value{}, fmt.Errorf("value is not callable")
	}
}

func (vm *BytecodeVM) callFunction(fn *Function, args []Value) (Value, error) {
	if fn == nil {
		return Value{}, fmt.Errorf("invalid function")
	}

	if err := checkFunctionSignatureArguments(fn.SignatureContract, args); err != nil {
		return Value{}, err
	}

	if fn.BytecodeFunction == nil {
		return Value{}, fmt.Errorf("bytecode VM cannot call an interpreter function")
	}

	result, err := vm.runFunction(*fn.BytecodeFunction, fn.BytecodeUpvalues, fn.DotMap, args)
	if err != nil {
		return Value{}, err
	}

	return checkFunctionReturnContracts(fn.ReturnContract, fn.SignatureContract, result)
}
