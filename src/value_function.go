package main

type Function struct {
	Parameters        []Token
	VariadicParameter *Token
	Body              []Statement
	Returns           []Expression
	Env               *Environment
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
