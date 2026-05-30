package main

import (
	"fmt"
	"os"
)

func NewStdSystemMap(args []string) Value {
	entries := map[string]Binding{
		"ARGS": NewImmutableBinding(stdSystemArgsValue(args)),
		"CWD":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdSystemCWD)),
		"ENV":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdSystemEnv)),
		"EXIT": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdSystemExit)),
	}

	return NewMapValue(entries, true)
}

func stdSystemArgsValue(args []string) Value {
	elements := make([]Value, 0, len(args))

	for _, arg := range args {
		elements = append(elements, NewStringValue(arg))
	}

	return NewArrayValue(elements, true)
}

func builtinStdSystemCWD(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 0 {
		return Value{}, fmt.Errorf("SYSTEM.CWD expected 0 arguments, got %d", len(args))
	}

	path, err := os.Getwd()
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(path), nil
}

func builtinStdSystemEnv(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("SYSTEM.ENV expected 1 argument, got %d", len(args))
	}

	name, err := stdSystemStringArg("SYSTEM.ENV", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	value, ok := os.LookupEnv(name)
	if !ok {
		return NewVoidValue(), nil
	}

	return NewStringValue(value), nil
}

func builtinStdSystemExit(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("SYSTEM.EXIT expected 1 argument, got %d", len(args))
	}

	code, err := stdSystemExitCodeArg(args[0], 1)
	if err != nil {
		return Value{}, err
	}

	os.Exit(code)

	return NewVoidValue(), nil
}

func stdSystemStringArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string", name, position)
	}

	return value.Text.String(), nil
}

func stdSystemExitCodeArg(arg Value, position int) (int, error) {
	code, err := numberToInt64(arg, fmt.Sprintf("SYSTEM.EXIT argument %d", position))
	if err != nil {
		return 0, err
	}

	if code < 0 {
		return 0, fmt.Errorf("SYSTEM.EXIT argument %d cannot be negative", position)
	}

	if code > 255 {
		return 0, fmt.Errorf("SYSTEM.EXIT argument %d cannot be greater than 255", position)
	}

	return int(code), nil
}
