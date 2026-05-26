package main

import (
	"fmt"
	"strings"
)

func NewStdTypesBoolMap() Value {
	entries := map[string]Binding{
		"TRUE":  NewImmutableBinding(NewBoolValue(true)),
		"FALSE": NewImmutableBinding(NewBoolValue(false)),

		"NEW": NewImmutableBinding(NewBuiltinFunctionValue(stdTypesBoolNew)),
	}

	return NewMapValue(entries, true)
}

func stdTypesBoolNew(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.BOOL.NEW expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueBool:
		return value, nil

	case ValueNumber:
		return NewBoolValue(value.Number.Sign() != 0), nil

	case ValueString:
		if value.Text == nil {
			return NewVoidValue(), nil
		}

		text := strings.ToUpper(strings.TrimSpace(value.Text.String()))

		switch text {
		case "T", "TRUE", "1":
			return NewBoolValue(true), nil

		case "F", "FALSE", "0":
			return NewBoolValue(false), nil

		default:
			return NewVoidValue(), nil
		}

	case ValueVoid,
		ValueMap,
		ValueArray,
		ValueEmptyCollection,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPES.BOOL.NEW cannot convert unknown value kind")
	}
}
