package main

import (
	"fmt"
)

func NewStdIOMap() Value {
	entries := map[string]Binding{
		"PRINT": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOPrint)),
	}

	return NewMapValue(entries, true)
}

func builtinStdIOPrint(_ *Interpreter, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Print(arg.PrintString())
	}

	fmt.Println()

	return NewVoidValue(), nil
}
