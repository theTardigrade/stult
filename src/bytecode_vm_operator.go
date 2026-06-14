package main

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
