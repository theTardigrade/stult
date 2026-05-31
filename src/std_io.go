package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var stdIOStdinReader = bufio.NewReader(os.Stdin)

func NewStdIOMap() Value {
	entries := map[string]Binding{
		"PRINT":       NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOWriteLine)), // alias for WRITE_LINE
		"PROMPT":      NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOPrompt)),
		"READ_LINE":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOReadLine)),
		"WRITE":       NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOWrite)),
		"WRITE_ERROR": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOWriteError)),
		"WRITE_LINE":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOWriteLine)),
	}

	return NewMapValue(entries, true)
}

func builtinStdIOWrite(_ *RuntimeContext, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Print(arg.PrintString())
	}

	return NewVoidValue(), nil
}

func builtinStdIOWriteLine(_ *RuntimeContext, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Print(arg.PrintString())
	}

	fmt.Println()

	return NewVoidValue(), nil
}

func builtinStdIOWriteError(_ *RuntimeContext, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Fprint(os.Stderr, arg.PrintString())
	}

	fmt.Fprintln(os.Stderr)

	return NewVoidValue(), nil
}

func builtinStdIOReadLine(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return Value{}, fmt.Errorf("IO.READ_LINE expected 0 arguments, got %d", len(args))
	}

	return stdIOReadLine()
}

func builtinStdIOPrompt(_ *RuntimeContext, args []Value) (Value, error) {
	for _, arg := range args {
		fmt.Print(arg.PrintString())
	}

	return stdIOReadLine()
}

func stdIOReadLine() (Value, error) {
	line, err := stdIOStdinReader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return Value{}, err
	}

	if errors.Is(err, io.EOF) && line == "" {
		return NewVoidValue(), nil
	}

	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	return NewStringValue(line), nil
}
