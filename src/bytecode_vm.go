package main

import "fmt"

type BytecodeVM struct {
	chunk                    *BytecodeChunk
	ip                       int
	stack                    []bytecodeVMStackEntry
	globals                  map[string]Binding
	locals                   []*bytecodeVMCell
	upvalues                 []*bytecodeVMCell
	iterators                []bytecodeVMIterator
	errorHandlers            []bytecodeVMErrorHandler
	currentDotMap            *Map
	dotMapStack              []*Map
	runtime                  *RuntimeContext
	args                     []string
	localIndexCache          map[*BytecodeChunk]map[string]int
	resetLocalIndexesByDepth map[*BytecodeChunk]map[int][]int
}

type bytecodeVMCell struct {
	Value       Value
	Initialized bool
	IsImmutable bool
}

type bytecodeVMStackEntry struct {
	Value            Value
	IsRangeSegment   bool
	IsSpreadArgument bool
}

type bytecodeVMExecutionState struct {
	Chunk         *BytecodeChunk
	IP            int
	Stack         []bytecodeVMStackEntry
	Locals        []*bytecodeVMCell
	Upvalues      []*bytecodeVMCell
	Iterators     []bytecodeVMIterator
	ErrorHandlers []bytecodeVMErrorHandler
	CurrentDotMap *Map
	DotMapStack   []*Map
}

func RunBytecode(chunk *BytecodeChunk) (Value, error) {
	return RunBytecodeWithArgs(chunk, nil)
}

func RunBytecodeWithArgs(chunk *BytecodeChunk, args []string) (Value, error) {
	vm := NewBytecodeVM(args)

	return vm.Run(chunk)
}

func NewBytecodeVM(args []string) *BytecodeVM {
	return NewBytecodeVMWithRuntime(NewRuntimeContext(args))
}

func NewBytecodeVMWithRuntime(runtime *RuntimeContext) *BytecodeVM {
	if runtime == nil {
		runtime = NewRuntimeContext(nil)
	}

	return &BytecodeVM{
		stack:                    []bytecodeVMStackEntry{},
		globals:                  bytecodeInitialGlobals(runtime),
		locals:                   []*bytecodeVMCell{},
		upvalues:                 []*bytecodeVMCell{},
		iterators:                []bytecodeVMIterator{},
		errorHandlers:            []bytecodeVMErrorHandler{},
		dotMapStack:              []*Map{},
		runtime:                  runtime,
		args:                     append([]string{}, runtime.Args...),
		localIndexCache:          map[*BytecodeChunk]map[string]int{},
		resetLocalIndexesByDepth: map[*BytecodeChunk]map[int][]int{},
	}
}

func bytecodeInitialGlobals(runtime *RuntimeContext) map[string]Binding {
	return map[string]Binding{
		"STD": NewImmutableBinding(NewStdMap(runtime)),
	}
}

func (vm *BytecodeVM) Run(chunk *BytecodeChunk) (Value, error) {
	return vm.runChunk(chunk, true)
}

func (vm *BytecodeVM) runChunk(chunk *BytecodeChunk, initializeLocals bool) (Value, error) {
	if chunk == nil {
		return Value{}, fmt.Errorf("bytecode VM cannot run a nil chunk")
	}

	vm.chunk = chunk
	vm.ip = 0
	vm.stack = vm.stack[:0]
	vm.iterators = vm.iterators[:0]
	vm.errorHandlers = vm.errorHandlers[:0]
	vm.currentDotMap = nil
	vm.dotMapStack = vm.dotMapStack[:0]

	if initializeLocals {
		vm.initializeLocals(chunk)
	}

	return vm.runActiveChunk()
}

func (vm *BytecodeVM) runActiveChunk() (Value, error) {
	for vm.ip < len(vm.chunk.Instructions) {
		instructionIndex := vm.ip
		instruction := vm.chunk.Instructions[instructionIndex]
		vm.ip++

		result, returned, err := vm.executeInstruction(instructionIndex, instruction)
		if err != nil {
			if vm.handleRuntimeError(err) {
				continue
			}

			return Value{}, err
		}

		if returned {
			return result, nil
		}
	}

	return NewVoidValue(), nil
}

func (vm *BytecodeVM) initializeLocals(chunk *BytecodeChunk) {
	vm.locals = make([]*bytecodeVMCell, len(chunk.Locals))

	for index, local := range chunk.Locals {
		vm.locals[index] = newBytecodeVMCell(local)
	}
}

func (vm *BytecodeVM) executeInstruction(
	instructionIndex int,
	instruction BytecodeInstruction,
) (Value, bool, error) {
	switch instruction.Opcode {
	case BytecodeOpLoadVoid:
		vm.pushValue(NewVoidValue())
		return Value{}, false, nil

	case BytecodeOpLoadConst:
		value, err := vm.constantValue(instruction.Operand)
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		vm.pushValue(bytecodeCloneValueForLoad(value))
		return Value{}, false, nil

	case BytecodeOpLoadTrue:
		vm.pushValue(NewBoolValue(true))
		return Value{}, false, nil

	case BytecodeOpLoadFalse:
		vm.pushValue(NewBoolValue(false))
		return Value{}, false, nil

	case BytecodeOpLoadGlobal:
		name, err := vm.constantName(instruction.Operand)
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		value, err := vm.loadGlobal(name)
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		vm.pushValue(value)
		return Value{}, false, nil

	case BytecodeOpLoadLocal:
		value, err := vm.loadLocal(instruction.Operand)
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		vm.pushValue(value)
		return Value{}, false, nil

	case BytecodeOpLoadUpvalue:
		value, err := vm.loadUpvalue(instruction.Operand)
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		vm.pushValue(value)
		return Value{}, false, nil

	case BytecodeOpStoreGlobalMutable:
		return Value{}, false, vm.storeGlobalFromStack(instructionIndex, instruction.Operand, false)

	case BytecodeOpStoreGlobalImmutable:
		return Value{}, false, vm.storeGlobalFromStack(instructionIndex, instruction.Operand, true)

	case BytecodeOpStoreExistingGlobal:
		return Value{}, false, vm.storeExistingGlobalFromStack(instructionIndex, instruction.Operand)

	case BytecodeOpStoreLocalMutable:
		return Value{}, false, vm.storeLocalFromStack(instructionIndex, instruction.Operand, false)

	case BytecodeOpStoreLocalImmutable:
		return Value{}, false, vm.storeLocalFromStack(instructionIndex, instruction.Operand, true)

	case BytecodeOpStoreUpvalueMutable:
		return Value{}, false, vm.storeUpvalueFromStack(instructionIndex, instruction.Operand, false)

	case BytecodeOpStoreUpvalueImmutable:
		return Value{}, false, vm.storeUpvalueFromStack(instructionIndex, instruction.Operand, true)

	case BytecodeOpStoreIndex:
		value, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		index, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		object, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		if _, err := bytecodeAssignIndex(object, index, value); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpBuildArray:
		if err := vm.buildArray(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpBuildMap:
		if err := vm.buildMap(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpBeginMap:
		vm.beginMapLiteral()
		return Value{}, false, nil

	case BytecodeOpCheckMapEntry:
		if err := vm.checkMapEntry(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpAddMapEntry:
		if err := vm.addMapEntry(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpEndMap:
		if err := vm.endMapLiteral(); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpBuildRange:
		if err := vm.buildRange(instruction.Operand == 1); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpFreezeCollection:
		if err := vm.freezeCollectionFromStack(); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpMakeFunction:
		if err := vm.makeFunction(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpLoadDotMap:
		if err := vm.loadDotMap(instructionIndex); err != nil {
			return Value{}, false, err
		}

		return Value{}, false, nil

	case BytecodeOpIndex:
		index, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		object, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		value, err := bytecodeIndexValue(object, index)
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		vm.pushValue(value)
		return Value{}, false, nil

	case BytecodeOpRangeIndex:
		step, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		end, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		start, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		object, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		value, err := rangeIndexValue(object, start, end, step, instruction.Operand == 1)
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		vm.pushValue(value)
		return Value{}, false, nil

	case BytecodeOpCall:
		if err := vm.callValue(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpSpreadArgument:
		if err := vm.markSpreadArgument(); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpIteratorInit:
		if err := vm.iteratorInit(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpIteratorRangeInit:
		if err := vm.iteratorRangeInit(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpIteratorNext:
		if err := vm.iteratorNext(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpIteratorValue:
		return Value{}, false, vm.storeIteratorValue(instructionIndex, instruction.Operand)

	case BytecodeOpIteratorKey:
		return Value{}, false, vm.storeIteratorKey(instructionIndex, instruction.Operand)

	case BytecodeOpIteratorCollection:
		return Value{}, false, vm.storeIteratorCollection(instructionIndex, instruction.Operand)

	case BytecodeOpIteratorPosition:
		return Value{}, false, vm.storeIteratorPosition(instructionIndex, instruction.Operand)

	case BytecodeOpIteratorEnd:
		if len(vm.iterators) == 0 {
			return Value{}, false, vm.runtimeError(instructionIndex, "no active iterator to end")
		}

		vm.iterators = vm.iterators[:len(vm.iterators)-1]
		return Value{}, false, nil

	case BytecodeOpResetLocals:
		if err := vm.resetLocalsFromDepth(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpTryStart:
		if err := vm.tryStart(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpTryEnd:
		if err := vm.tryEnd(); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpJump:
		if err := vm.jump(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpJumpIfFalse:
		value, err := vm.peekValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		condition, err := bytecodeBoolValue(value)
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		if !condition {
			if err := vm.jump(instruction.Operand); err != nil {
				return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
			}
		}

		return Value{}, false, nil

	case BytecodeOpJumpIfTrue:
		value, err := vm.peekValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		condition, err := bytecodeBoolValue(value)
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		if condition {
			if err := vm.jump(instruction.Operand); err != nil {
				return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
			}
		}

		return Value{}, false, nil

	case BytecodeOpJumpIfCollection:
		value, err := vm.peekValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		if bytecodeValueIsLoopIterable(value) {
			if err := vm.jump(instruction.Operand); err != nil {
				return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
			}
		}

		return Value{}, false, nil

	case BytecodeOpDuplicateTopTwo:
		if len(vm.stack) < 2 {
			return Value{}, false, vm.runtimeError(
				instructionIndex,
				"DUPLICATE_TOP_TWO requires at least two stack values",
			)
		}

		left := vm.stack[len(vm.stack)-2]
		right := vm.stack[len(vm.stack)-1]

		vm.stack = append(vm.stack, left, right)
		return Value{}, false, nil

	case BytecodeOpPositive:
		return Value{}, false, vm.applyPositive(instructionIndex)

	case BytecodeOpNegate:
		return Value{}, false, vm.applyNegate(instructionIndex)

	case BytecodeOpNot:
		return Value{}, false, vm.applyNot(instructionIndex)

	case BytecodeOpAdd:
		return Value{}, false, vm.applyAdd(instructionIndex)

	case BytecodeOpSubtract:
		return Value{}, false, vm.applySubtract(instructionIndex)

	case BytecodeOpMultiply:
		return Value{}, false, vm.applyMultiply(instructionIndex)

	case BytecodeOpDivide:
		return Value{}, false, vm.applyDivide(instructionIndex)

	case BytecodeOpEqual:
		return Value{}, false, vm.applyEqual(instructionIndex)

	case BytecodeOpNotEqual:
		return Value{}, false, vm.applyNotEqual(instructionIndex)

	case BytecodeOpLess:
		return Value{}, false, vm.applyLess(instructionIndex)

	case BytecodeOpLessEqual:
		return Value{}, false, vm.applyLessEqual(instructionIndex)

	case BytecodeOpGreater:
		return Value{}, false, vm.applyGreater(instructionIndex)

	case BytecodeOpGreaterEqual:
		return Value{}, false, vm.applyGreaterEqual(instructionIndex)

	case BytecodeOpPop:
		if _, err := vm.popEntry(); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpReturn:
		if len(vm.stack) == 0 {
			return NewVoidValue(), true, nil
		}

		value, err := vm.popValue()
		if err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return value, true, nil

	default:
		return Value{}, false, vm.runtimeError(
			instructionIndex,
			"unknown bytecode opcode %d",
			instruction.Opcode,
		)
	}
}
