package main

import (
	"fmt"
	"os"
)

func NewStdIOOutputMap() Value {
	entries := map[string]Binding{
		"PRINT":       NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOOutputWriteLine)), // alias for WRITE_LINE
		"WRITE":       NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOOutputWrite)),
		"WRITE_ERROR": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOOutputWriteError)),
		"WRITE_LINE":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOOutputWriteLine)),
	}

	return NewMapValue(entries, true)
}

func builtinStdIOOutputWrite(_ *RuntimeContext, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Print(arg.PrintString())
	}

	return NewVoidValue(), nil
}

func builtinStdIOOutputWriteLine(_ *RuntimeContext, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Print(arg.PrintString())
	}

	fmt.Println()

	return NewVoidValue(), nil
}

func builtinStdIOOutputWriteError(_ *RuntimeContext, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Fprint(os.Stderr, arg.PrintString())
	}

	fmt.Fprintln(os.Stderr)

	return NewVoidValue(), nil
}
