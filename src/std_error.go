package main

import "fmt"

func NewStdErrorMap() Value {
	entries := map[string]Binding{
		"ASSERT": NewImmutableBinding(NewStdErrorAssertMap()),
		"RAISE":  NewImmutableBinding(NewBuiltinFunctionValue(StdErrorRaise)),
	}

	return NewMapValue(entries, true)
}

func StdErrorRaise(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) > 1 {
		return Value{}, fmt.Errorf("ERROR.RAISE expected 0 or 1 arguments, got %d", len(args))
	}

	message := "error raised"

	if len(args) == 1 {
		value := resolveSpecializedValue(args[0])
		if value.Kind != ValueString {
			return Value{}, fmt.Errorf("ERROR.RAISE argument 1 expected a string message")
		}

		if value.Text == nil {
			return Value{}, fmt.Errorf("ERROR.RAISE argument 1 expected a valid string message")
		}

		message = value.Text.String()
	}

	return Value{}, fmt.Errorf("%s", message)
}
