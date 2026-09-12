package main

import (
	"fmt"
	"unicode/utf8"
)

func NewStdFileBundledMap(runtime *RuntimeContext) Value {
	if runtime == nil {
		runtime = NewRuntimeContext(nil)
	}

	entries := map[string]Binding{
		"EXISTS": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileBundledExists)),
		"KEYS":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileBundledKeys)),
		"LIST":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileBundledList)),
		"READ":   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdFileBundledRead)),
	}

	return NewMapValue(entries, true)
}

func builtinStdFileBundledRead(runtime *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 3 {
		return Value{}, fmt.Errorf("FILE.BUNDLED.READ expected 1 to 3 arguments, got %d", len(args))
	}

	store := stdFileBundledAssetStore(runtime)
	name, err := stdFileBundledStringArg("FILE.BUNDLED.READ", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	relativePath, hasRelativePath, err := stdFileBundledOptionalPathArg("FILE.BUNDLED.READ", args, 2)
	if err != nil {
		return Value{}, err
	}

	useBytes, err := stdFileBundledOptionalBoolArg("FILE.BUNDLED.READ", args, 3, false)
	if err != nil {
		return Value{}, err
	}

	data, err := store.Read(name, relativePath, hasRelativePath)
	if err != nil {
		return Value{}, err
	}

	if useBytes {
		return stdFileBundledBytesValue(data), nil
	}

	return stdFileBundledTextValue(data)
}

func builtinStdFileBundledExists(runtime *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return Value{}, fmt.Errorf("FILE.BUNDLED.EXISTS expected 1 or 2 arguments, got %d", len(args))
	}

	store := stdFileBundledAssetStore(runtime)
	name, err := stdFileBundledStringArg("FILE.BUNDLED.EXISTS", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	relativePath, hasRelativePath, err := stdFileBundledOptionalPathArg("FILE.BUNDLED.EXISTS", args, 2)
	if err != nil {
		return Value{}, err
	}

	exists, err := store.Exists(name, relativePath, hasRelativePath)
	if err != nil {
		return Value{}, err
	}

	return NewBoolValue(exists), nil
}

func builtinStdFileBundledKeys(runtime *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return Value{}, fmt.Errorf("FILE.BUNDLED.KEYS expected 0 arguments, got %d", len(args))
	}

	store := stdFileBundledAssetStore(runtime)
	return stdFileBundledStringArrayValue(store.Keys()), nil
}

func builtinStdFileBundledList(runtime *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return Value{}, fmt.Errorf("FILE.BUNDLED.LIST expected 1 or 2 arguments, got %d", len(args))
	}

	store := stdFileBundledAssetStore(runtime)
	name, err := stdFileBundledStringArg("FILE.BUNDLED.LIST", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	relativePath, hasRelativePath, err := stdFileBundledOptionalPathArg("FILE.BUNDLED.LIST", args, 2)
	if err != nil {
		return Value{}, err
	}

	names, err := store.List(name, relativePath, hasRelativePath)
	if err != nil {
		return Value{}, err
	}

	return stdFileBundledStringArrayValue(names), nil
}

func stdFileBundledAssetStore(runtime *RuntimeContext) *BundleAssetStore {
	if runtime == nil || runtime.BundleAssets == nil {
		return NewEmptyBundleAssetStore()
	}

	return runtime.BundleAssets
}

func stdFileBundledStringArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string", name, position)
	}

	return value.Text.String(), nil
}

func stdFileBundledOptionalPathArg(name string, args []Value, position int) (string, bool, error) {
	if len(args) < position {
		return "", false, nil
	}

	value := resolveSpecializedValue(args[position-1])
	if value.Kind == ValueVoid {
		return "", false, nil
	}
	if value.Kind != ValueString || value.Text == nil {
		return "", false, fmt.Errorf("%s argument %d expected a string path or _", name, position)
	}

	return value.Text.String(), true, nil
}

func stdFileBundledOptionalBoolArg(name string, args []Value, position int, defaultValue bool) (bool, error) {
	if len(args) < position {
		return defaultValue, nil
	}

	value := resolveSpecializedValue(args[position-1])
	if value.Kind == ValueVoid {
		return defaultValue, nil
	}
	if value.Kind != ValueBool {
		return false, fmt.Errorf("%s argument %d expected a boolean or _", name, position)
	}

	return value.Bool, nil
}

func stdFileBundledTextValue(data []byte) (Value, error) {
	if !utf8.Valid(data) {
		return Value{}, fmt.Errorf("FILE.BUNDLED.READ text mode expected valid UTF-8")
	}

	return NewStringValue(string(data)), nil
}

func stdFileBundledBytesValue(data []byte) Value {
	values := make([]Value, 0, len(data))
	for _, b := range data {
		values = append(values, NewNumberValueFromInt64(int64(b)))
	}

	return NewArrayValue(values, false)
}

func stdFileBundledStringArrayValue(names []string) Value {
	values := make([]Value, 0, len(names))
	for _, name := range names {
		values = append(values, NewStringValue(name))
	}

	return NewArrayValue(values, false)
}
