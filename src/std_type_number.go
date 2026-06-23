package main

import (
	"fmt"
	"math/big"
	"strings"
)

func NewStdTypeNumberMap() Value {
	entries := map[string]Binding{
		"CLAMP":                  NewImmutableBinding(NewBuiltinFunctionValue(StdTypeNumberClamp)),
		"DEFAULT_DECIMAL_PLACES": NewImmutableBinding(NewNumberValueFromInt(DefaultDecimalPlacesToDisplay)),
		"FORMAT":                 NewImmutableBinding(NewBuiltinFunctionValue(StdTypeNumberFormat)),
		"FORMAT_SCIENTIFIC":      NewImmutableBinding(NewBuiltinFunctionValue(StdTypeNumberFormatScientific)),
		"IS_WHOLE":               NewImmutableBinding(NewBuiltinFunctionValue(StdTypeNumberIsWhole)),
		"MAX_DECIMAL_PLACES":     NewImmutableBinding(NewNumberValueFromInt(MaxDecimalPlaces)),
		"NEW":                    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeNumberNew)),
	}

	return NewMapValue(entries, true)
}

func StdTypeNumberIsWhole(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.NUMBER.IS_WHOLE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])
	if value.Kind != ValueNumber || value.Number == nil {
		return NewBoolValue(false), nil
	}

	_, accuracy := value.Number.Int(nil)

	return NewBoolValue(accuracy == big.Exact), nil
}

func StdTypeNumberClamp(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 3 {
		return Value{}, fmt.Errorf("TYPE.NUMBER.CLAMP expected 3 arguments, got %d", len(args))
	}

	value, err := StdTypeNumberArg("TYPE.NUMBER.CLAMP", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	minimum, err := StdTypeNumberArg("TYPE.NUMBER.CLAMP", args[1], 2)
	if err != nil {
		return Value{}, err
	}

	maximum, err := StdTypeNumberArg("TYPE.NUMBER.CLAMP", args[2], 3)
	if err != nil {
		return Value{}, err
	}

	if minimum.Cmp(maximum) > 0 {
		return Value{}, fmt.Errorf("TYPE.NUMBER.CLAMP minimum cannot be greater than maximum")
	}

	if value.Cmp(minimum) < 0 {
		return NewNumberValueFromNumber(CloneNumber(minimum)), nil
	}

	if value.Cmp(maximum) > 0 {
		return NewNumberValueFromNumber(CloneNumber(maximum)), nil
	}

	return NewNumberValueFromNumber(CloneNumber(value)), nil
}

func StdTypeNumberFormat(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 3 {
		return Value{}, fmt.Errorf("TYPE.NUMBER.FORMAT expected 1 to 3 arguments, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])
	if value.Kind != ValueNumber {
		return Value{}, fmt.Errorf("TYPE.NUMBER.FORMAT argument 1 expected a number")
	}

	decimalPlaces := int64(DefaultDecimalPlacesToDisplay)
	options := numberFormatOptions{}

	if len(args) >= 2 {
		second := resolveSpecializedValue(args[1])

		switch second.Kind {
		case ValueNumber:
			parsedDecimalPlaces, err := numberToInt64(second, "TYPE.NUMBER.FORMAT argument 2")
			if err != nil {
				return Value{}, err
			}

			decimalPlaces = parsedDecimalPlaces

		case ValueMap:
			if len(args) == 3 {
				return Value{}, fmt.Errorf("TYPE.NUMBER.FORMAT argument 2 expected a number when argument 3 is supplied")
			}

			parsedOptions, err := parseNumberFormatOptions(second, "TYPE.NUMBER.FORMAT argument 2")
			if err != nil {
				return Value{}, err
			}

			options = parsedOptions

		default:
			return Value{}, fmt.Errorf("TYPE.NUMBER.FORMAT argument 2 expected a number or map")
		}
	}

	if len(args) == 3 {
		third := resolveSpecializedValue(args[2])
		if third.Kind != ValueMap {
			return Value{}, fmt.Errorf("TYPE.NUMBER.FORMAT argument 3 expected a map")
		}

		parsedOptions, err := parseNumberFormatOptions(third, "TYPE.NUMBER.FORMAT argument 3")
		if err != nil {
			return Value{}, err
		}

		options = parsedOptions
	}

	if decimalPlaces < 0 || decimalPlaces > MaxDecimalPlaces {
		return Value{}, fmt.Errorf("TYPE.NUMBER.FORMAT decimal places must be between 0 and %d", MaxDecimalPlaces)
	}

	number := value.Number
	if options.Percent {
		number = numberMultiply(number, NewSmallNumber(100))
	}

	text := number.Format(int(decimalPlaces))
	if options.GroupDigits {
		text = groupFormattedNumberDigits(text)
	}

	if options.Percent {
		text += "%"
	}

	return NewStringValue(text), nil
}

type numberFormatOptions struct {
	Percent     bool
	GroupDigits bool
}

func parseNumberFormatOptions(value Value, argumentName string) (numberFormatOptions, error) {
	value = resolveSpecializedValue(value)
	if value.Kind != ValueMap || value.Map == nil {
		return numberFormatOptions{}, fmt.Errorf("%s expected a map", argumentName)
	}

	options := numberFormatOptions{}

	for key, binding := range value.Map.Entries {
		optionValue := resolveSpecializedValue(binding.Value)
		if optionValue.Kind != ValueBool {
			return numberFormatOptions{}, fmt.Errorf("TYPE.NUMBER.FORMAT option %q expected a bool", key)
		}

		switch key {
		case "PERCENT":
			options.Percent = optionValue.Bool

		case "GROUP_DIGITS":
			options.GroupDigits = optionValue.Bool

		default:
			return numberFormatOptions{}, fmt.Errorf("TYPE.NUMBER.FORMAT unknown option %q", key)
		}
	}

	return options, nil
}

func groupFormattedNumberDigits(text string) string {
	if text == "" {
		return text
	}

	sign := ""
	if strings.HasPrefix(text, "-") {
		sign = "-"
		text = strings.TrimPrefix(text, "-")
	}

	whole := text
	fraction := ""

	if dot := strings.Index(text, "."); dot >= 0 {
		whole = text[:dot]
		fraction = text[dot:]
	}

	if len(whole) <= 3 {
		return sign + whole + fraction
	}

	firstGroupLength := len(whole) % 3
	if firstGroupLength == 0 {
		firstGroupLength = 3
	}

	var builder strings.Builder
	builder.WriteString(sign)
	builder.WriteString(whole[:firstGroupLength])

	for i := firstGroupLength; i < len(whole); i += 3 {
		builder.WriteString("'")
		builder.WriteString(whole[i : i+3])
	}

	builder.WriteString(fraction)

	return builder.String()
}

func StdTypeNumberFormatScientific(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPE.NUMBER.FORMAT_SCIENTIFIC expected 2 arguments, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])
	if value.Kind != ValueNumber {
		return Value{}, fmt.Errorf("TYPE.NUMBER.FORMAT_SCIENTIFIC argument 1 expected a number")
	}

	significantDigits, err := numberToInt64(args[1], "TYPE.NUMBER.FORMAT_SCIENTIFIC argument 2")
	if err != nil {
		return Value{}, err
	}

	if significantDigits < 1 || significantDigits > MaxDecimalPlaces {
		return Value{}, fmt.Errorf("TYPE.NUMBER.FORMAT_SCIENTIFIC significant digits must be between 1 and %d", MaxDecimalPlaces)
	}

	return NewStringValue(value.Number.FormatScientific(int(significantDigits))), nil
}

func StdTypeNumberNew(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.NUMBER.NEW expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueNumber:
		return NewNumberValueFromNumber(CloneNumber(value.Number)), nil

	case ValueBool:
		if value.Bool {
			return NewNumberValueFromInt(1), nil
		}

		return NewNumberValueFromInt(0), nil

	case ValueString:
		if value.Text == nil {
			return NewVoidValue(), nil
		}

		text := strings.TrimSpace(value.Text.String())
		if text == "" {
			return NewVoidValue(), nil
		}

		number, err := NewNumberValueFromString(text)
		if err != nil {
			return NewVoidValue(), nil
		}

		return number, nil

	case ValueVoid,
		ValueMap,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.NUMBER.NEW cannot convert unknown value kind")
	}
}

func StdTypeNumberArg(name string, arg Value, position int) (*Number, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueNumber || value.Number == nil {
		return nil, fmt.Errorf("%s argument %d expected a number", name, position)
	}

	return value.Number, nil
}
