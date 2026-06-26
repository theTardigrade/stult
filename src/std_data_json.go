package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func NewStdJSONMap() Value {
	entries := map[string]Binding{
		"IS_VALID": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdJSONIsValid)),
		"NEW":      NewImmutableBinding(NewBuiltinFunctionValue(builtinStdJSONNew)),
		"PARSE":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdJSONParse)),
	}

	return NewMapValue(entries, true)
}

func builtinStdJSONNew(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("JSON.NEW expected 1 argument, got %d", len(args))
	}

	goValue, err := stdJSONFromValue(args[0])
	if err != nil {
		return Value{}, err
	}

	data, err := json.Marshal(goValue)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(string(data)), nil
}

func builtinStdJSONParse(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("JSON.PARSE expected 1 argument, got %d", len(args))
	}

	text, err := stdJSONStringArg("JSON.PARSE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()

	var decoded any

	if err := decoder.Decode(&decoded); err != nil {
		return Value{}, err
	}

	var trailing any

	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Value{}, fmt.Errorf("JSON.PARSE expected exactly one JSON value")
		}

		return Value{}, err
	}

	return stdJSONToValue(decoded)
}

func builtinStdJSONIsValid(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("JSON.IS_VALID expected 1 argument, got %d", len(args))
	}

	text, err := stdJSONStringArg("JSON.IS_VALID", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	return NewBoolValue(json.Valid([]byte(text))), nil
}

func stdJSONStringArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string", name, position)
	}

	return value.Text.String(), nil
}

func stdJSONToValue(value any) (Value, error) {
	switch v := value.(type) {
	case nil:
		return NewVoidValue(), nil

	case bool:
		return NewBoolValue(v), nil

	case string:
		return NewStringValue(v), nil

	case json.Number:
		return NewNumberValueFromString(v.String())

	case []any:
		elements := make([]Value, 0, len(v))

		for _, item := range v {
			element, err := stdJSONToValue(item)
			if err != nil {
				return Value{}, err
			}

			elements = append(elements, element)
		}

		return NewArrayValue(elements, false), nil

	case map[string]any:
		mapValue := NewMap(nil, false)

		for key, item := range v {
			element, err := stdJSONToValue(item)
			if err != nil {
				return Value{}, err
			}

			if err := mapValue.Set(key, Binding{
				Value:       element,
				IsImmutable: isImmutableIdentifier(key),
			}); err != nil {
				return Value{}, err
			}
		}

		return Value{Kind: ValueMap, Map: mapValue}, nil

	default:
		return Value{}, fmt.Errorf("unsupported JSON value %T", value)
	}
}

func stdJSONFromValue(value Value) (any, error) {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueVoid:
		return nil, nil

	case ValueBool:
		return value.Bool, nil

	case ValueString:
		if value.Text == nil {
			return "", nil
		}

		return value.Text.String(), nil

	case ValueNumber:
		return json.Number(stdJSONNumberString(value)), nil

	case ValueArray:
		if value.Array == nil {
			return nil, fmt.Errorf("JSON.NEW cannot convert invalid array")
		}

		elements := make([]any, 0, value.Array.capacityHintHostLimited(int(arrayOrdinaryLimit)))

		if err := value.Array.ForEach(func(_ *Number, element Value) error {
			converted, err := stdJSONFromValue(element)
			if err != nil {
				return err
			}

			elements = append(elements, converted)
			return nil
		}); err != nil {
			return nil, err
		}

		return elements, nil

	case ValueMap:
		if value.Map == nil {
			return nil, fmt.Errorf("JSON.NEW cannot convert invalid map")
		}

		entries := make(map[string]any)

		if err := value.Map.ForEach(func(key string, binding Binding) error {
			converted, err := stdJSONFromValue(binding.Value)
			if err != nil {
				return err
			}

			entries[key] = converted
			return nil
		}); err != nil {
			return nil, err
		}

		return entries, nil

	case ValueFunction, ValueBuiltinFunction:
		return nil, fmt.Errorf("JSON.NEW cannot convert functions")

	default:
		return nil, fmt.Errorf("JSON.NEW cannot convert unknown value kind")
	}
}

func stdJSONNumberString(value Value) string {
	number := resolveSpecializedValue(value).Number
	if number == nil {
		return "0"
	}

	return number.Format(MaxDecimalPlaces)
}
