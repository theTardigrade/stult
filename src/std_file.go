package main

import (
	"fmt"
	"io"
	"os"
)

func NewStdFileMap() Value {
	entries := map[string]Binding{
		"APPEND": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileAppend)),
		"COPY":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileCopy)),
		"DELETE": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileDelete)),
		"EXISTS": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileExists)),
		"PATH":   NewImmutableBinding(NewStdFilePathMap()),
		"READ":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileRead)),
		"RENAME": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileRename)),
		"SIZE":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileSize)),
		"WRITE":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileWrite)),
	}

	return NewMapValue(entries, true)
}

func builtinStdFileRead(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.READ expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.READ", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(string(data)), nil
}

func builtinStdFileWrite(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("FILE.WRITE expected 2 arguments, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.WRITE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	content, err := stdFileContentArg("FILE.WRITE", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return Value{}, err
	}

	return NewVoidValue(), nil
}

func builtinStdFileAppend(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("FILE.APPEND expected 2 arguments, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.APPEND", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	content, err := stdFileContentArg("FILE.APPEND", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return Value{}, err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return Value{}, err
	}

	return NewVoidValue(), nil
}

func builtinStdFileExists(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.EXISTS expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.EXISTS", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	_, err = os.Stat(path)
	if err == nil {
		return NewBoolValue(true), nil
	}

	if os.IsNotExist(err) {
		return NewBoolValue(false), nil
	}

	return Value{}, err
}

func builtinStdFileDelete(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.DELETE expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.DELETE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	if err := os.Remove(path); err != nil {
		return Value{}, err
	}

	return NewVoidValue(), nil
}

func builtinStdFileRename(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("FILE.RENAME expected 2 arguments, got %d", len(args))
	}

	oldPath, err := stdFilePathArg("FILE.RENAME", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	newPath, err := stdFilePathArg("FILE.RENAME", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return Value{}, err
	}

	return NewVoidValue(), nil
}

func builtinStdFileCopy(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("FILE.COPY expected 2 arguments, got %d", len(args))
	}

	sourcePath, err := stdFilePathArg("FILE.COPY", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	destinationPath, err := stdFilePathArg("FILE.COPY", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return Value{}, err
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return Value{}, err
	}
	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return Value{}, err
	}
	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return Value{}, err
	}

	return NewVoidValue(), nil
}

func builtinStdFileSize(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.SIZE expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.SIZE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return Value{}, err
	}

	return NewNumberValueFromInt64(info.Size()), nil
}

func stdFilePathArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string path", name, position)
	}

	return value.Text.String(), nil
}

func stdFileContentArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	switch value.Kind {
	case ValueString:
		if value.Text == nil {
			return "", fmt.Errorf("%s argument %d expected valid content", name, position)
		}

		return value.Text.String(), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueMap,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return value.PrintString(), nil

	default:
		return "", fmt.Errorf("%s argument %d has unsupported content type", name, position)
	}
}
