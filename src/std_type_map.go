package main

import "fmt"

func NewStdTypeMapMap() Value {
	entries := map[string]Binding{
		"ENTRIES": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapEntries)),
		"KEYS":    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapKeys)),
		"MERGE":   NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapMerge)),
		"VALUES":  NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapValues)),
	}

	return NewMapValue(entries, true)
}

func StdTypeMapMerge(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return Value{}, fmt.Errorf("TYPE.MAP.MERGE expected at least 1 argument, got 0")
	}

	entries := make(map[string]Binding)

	for index, arg := range args {
		value := resolveSpecializedValue(arg)
		if value.Kind != ValueMap {
			return Value{}, fmt.Errorf("TYPE.MAP.MERGE argument %d must be a map", index+1)
		}

		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.MAP.MERGE cannot merge invalid map")
		}

		for key, binding := range value.Map.Entries {
			entries[key] = binding
		}
	}

	return NewMapValue(entries, false), nil
}

func StdTypeMapKeys(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.MAP.KEYS expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.MAP.KEYS cannot inspect invalid map")
		}

		keys := sortedMapKeys(value.Map)
		elements := make([]Value, 0, len(keys))

		for _, key := range keys {
			elements = append(elements, NewStringValue(key))
		}

		return NewArrayValue(elements, false), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueString,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.MAP.KEYS cannot inspect unknown value kind")
	}
}

func StdTypeMapEntries(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.MAP.ENTRIES expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.MAP.ENTRIES cannot inspect invalid map")
		}

		keys := sortedMapKeys(value.Map)
		elements := make([]Value, 0, len(keys))

		for _, key := range keys {
			pair := NewArrayValue([]Value{
				NewStringValue(key),
				value.Map.Entries[key].Value,
			}, false)

			elements = append(elements, pair)
		}

		return NewArrayValue(elements, false), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueString,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.MAP.ENTRIES cannot inspect unknown value kind")
	}
}

func StdTypeMapValues(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.MAP.VALUES expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.MAP.VALUES cannot inspect invalid map")
		}

		keys := sortedMapKeys(value.Map)
		elements := make([]Value, 0, len(keys))

		for _, key := range keys {
			elements = append(elements, value.Map.Entries[key].Value)
		}

		return NewArrayValue(elements, false), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueString,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.MAP.VALUES cannot inspect unknown value kind")
	}
}
