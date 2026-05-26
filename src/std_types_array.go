package main

import "fmt"

func NewStdTypesArrayMap() Value {
	entries := map[string]Binding{
		"APPEND": NewImmutableBinding(NewBuiltinFunctionValue(stdTypesArrayAppend)),
	}

	return NewMapValue(entries, true)
}

func stdTypesArrayAppend(interpreter *Interpreter, args []Value) (Value, error) {
	if len(args) < 2 {
		return Value{}, fmt.Errorf("TYPES.ARRAY.APPEND expected at least 2 arguments, got %d", len(args))
	}

	target := resolveSpecializedValue(args[0])

	for _, value := range args[1:] {
		if err := appendArrayValue(target, value); err != nil {
			return Value{}, err
		}

		target = resolveSpecializedValue(target)
	}

	return NewVoidValue(), nil
}

func appendArrayValue(target Value, value Value) error {
	switch target.Kind {
	case ValueArray:
		_, err := assignArrayIndex(
			target.Array,
			NewNumberValueFromInt(len(target.Array.Elements)),
			value,
		)
		return err

	case ValueEmptyCollection:
		_, err := specializeEmptyCollection(
			target,
			NewNumberValueFromInt(0),
			value,
		)
		return err

	default:
		return fmt.Errorf("TYPES.ARRAY.APPEND expected an array or empty collection")
	}
}
