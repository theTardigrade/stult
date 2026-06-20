package main

import (
	"fmt"
	"strings"
)

func NewStdTypeStringMap() Value {
	entries := map[string]Binding{
		"CHARACTERS":        NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringCharacters)),
		"IS_FOUND_AT_END":   NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringIsFoundAtEnd)),
		"IS_FOUND_AT_START": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringIsFoundAtStart)),
		"IS_FOUND_IN":       NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringIsFoundIn)),
		"JOIN":              NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringJoin)),
		"NEW":               NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringNew)),
		"REPLACE":           NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringReplace)),
		"SPLIT":             NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringSplit)),
		"TO_LOWER":          NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringToLower)),
		"TO_UPPER":          NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringToUpper)),
		"TRIM":              NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringTrim)),
		"TRIM_END":          NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringTrimEnd)),
		"TRIM_START":        NewImmutableBinding(NewBuiltinFunctionValue(StdTypeStringTrimStart)),
	}

	return NewMapValue(entries, true)
}

func StdTypeStringNew(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.STRING.NEW expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueString:
		if value.Text == nil {
			return NewStringValue(""), nil
		}

		return NewStringValue(value.Text.String()), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueMap,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return NewStringValue(value.String()), nil

	default:
		return Value{}, fmt.Errorf("TYPE.STRING.NEW cannot convert unknown value kind")
	}
}

func StdTypeStringCharacters(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.STRING.CHARACTERS expected 1 argument, got %d", len(args))
	}

	text, err := StdTypeStringArg("TYPE.STRING.CHARACTERS", args[0], 1)
	if err != nil {
		return NewVoidValue(), nil
	}

	elements := make([]Value, 0, len([]rune(text)))

	for _, ch := range text {
		elements = append(elements, NewStringValue(string(ch)))
	}

	return NewArrayValue(elements, false), nil
}

func StdTypeStringTrim(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.STRING.TRIM expected 1 argument, got %d", len(args))
	}

	text, err := StdTypeStringArg("TYPE.STRING.TRIM", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.TrimSpace(text)), nil
}

func StdTypeStringTrimStart(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.STRING.TRIM_START expected 1 argument, got %d", len(args))
	}

	text, err := StdTypeStringArg("TYPE.STRING.TRIM_START", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.TrimLeftFunc(text, func(ch rune) bool {
		return strings.ContainsRune(" \t\n\r\v\f", ch)
	})), nil
}

func StdTypeStringTrimEnd(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.STRING.TRIM_END expected 1 argument, got %d", len(args))
	}

	text, err := StdTypeStringArg("TYPE.STRING.TRIM_END", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.TrimRightFunc(text, func(ch rune) bool {
		return strings.ContainsRune(" \t\n\r\v\f", ch)
	})), nil
}

func StdTypeStringToLower(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.STRING.TO_LOWER expected 1 argument, got %d", len(args))
	}

	text, err := StdTypeStringArg("TYPE.STRING.TO_LOWER", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.ToLower(text)), nil
}

func StdTypeStringToUpper(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.STRING.TO_UPPER expected 1 argument, got %d", len(args))
	}

	text, err := StdTypeStringArg("TYPE.STRING.TO_UPPER", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.ToUpper(text)), nil
}

func StdTypeStringIsFoundIn(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPE.STRING.IS_FOUND_IN expected 2 arguments, got %d", len(args))
	}

	search, text, err := StdTypeStringSearchArgs("TYPE.STRING.IS_FOUND_IN", args)
	if err != nil {
		return Value{}, err
	}

	return NewBoolValue(strings.Contains(text, search)), nil
}

func StdTypeStringIsFoundAtStart(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPE.STRING.IS_FOUND_AT_START expected 2 arguments, got %d", len(args))
	}

	search, text, err := StdTypeStringSearchArgs("TYPE.STRING.IS_FOUND_AT_START", args)
	if err != nil {
		return Value{}, err
	}

	return NewBoolValue(strings.HasPrefix(text, search)), nil
}

func StdTypeStringIsFoundAtEnd(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPE.STRING.IS_FOUND_AT_END expected 2 arguments, got %d", len(args))
	}

	search, text, err := StdTypeStringSearchArgs("TYPE.STRING.IS_FOUND_AT_END", args)
	if err != nil {
		return Value{}, err
	}

	return NewBoolValue(strings.HasSuffix(text, search)), nil
}

func StdTypeStringReplace(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 3 {
		return Value{}, fmt.Errorf("TYPE.STRING.REPLACE expected 3 arguments, got %d", len(args))
	}

	text, err := StdTypeStringArg("TYPE.STRING.REPLACE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	oldText, err := StdTypeStringArg("TYPE.STRING.REPLACE", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	newText, err := StdTypeStringArg("TYPE.STRING.REPLACE", args[2], 3)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.ReplaceAll(text, oldText, newText)), nil
}

func StdTypeStringSplit(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPE.STRING.SPLIT expected 2 arguments, got %d", len(args))
	}

	text, err := StdTypeStringArg("TYPE.STRING.SPLIT", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	separator, err := StdTypeStringArg("TYPE.STRING.SPLIT", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	parts := strings.Split(text, separator)
	elements := make([]Value, 0, len(parts))

	for _, part := range parts {
		elements = append(elements, NewStringValue(part))
	}

	return NewArrayValue(elements, false), nil
}

func StdTypeStringJoin(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPE.STRING.JOIN expected 2 arguments, got %d", len(args))
	}

	array := resolveSpecializedValue(args[0])
	if array.Kind != ValueArray || array.Array == nil {
		return Value{}, fmt.Errorf("TYPE.STRING.JOIN argument 1 expected an array")
	}

	separator, err := StdTypeStringArg("TYPE.STRING.JOIN", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	parts := make([]string, 0, len(array.Array.Ordinary))

	if err := array.Array.ForEach(func(_ *Number, element Value) error {
		value := resolveSpecializedValue(element)

		if value.Kind == ValueString {
			if value.Text == nil {
				parts = append(parts, "")
			} else {
				parts = append(parts, value.Text.String())
			}

			return nil
		}

		parts = append(parts, value.PrintString())
		return nil
	}); err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.Join(parts, separator)), nil
}

func StdTypeStringArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string", name, position)
	}

	return value.Text.String(), nil
}

func StdTypeStringSearchArgs(name string, args []Value) (string, string, error) {
	search, err := StdTypeStringArg(name, args[0], 1)
	if err != nil {
		return "", "", err
	}

	text, err := StdTypeStringArg(name, args[1], 2)
	if err != nil {
		return "", "", err
	}

	return search, text, nil
}
