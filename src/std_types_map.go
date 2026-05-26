package main

import "fmt"

func NewStdTypesMapMap() Value {
	entries := map[string]Binding{
		"KEYS":   NewImmutableBinding(NewBuiltinFunctionValue(stdTypesMapKeys)),
		"VALUES": NewImmutableBinding(NewBuiltinFunctionValue(stdTypesMapValues)),
	}

	return NewMapValue(entries, true)
}

func stdTypesMapKeys(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.MAP.KEYS expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPES.MAP.KEYS cannot inspect invalid map")
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
		return Value{}, fmt.Errorf("TYPES.MAP.KEYS cannot inspect unknown value kind")
	}
}

func stdTypesMapValues(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.MAP.VALUES expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPES.MAP.VALUES cannot inspect invalid map")
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
		return Value{}, fmt.Errorf("TYPES.MAP.VALUES cannot inspect unknown value kind")
	}
}
