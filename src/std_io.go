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
	for index, arg := range args {
		if index > 0 {
			fmt.Print(" ")
		}

		fmt.Print(arg.PrintString())
	}

	fmt.Println()

	return NewBoolValue(true), nil
}
