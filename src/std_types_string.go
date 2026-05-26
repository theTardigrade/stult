package main

import (
	"fmt"
	"strings"
)

func NewStdTypesStringMap() Value {
	entries := map[string]Binding{
		"CHARACTERS":        NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringCharacters)),
		"IS_FOUND_AT_END":   NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringIsFoundAtEnd)),
		"IS_FOUND_AT_START": NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringIsFoundAtStart)),
		"IS_FOUND_IN":       NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringIsFoundIn)),
		"JOIN":              NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringJoin)),
		"NEW":               NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringNew)),
		"REPLACE":           NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringReplace)),
		"SPLIT":             NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringSplit)),
		"TO_LOWER":          NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringToLower)),
		"TO_UPPER":          NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringToUpper)),
		"TRIM":              NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringTrim)),
		"TRIM_END":          NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringTrimEnd)),
		"TRIM_START":        NewImmutableBinding(NewBuiltinFunctionValue(stdTypesStringTrimStart)),
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

func stdTypesStringCharacters(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.STRING.CHARACTERS expected 1 argument, got %d", len(args))
	}

	text, err := stdTypesStringArg("TYPES.STRING.CHARACTERS", args[0], 1)
	if err != nil {
		return NewVoidValue(), nil
	}

	elements := make([]Value, 0, len([]rune(text)))

	for _, ch := range text {
		elements = append(elements, NewStringValue(string(ch)))
	}

	return NewArrayValue(elements, false), nil
}

func stdTypesStringTrim(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.STRING.TRIM expected 1 argument, got %d", len(args))
	}

	text, err := stdTypesStringArg("TYPES.STRING.TRIM", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.TrimSpace(text)), nil
}

func stdTypesStringTrimStart(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.STRING.TRIM_START expected 1 argument, got %d", len(args))
	}

	text, err := stdTypesStringArg("TYPES.STRING.TRIM_START", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.TrimLeftFunc(text, func(ch rune) bool {
		return strings.ContainsRune(" \t\n\r\v\f", ch)
	})), nil
}

func stdTypesStringTrimEnd(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.STRING.TRIM_END expected 1 argument, got %d", len(args))
	}

	text, err := stdTypesStringArg("TYPES.STRING.TRIM_END", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.TrimRightFunc(text, func(ch rune) bool {
		return strings.ContainsRune(" \t\n\r\v\f", ch)
	})), nil
}

func stdTypesStringToLower(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.STRING.TO_LOWER expected 1 argument, got %d", len(args))
	}

	text, err := stdTypesStringArg("TYPES.STRING.TO_LOWER", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.ToLower(text)), nil
}

func stdTypesStringToUpper(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.STRING.TO_UPPER expected 1 argument, got %d", len(args))
	}

	text, err := stdTypesStringArg("TYPES.STRING.TO_UPPER", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.ToUpper(text)), nil
}

func stdTypesStringIsFoundIn(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPES.STRING.IS_FOUND_IN expected 2 arguments, got %d", len(args))
	}

	search, text, err := stdTypesStringSearchArgs("TYPES.STRING.IS_FOUND_IN", args)
	if err != nil {
		return Value{}, err
	}

	return NewBoolValue(strings.Contains(text, search)), nil
}

func stdTypesStringIsFoundAtStart(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPES.STRING.IS_FOUND_AT_START expected 2 arguments, got %d", len(args))
	}

	search, text, err := stdTypesStringSearchArgs("TYPES.STRING.IS_FOUND_AT_START", args)
	if err != nil {
		return Value{}, err
	}

	return NewBoolValue(strings.HasPrefix(text, search)), nil
}

func stdTypesStringIsFoundAtEnd(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPES.STRING.IS_FOUND_AT_END expected 2 arguments, got %d", len(args))
	}

	search, text, err := stdTypesStringSearchArgs("TYPES.STRING.IS_FOUND_AT_END", args)
	if err != nil {
		return Value{}, err
	}

	return NewBoolValue(strings.HasSuffix(text, search)), nil
}

func stdTypesStringReplace(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 3 {
		return Value{}, fmt.Errorf("TYPES.STRING.REPLACE expected 3 arguments, got %d", len(args))
	}

	text, err := stdTypesStringArg("TYPES.STRING.REPLACE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	oldText, err := stdTypesStringArg("TYPES.STRING.REPLACE", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	newText, err := stdTypesStringArg("TYPES.STRING.REPLACE", args[2], 3)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(strings.ReplaceAll(text, oldText, newText)), nil
}

func stdTypesStringSplit(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPES.STRING.SPLIT expected 2 arguments, got %d", len(args))
	}

	text, err := stdTypesStringArg("TYPES.STRING.SPLIT", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	separator, err := stdTypesStringArg("TYPES.STRING.SPLIT", args[1], 2)
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

func stdTypesStringJoin(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPES.STRING.JOIN expected 2 arguments, got %d", len(args))
	}

	array := resolveSpecializedValue(args[0])
	if array.Kind != ValueArray || array.Array == nil {
		return Value{}, fmt.Errorf("TYPES.STRING.JOIN argument 1 expected an array")
	}

	separator, err := stdTypesStringArg("TYPES.STRING.JOIN", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	parts := make([]string, 0, len(array.Array.Elements))

	for _, element := range array.Array.Elements {
		value := resolveSpecializedValue(element)

		if value.Kind == ValueString {
			if value.Text == nil {
				parts = append(parts, "")
			} else {
				parts = append(parts, value.Text.String())
			}

			continue
		}

		parts = append(parts, value.PrintString())
	}

	return NewStringValue(strings.Join(parts, separator)), nil
}

func stdTypesStringArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string", name, position)
	}

	return value.Text.String(), nil
}

func stdTypesStringSearchArgs(name string, args []Value) (string, string, error) {
	search, err := stdTypesStringArg(name, args[0], 1)
	if err != nil {
		return "", "", err
	}

	text, err := stdTypesStringArg(name, args[1], 2)
	if err != nil {
		return "", "", err
	}

	return search, text, nil
}
