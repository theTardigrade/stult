package main

import "fmt"

func NewStdTypesStringMap() Value {
	entries := map[string]Binding{
		"NEW": NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringNew)),
	}

	return NewMapValue(entries, true)
}

func stdTypesStringNew(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.STRING.NEW expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueString:
		if value.Text == nil {
			return NewStringValue(""), nil
		}

		return NewStringValue(value.Text.String()), nil

	case ValueVoid,
		ValueEmptyCollection,
		ValueNumber,
		ValueBool,
		ValueMap,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return NewStringValue(value.String()), nil

	default:
		return Value{}, fmt.Errorf("TYPES.STRING.NEW cannot convert unknown value kind")
	}
}
