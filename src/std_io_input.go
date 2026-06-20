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

func NewStdIOInputMap() Value {
	entries := map[string]Binding{
		"PROMPT":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOInputPrompt)),
		"READ_LINE": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdIOInputReadLine)),
	}

	return NewMapValue(entries, true)
}

func builtinStdIOInputReadLine(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return Value{}, fmt.Errorf("IO.INPUT.READ_LINE expected 0 arguments, got %d", len(args))
	}

	return stdIOReadLine()
}

func builtinStdIOInputPrompt(_ *RuntimeContext, args []Value) (Value, error) {
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
