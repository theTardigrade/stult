package main

import (
	"fmt"
	"strings"
)

func NewStdTypeBoolMap() Value {
	entries := map[string]Binding{
		"FALSE": NewImmutableBinding(NewBoolValue(false)),
		"TRUE":  NewImmutableBinding(NewBoolValue(true)),

		"NEW": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeBoolNew)),
	}

	return NewMapValue(entries, true)
}

func StdTypeBoolNew(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.BOOL.NEW expected 1 argument, got %d", len(args))
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
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.BOOL.NEW cannot convert unknown value kind")
	}
}
