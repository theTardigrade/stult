package main

import (
	"fmt"
)

func NewStdIOMap() Value {
	entries := map[string]Binding{
		"PRINT": {
			Value:       NewBuiltinFunctionValue(builtinStdIOPrint),
			IsImmutable: true,
		},
	}

	return NewMapValue(entries, true)
}

func builtinStdIOPrint(_ *Interpreter, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Print(arg.PrintString())
	}

	fmt.Println()

	return NewEmptyValue(), nil
}
