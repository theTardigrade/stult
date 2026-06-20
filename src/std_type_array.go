package main

import "fmt"

func NewStdTypeArrayMap() Value {
	entries := map[string]Binding{
		"APPEND": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeArrayAppend)),
	}

	return NewMapValue(entries, true)
}

func StdTypeArrayAppend(_ *RuntimeContext, args []Value) (Value, error) {
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

	if target.Array == nil {
		return fmt.Errorf("TYPE.ARRAY.APPEND cannot append to invalid array")
	}

	if target.Array.IsFrozen {
		return fmt.Errorf("cannot modify frozen array")
	}

	return target.Array.Append(value)
}
