package main

import "fmt"

type BytecodeVM struct {
	chunk                    *BytecodeChunk
	ip                       int
	stack                    []bytecodeVMStackEntry
	globals                  map[string]Binding
	locals                   []bytecodeVMCell
	upvalues                 []*bytecodeVMCell
	iterators                []bytecodeVMIterator
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
	Value          Value
	IsRangeSegment bool
}

type bytecodeVMIterator struct {
	Collection    Value
	Position      int
	CurrentKey    Value
	CurrentValue  Value
	HasCurrent    bool
	LastMapKey    string
	HasLastMapKey bool
}

type bytecodeVMExecutionState struct {
	Chunk     *BytecodeChunk
	IP        int
	Stack     []bytecodeVMStackEntry
	Locals    []bytecodeVMCell
	Upvalues  []*bytecodeVMCell
	Iterators []bytecodeVMIterator
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
		locals:                   []bytecodeVMCell{},
		upvalues:                 []*bytecodeVMCell{},
		iterators:                []bytecodeVMIterator{},
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
			return Value{}, err
		}

		if returned {
			return result, nil
		}
	}

	return NewVoidValue(), nil
}

func (vm *BytecodeVM) initializeLocals(chunk *BytecodeChunk) {
	vm.locals = make([]bytecodeVMCell, len(chunk.Locals))

	for index, local := range chunk.Locals {
		vm.locals[index] = bytecodeVMCell{
			Value:       NewVoidValue(),
			Initialized: false,
			IsImmutable: local.IsImmutable,
		}
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

	case BytecodeOpBuildRange:
		if err := vm.buildRange(instruction.Operand == 1); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpMakeFunction:
		if err := vm.makeFunction(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
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

	case BytecodeOpCall:
		if err := vm.callValue(instruction.Operand); err != nil {
			return Value{}, false, vm.runtimeError(instructionIndex, "%s", err.Error())
		}

		return Value{}, false, nil

	case BytecodeOpIteratorInit:
		if err := vm.iteratorInit(); err != nil {
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

		if bytecodeValueIsCollection(value) {
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

func (vm *BytecodeVM) popResolvedValue(instructionIndex int) (Value, error) {
	value, err := vm.popValue()
	if err != nil {
		return Value{}, vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	return resolveSpecializedValue(value), nil
}

func (vm *BytecodeVM) popResolvedBinaryValues(instructionIndex int) (Value, Value, error) {
	right, err := vm.popResolvedValue(instructionIndex)
	if err != nil {
		return Value{}, Value{}, err
	}

	left, err := vm.popResolvedValue(instructionIndex)
	if err != nil {
		return Value{}, Value{}, err
	}

	return left, right, nil
}

func (vm *BytecodeVM) applyNegate(instructionIndex int) error {
	right, err := vm.popResolvedValue(instructionIndex)
	if err != nil {
		return err
	}

	if right.Kind != ValueNumber {
		return vm.runtimeError(instructionIndex, "unary '-' requires a number")
	}

	out := CloneNumber(right.Number)
	out.Neg(out)

	vm.pushValue(Value{Kind: ValueNumber, Number: out})

	return nil
}

func (vm *BytecodeVM) applyNot(instructionIndex int) error {
	right, err := vm.popResolvedValue(instructionIndex)
	if err != nil {
		return err
	}

	if right.Kind != ValueBool {
		return vm.runtimeError(instructionIndex, "unary '!' requires a bool")
	}

	vm.pushValue(NewBoolValue(!right.Bool))

	return nil
}

func (vm *BytecodeVM) applyAdd(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	if left.Kind == ValueString && right.Kind == ValueString {
		vm.pushValue(NewStringValue(left.Text.String() + right.Text.String()))
		return nil
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return vm.runtimeError(instructionIndex, "operator %q requires numbers", "+")
	}

	out := newFloat()
	out.Add(left.Number, right.Number)

	vm.pushValue(Value{Kind: ValueNumber, Number: out})

	return nil
}

func (vm *BytecodeVM) applySubtract(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return vm.runtimeError(instructionIndex, "operator %q requires numbers", "-")
	}

	out := newFloat()
	out.Sub(left.Number, right.Number)

	vm.pushValue(Value{Kind: ValueNumber, Number: out})

	return nil
}

func (vm *BytecodeVM) applyMultiply(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return vm.runtimeError(instructionIndex, "operator %q requires numbers", "*")
	}

	out := newFloat()
	out.Mul(left.Number, right.Number)

	vm.pushValue(Value{Kind: ValueNumber, Number: out})

	return nil
}

func (vm *BytecodeVM) applyDivide(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return vm.runtimeError(instructionIndex, "operator %q requires numbers", "/")
	}

	if right.Number.Sign() == 0 {
		return vm.runtimeError(instructionIndex, "division by zero")
	}

	out := newFloat()
	out.Quo(left.Number, right.Number)

	vm.pushValue(Value{Kind: ValueNumber, Number: out})

	return nil
}

func (vm *BytecodeVM) applyEqual(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	equal, err := valuesEqual(left, right)
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	vm.pushValue(NewBoolValue(equal))

	return nil
}

func (vm *BytecodeVM) applyNotEqual(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	equal, err := valuesEqual(left, right)
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	vm.pushValue(NewBoolValue(!equal))

	return nil
}

func (vm *BytecodeVM) applyLess(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return vm.runtimeError(instructionIndex, "operator %q requires numbers", "<")
	}

	vm.pushValue(NewBoolValue(left.Number.Cmp(right.Number) < 0))

	return nil
}

func (vm *BytecodeVM) applyLessEqual(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return vm.runtimeError(instructionIndex, "operator %q requires numbers", "<=")
	}

	vm.pushValue(NewBoolValue(left.Number.Cmp(right.Number) <= 0))

	return nil
}

func (vm *BytecodeVM) applyGreater(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return vm.runtimeError(instructionIndex, "operator %q requires numbers", ">")
	}

	vm.pushValue(NewBoolValue(left.Number.Cmp(right.Number) > 0))

	return nil
}

func (vm *BytecodeVM) applyGreaterEqual(instructionIndex int) error {
	left, right, err := vm.popResolvedBinaryValues(instructionIndex)
	if err != nil {
		return err
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return vm.runtimeError(instructionIndex, "operator %q requires numbers", ">=")
	}

	vm.pushValue(NewBoolValue(left.Number.Cmp(right.Number) >= 0))

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
	fixedCount := len(function.Parameters)

	if function.VariadicParameter == nil {
		if len(args) != fixedCount {
			return fmt.Errorf(
				"function expected %d argument(s), got %d",
				fixedCount,
				len(args),
			)
		}
	} else if len(args) < fixedCount {
		return fmt.Errorf(
			"function expected at least %d argument(s), got %d",
			fixedCount,
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

		if err := vm.storeLocal(localIndex, args[index], parameter.IsImmutable, true); err != nil {
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

		extra := append([]Value{}, args[fixedCount:]...)

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
