package main

type Function struct {
	Parameters        []FunctionParameter
	VariadicParameter *Token
	Body              []Statement
	Returns           []Expression
	Env               *Environment
	BytecodeFunction  *BytecodeFunction
	BytecodeUpvalues  []*bytecodeVMCell
}

type BuiltinFunction func(runtime *RuntimeContext, args []Value) (Value, error)

func NewFunctionValue(fn *Function) Value {
	return Value{
		Kind:     ValueFunction,
		Function: fn,
	}
}

func NewBuiltinFunctionValue(fn BuiltinFunction) Value {
	return Value{
		Kind:            ValueBuiltinFunction,
		BuiltinFunction: fn,
	}
}

func functionCanAcceptArgumentCount(fn *Function, count int) bool {
	if fn == nil {
		return false
	}

	if fn.BytecodeFunction != nil {
		return bytecodeFunctionCanAcceptArgumentCount(*fn.BytecodeFunction, count)
	}

	requiredCount := requiredFunctionParameterCount(fn.Parameters)
	maxCount := len(fn.Parameters)

	if count < requiredCount {
		return false
	}

	if fn.VariadicParameter != nil {
		return true
	}

	return count <= maxCount
}
