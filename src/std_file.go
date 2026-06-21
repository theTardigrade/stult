package main

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"unicode/utf8"
)

func NewStdFileMap() Value {
	entries := map[string]Binding{
		"COPY":     NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileCopy)),
		"DELETE":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileDelete)),
		"EXISTS":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileExists)),
		"IS_DIR":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileIsDir)),
		"IS_FILE":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileIsFile)),
		"LIST":     NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileList)),
		"MAKE_DIR": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileMakeDir)),
		"PATH":     NewImmutableBinding(NewStdFilePathMap()),
		"READ":     NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileRead)),
		"RENAME":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileRename)),
		"SIZE":     NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileSize)),
		"WRITE":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileWrite)),
	}

	return NewMapValue(entries, true)
}

func builtinStdFileRead(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 4 {
		return Value{}, fmt.Errorf("FILE.READ expected 1 to 4 arguments, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.READ", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	useBytes, err := stdFileOptionalBoolArg("FILE.READ", args, 2, false)
	if err != nil {
		return Value{}, err
	}

	var offset *big.Int
	if len(args) >= 3 {
		offset, err = stdFileNonNegativeWholeNumberArg("FILE.READ", args[2], 3)
		if err != nil {
			return Value{}, err
		}
	} else {
		offset = new(big.Int)
	}

	var length *big.Int
	if len(args) >= 4 {
		length, err = stdFileNonNegativeWholeNumberArg("FILE.READ", args[3], 4)
		if err != nil {
			return Value{}, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Value{}, err
	}

	if useBytes {
		return stdFileReadBytes(data, offset, length), nil
	}

	return stdFileReadText(data, offset, length)
}

func builtinStdFileWrite(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return Value{}, fmt.Errorf("FILE.WRITE expected 2 or 3 arguments, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.WRITE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	content, err := stdFileContentBytesArg("FILE.WRITE", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	appendMode, err := stdFileOptionalBoolArg("FILE.WRITE", args, 3, false)
	if err != nil {
		return Value{}, err
	}

	if appendMode {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return Value{}, err
		}
		defer file.Close()

		if _, err := file.Write(content); err != nil {
			return Value{}, err
		}

		return NewVoidValue(), nil
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
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

func builtinStdFileIsFile(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.IS_FILE expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.IS_FILE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	info, err := os.Stat(path)
	if err == nil {
		return NewBoolValue(info.Mode().IsRegular()), nil
	}

	if os.IsNotExist(err) {
		return NewBoolValue(false), nil
	}

	return Value{}, err
}

func builtinStdFileIsDir(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("FILE.IS_DIR expected 1 argument, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.IS_DIR", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	info, err := os.Stat(path)
	if err == nil {
		return NewBoolValue(info.IsDir()), nil
	}

	if os.IsNotExist(err) {
		return NewBoolValue(false), nil
	}

	return Value{}, err
}

func builtinStdFileList(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return Value{}, fmt.Errorf("FILE.LIST expected 1 or 2 arguments, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.LIST", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	absolute, err := stdFileOptionalBoolArg("FILE.LIST", args, 2, false)
	if err != nil {
		return Value{}, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return Value{}, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	basePath := path
	if absolute {
		basePath, err = filepath.Abs(path)
		if err != nil {
			return Value{}, err
		}
	}

	values := make([]Value, 0, len(names))
	for _, name := range names {
		value := name
		if absolute {
			value = filepath.Join(basePath, name)
		}

		values = append(values, NewStringValue(value))
	}

	return NewArrayValue(values, false), nil
}

func builtinStdFileMakeDir(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return Value{}, fmt.Errorf("FILE.MAKE_DIR expected 1 or 2 arguments, got %d", len(args))
	}

	path, err := stdFilePathArg("FILE.MAKE_DIR", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	recursive, err := stdFileOptionalBoolArg("FILE.MAKE_DIR", args, 2, false)
	if err != nil {
		return Value{}, err
	}

	if recursive {
		if err := os.MkdirAll(path, 0755); err != nil {
			return Value{}, err
		}

		info, err := os.Stat(path)
		if err != nil {
			return Value{}, err
		}

		if !info.IsDir() {
			return Value{}, fmt.Errorf("FILE.MAKE_DIR path exists and is not a directory")
		}

		return NewVoidValue(), nil
	}

	if err := os.Mkdir(path, 0755); err != nil {
		return Value{}, err
	}

	return NewVoidValue(), nil
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

func stdFileOptionalBoolArg(name string, args []Value, position int, defaultValue bool) (bool, error) {
	if len(args) < position {
		return defaultValue, nil
	}

	value := resolveSpecializedValue(args[position-1])
	if value.Kind != ValueBool {
		return false, fmt.Errorf("%s argument %d expected a boolean", name, position)
	}

	return value.Bool, nil
}

func stdFileNonNegativeWholeNumberArg(name string, arg Value, position int) (*big.Int, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind != ValueNumber || value.Number == nil {
		return nil, fmt.Errorf("%s argument %d expected a non-negative whole number", name, position)
	}

	integer, accuracy := value.Number.Int(nil)
	if accuracy != big.Exact || integer.Sign() < 0 {
		return nil, fmt.Errorf("%s argument %d expected a non-negative whole number", name, position)
	}

	return integer, nil
}

func stdFileReadText(data []byte, offset *big.Int, length *big.Int) (Value, error) {
	if !utf8.Valid(data) {
		return Value{}, fmt.Errorf("FILE.READ text mode expected valid UTF-8")
	}

	runes := []rune(string(data))
	start, end := stdFileSliceRange(len(runes), offset, length)
	return NewStringValue(string(runes[start:end])), nil
}

func stdFileReadBytes(data []byte, offset *big.Int, length *big.Int) Value {
	start, end := stdFileSliceRange(len(data), offset, length)
	values := make([]Value, 0, end-start)
	for _, b := range data[start:end] {
		values = append(values, NewNumberValueFromInt64(int64(b)))
	}

	return NewArrayValue(values, false)
}

func stdFileSliceRange(total int, offset *big.Int, length *big.Int) (int, int) {
	if offset == nil {
		offset = new(big.Int)
	}

	totalBig := big.NewInt(int64(total))
	if offset.Cmp(totalBig) >= 0 {
		return total, total
	}

	start := int(offset.Int64())
	if length == nil {
		return start, total
	}

	if length.Sign() == 0 {
		return start, start
	}

	remaining := total - start
	remainingBig := big.NewInt(int64(remaining))
	if length.Cmp(remainingBig) >= 0 {
		return start, total
	}

	end := start + int(length.Int64())
	return start, end
}

func stdFileContentBytesArg(name string, arg Value, position int) ([]byte, error) {
	value := resolveSpecializedValue(arg)

	switch value.Kind {
	case ValueString:
		if value.Text == nil {
			return nil, fmt.Errorf("%s argument %d expected valid content", name, position)
		}

		return []byte(value.Text.String()), nil

	case ValueArray:
		return stdFileByteArrayArg(name, value, position)

	default:
		return nil, fmt.Errorf("%s argument %d has unsupported content type", name, position)
	}
}

func stdFileByteArrayArg(name string, value Value, position int) ([]byte, error) {
	if value.Array == nil {
		return nil, fmt.Errorf("%s argument %d expected a valid byte array", name, position)
	}

	bytes := make([]byte, 0)
	err := value.Array.ForEach(func(index *Number, item Value) error {
		item = resolveSpecializedValue(item)
		if item.Kind != ValueNumber || item.Number == nil {
			return fmt.Errorf("%s argument %d byte array item at index %s expected a number from 0 to 255", name, position, index.String())
		}

		integer, accuracy := item.Number.Int64()
		if accuracy != big.Exact || integer < 0 || integer > 255 {
			return fmt.Errorf("%s argument %d byte array item at index %s expected a whole number from 0 to 255", name, position, index.String())
		}

		bytes = append(bytes, byte(integer))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
