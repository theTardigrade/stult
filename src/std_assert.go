package main

import "fmt"

func NewStdAssertMap() Value {
	entries := map[string]Binding{
		"EQUAL": NewImmutableBinding(NewBuiltinFunctionValue(StdAssertEqual)),
		"FALSE": NewImmutableBinding(NewBuiltinFunctionValue(StdAssertFalse)),
		"TRUE":  NewImmutableBinding(NewBuiltinFunctionValue(StdAssertTrue)),
	}

	return NewMapValue(entries, true)
}

func StdAssertTrue(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return Value{}, fmt.Errorf("ERROR.ASSERT.TRUE expected 1 or 2 arguments, got %d", len(args))
	}

	condition := resolveSpecializedValue(args[0])
	if condition.Kind != ValueBool {
		return Value{}, fmt.Errorf("ERROR.ASSERT.TRUE argument 1 expected a boolean")
	}

	message, err := stdAssertOptionalMessageArg("ERROR.ASSERT.TRUE", args, 2, "expected condition to be true")
	if err != nil {
		return Value{}, err
	}

	if !condition.Bool {
		return Value{}, fmt.Errorf("assertion failed: %s", message)
	}

	return NewVoidValue(), nil
}

func StdAssertFalse(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return Value{}, fmt.Errorf("ERROR.ASSERT.FALSE expected 1 or 2 arguments, got %d", len(args))
	}

	condition := resolveSpecializedValue(args[0])
	if condition.Kind != ValueBool {
		return Value{}, fmt.Errorf("ERROR.ASSERT.FALSE argument 1 expected a boolean")
	}

	message, err := stdAssertOptionalMessageArg("ERROR.ASSERT.FALSE", args, 2, "expected condition to be false")
	if err != nil {
		return Value{}, err
	}

	if condition.Bool {
		return Value{}, fmt.Errorf("assertion failed: %s", message)
	}

	return NewVoidValue(), nil
}

func StdAssertEqual(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return Value{}, fmt.Errorf("ERROR.ASSERT.EQUAL expected 2 or 3 arguments, got %d", len(args))
	}

	actual := resolveSpecializedValue(args[0])
	expected := resolveSpecializedValue(args[1])

	message, err := stdAssertOptionalMessageArg("ERROR.ASSERT.EQUAL", args, 3, "expected values to be equal")
	if err != nil {
		return Value{}, err
	}

	ok, err := valuesEqual(actual, expected)
	if err != nil {
		return Value{}, fmt.Errorf("ERROR.ASSERT.EQUAL cannot compare values: %w", err)
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

func stdAssertOptionalMessageArg(name string, args []Value, position int, defaultMessage string) (string, error) {
	if len(args) < position {
		return defaultMessage, nil
	}

	value := resolveSpecializedValue(args[position-1])

	if value.Kind != ValueString {
		return "", fmt.Errorf("%s argument %d expected a string message", name, position)
	}

	if value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a valid string message", name, position)
	}

	return value.Text.String(), nil
}
