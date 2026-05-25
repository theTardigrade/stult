package main

import (
	"fmt"
	"unicode/utf8"
)

func NewStdTypesMap() Value {
	entries := map[string]Binding{
		"ARRAY":  NewImmutableBinding(NewStdTypesArrayMap()),
		"BOOL":   NewImmutableBinding(NewStdTypesBoolMap()),
		"NUMBER": NewImmutableBinding(NewStdTypesNumberMap()),

		"SIZE": NewImmutableBinding(NewBuiltinFunctionValue(stdTypesSize)),
	}

	return NewMapValue(entries, true)
}

func stdTypesSize(interpreter *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.SIZE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		return NewNumberValueFromInt(len(value.Map.Entries)), nil

	case ValueArray:
		return NewNumberValueFromInt(len(value.Array.Elements)), nil

	case ValueEmptyCollection:
		return NewNumberValueFromInt(0), nil

	case ValueString:
		return NewNumberValueFromInt(utf8.RuneCountInString(value.Text)), nil

	case ValueEmpty,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewEmptyValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPES.SIZE cannot determine size of unknown value kind")
	}
}
