package main

import (
	"fmt"
	"path/filepath"
)

func NewStdPathMap() Value {
	entries := map[string]Binding{
		"ABS":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdPathAbs)),
		"BASE":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdPathBase)),
		"CLEAN": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdPathClean)),
		"DIR":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdPathDir)),
		"EXT":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdPathExt)),
		"JOIN":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdPathJoin)),
	}

	return NewMapValue(entries, true)
}

func builtinStdPathJoin(_ *Interpreter, args []Value) (Value, error) {
	if len(args) == 0 {
		return Value{}, fmt.Errorf("PATH.JOIN expected at least 1 argument, got 0")
	}

	parts := make([]string, 0, len(args))

	for index, arg := range args {
		part, err := stdPathArg("PATH.JOIN", arg, index+1)
		if err != nil {
			return Value{}, err
		}

		parts = append(parts, part)
	}

	return NewStringValue(filepath.Join(parts...)), nil
}

func builtinStdPathBase(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("PATH.BASE expected 1 argument, got %d", len(args))
	}

	path, err := stdPathArg("PATH.BASE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(filepath.Base(path)), nil
}

func builtinStdPathDir(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("PATH.DIR expected 1 argument, got %d", len(args))
	}

	path, err := stdPathArg("PATH.DIR", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(filepath.Dir(path)), nil
}

func builtinStdPathExt(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("PATH.EXT expected 1 argument, got %d", len(args))
	}

	path, err := stdPathArg("PATH.EXT", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(filepath.Ext(path)), nil
}

func builtinStdPathClean(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("PATH.CLEAN expected 1 argument, got %d", len(args))
	}

	path, err := stdPathArg("PATH.CLEAN", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(filepath.Clean(path)), nil
}

func builtinStdPathAbs(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("PATH.ABS expected 1 argument, got %d", len(args))
	}

	path, err := stdPathArg("PATH.ABS", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(absolutePath), nil
}

func stdPathArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string path", name, position)
	}

	return value.Text.String(), nil
}
