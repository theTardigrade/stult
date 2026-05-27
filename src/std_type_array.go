package main

import "fmt"

func NewStdTypeArrayMap() Value {
	entries := map[string]Binding{
		"APPEND": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeArrayAppend)),
	}

	return NewMapValue(entries, true)
}

func StdTypeArrayAppend(_ *Interpreter, args []Value) (Value, error) {
	if len(args) < 2 {
		return Value{}, fmt.Errorf("TYPE.ARRAY.APPEND expected at least 2 arguments, got %d", len(args))
	}

	target := resolveSpecializedValue(args[0])

	if target.Kind != ValueArray {
		return Value{}, fmt.Errorf("TYPE.ARRAY.APPEND expected an array")
	}

	for _, value := range args[1:] {
		if err := appendArrayValue(target, value); err != nil {
			return Value{}, err
		}
	}

	return NewVoidValue(), nil
}

func appendArrayValue(target Value, value Value) error {
	if target.Kind != ValueArray {
		return fmt.Errorf("TYPE.ARRAY.APPEND expected an array")
	}

	_, err := assignArrayIndex(
		target.Array,
		NewNumberValueFromInt(len(target.Array.Elements)),
		value,
	)

	return err
}
