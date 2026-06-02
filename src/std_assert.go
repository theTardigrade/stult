package main

import "fmt"

func NewStdAssertMap() Value {
	entries := map[string]Binding{
		"EQUAL": NewImmutableBinding(NewBuiltinFunctionValue(StdAssertEqual)),
		"TRUE":  NewImmutableBinding(NewBuiltinFunctionValue(StdAssertTrue)),
	}

	return NewMapValue(entries, true)
}

func StdAssertTrue(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("ASSERT.TRUE expected 2 arguments, got %d", len(args))
	}

	condition := resolveSpecializedValue(args[0])
	if condition.Kind != ValueBool {
		return Value{}, fmt.Errorf("ASSERT.TRUE argument 1 expected a boolean")
	}

	message, err := stdAssertMessageArg("ASSERT.TRUE", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	if !condition.Bool {
		return Value{}, fmt.Errorf("assertion failed: %s", message)
	}

	return NewVoidValue(), nil
}

func StdAssertEqual(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 3 {
		return Value{}, fmt.Errorf("ASSERT.EQUAL expected 3 arguments, got %d", len(args))
	}

	actual := resolveSpecializedValue(args[0])
	expected := resolveSpecializedValue(args[1])

	message, err := stdAssertMessageArg("ASSERT.EQUAL", args[2], 3)
	if err != nil {
		return Value{}, err
	}

	ok, err := valuesEqual(actual, expected)
	if err != nil {
		return Value{}, fmt.Errorf("ASSERT.EQUAL cannot compare values: %w", err)
	}

	if !ok {
		return Value{}, fmt.Errorf(
			"assertion failed: %s; expected %s, got %s",
			message,
			expected.DebugString(),
			actual.DebugString(),
		)
	}

	return NewVoidValue(), nil
}

func stdAssertMessageArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueString {
		return "", fmt.Errorf("%s argument %d expected a string message", name, position)
	}

	if value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a valid string message", name, position)
	}

	return value.Text.String(), nil
}
