package main

type Function struct {
	Parameters []Token
	Body       []Statement
	Returns    []Expression
	Env        *Environment
}

type BuiltinFunction func(interpreter *Interpreter, args []Value) (Value, error)

func NewBuiltinFunctionValue(fn BuiltinFunction) Value {
	return Value{
		Kind:            ValueBuiltinFunction,
		BuiltinFunction: fn,
	}
}

func NewFunctionValue(fn *Function) Value {
	return Value{
		Kind:     ValueFunction,
		Function: fn,
	}
}
