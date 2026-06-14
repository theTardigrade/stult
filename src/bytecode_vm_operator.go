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

	vm.pushValue(NewNumberValueFromNumber(numberNegate(right.Number)))

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

	vm.pushValue(NewNumberValueFromNumber(numberAdd(left.Number, right.Number)))

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

	vm.pushValue(NewNumberValueFromNumber(numberSubtract(left.Number, right.Number)))

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

	vm.pushValue(NewNumberValueFromNumber(numberMultiply(left.Number, right.Number)))

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

	out, err := numberDivide(left.Number, right.Number)
	if err != nil {
		return vm.runtimeError(instructionIndex, "%s", err.Error())
	}

	vm.pushValue(NewNumberValueFromNumber(out))

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

	vm.pushValue(NewBoolValue(numberCompare(left.Number, right.Number) < 0))

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

	vm.pushValue(NewBoolValue(numberCompare(left.Number, right.Number) <= 0))

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

	vm.pushValue(NewBoolValue(numberCompare(left.Number, right.Number) > 0))

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

	vm.pushValue(NewBoolValue(numberCompare(left.Number, right.Number) >= 0))

	return nil
}
