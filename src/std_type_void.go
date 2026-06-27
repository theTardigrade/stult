package main

import "fmt"

func NewStdTypeVoidMap() Value {
	entries := map[string]Binding{
		"NEW": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeVoidNew)),
	}

	return NewMapValue(entries, true)
}

func StdTypeVoidNew(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return Value{}, fmt.Errorf("TYPE.VOID.NEW expected 0 arguments, got %d", len(args))
	}

	return NewVoidValue(), nil
}
