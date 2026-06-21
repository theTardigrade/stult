package main

import (
	"fmt"
	"path/filepath"
)

func NewStdFilePathMap() Value {
	entries := map[string]Binding{
		"ABS":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFilePathAbs)),
		"BASE":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFilePathBase)),
		"CLEAN": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFilePathClean)),
		"DIR":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFilePathDir)),
		"EXT":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFilePathExt)),
		"JOIN":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFilePathJoin)),
	}

	return NewMapValue(entries, true)
}

func builtinStdFilePathJoin(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return Value{}, fmt.Errorf("FILE.PATH.JOIN expected at least 1 argument, got 0")
	}

	parts := make([]string, 0, len(args))

	for index, arg := range args {
		part, err := stdFilePathStringArg("FILE.PATH.JOIN", arg, index+1)
		if err != nil {
			return Value{}, err
		}

		parts = append(parts, part)
	}

	return NewStringValue(filepath.Join(parts...)), nil
}

func builtinStdFilePathBase(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.PATH.BASE expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathStringArg("FILE.PATH.BASE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(filepath.Base(path)), nil
}

func builtinStdFilePathDir(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.PATH.DIR expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathStringArg("FILE.PATH.DIR", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(filepath.Dir(path)), nil
}

func builtinStdFilePathExt(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.PATH.EXT expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathStringArg("FILE.PATH.EXT", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(filepath.Ext(path)), nil
}

func builtinStdFilePathClean(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.PATH.CLEAN expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathStringArg("FILE.PATH.CLEAN", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(filepath.Clean(path)), nil
}

func builtinStdFilePathAbs(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.PATH.ABS expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathStringArg("FILE.PATH.ABS", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(absolutePath), nil
}

func stdFilePathStringArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string path", name, position)
	}

	return value.Text.String(), nil
}
