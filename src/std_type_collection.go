package main

import "fmt"

func NewStdTypeCollectionMap() Value {
	entries := map[string]Binding{
		"CLEAR":    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionClear)),
		"HAS":      NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionHas)),
		"IS_EMPTY": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionIsEmpty)),
		"SIZE":     NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionSize)),
	}

	return NewMapValue(entries, true)
}

func StdTypeCollectionSize(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.SIZE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.SIZE cannot determine size of invalid map")
		}

		return NewNumberValueFromInt(len(value.Map.Entries)), nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.SIZE cannot determine size of invalid array")
		}

		return NewNumberValueFromInt(len(value.Array.Elements)), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.SIZE cannot determine size of invalid string")
		}

		return NewNumberValueFromInt(len(value.Text.Runes)), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.SIZE cannot determine size of unknown value kind")
	}
}

func StdTypeCollectionIsEmpty(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_EMPTY expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_EMPTY cannot determine emptiness of invalid map")
		}

		return NewBoolValue(len(value.Map.Entries) == 0), nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_EMPTY cannot determine emptiness of invalid array")
		}

		return NewBoolValue(len(value.Array.Elements) == 0), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_EMPTY cannot determine emptiness of invalid string")
		}

		return NewBoolValue(len(value.Text.Runes) == 0), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_EMPTY cannot determine emptiness of unknown value kind")
	}
}

func StdTypeCollectionHas(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.HAS expected 2 arguments, got %d", len(args))
	}

	collection := resolveSpecializedValue(args[0])
	key := resolveSpecializedValue(args[1])

	switch collection.Kind {
	case ValueMap:
		if collection.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.HAS cannot inspect invalid map")
		}

		if key.Kind != ValueString || key.Text == nil {
			return NewBoolValue(false), nil
		}

		_, exists := collection.Map.Entries[key.Text.String()]
		return NewBoolValue(exists), nil

	case ValueArray:
		if collection.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.HAS cannot inspect invalid array")
		}

		index, err := numberToArrayIndex(key)
		if err != nil {
			return NewBoolValue(false), nil
		}

		return NewBoolValue(index >= 0 && index < len(collection.Array.Elements)), nil

	case ValueString:
		if collection.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.HAS cannot inspect invalid string")
		}

		index, err := numberToArrayIndex(key)
		if err != nil {
			return NewBoolValue(false), nil
		}

		return NewBoolValue(index >= 0 && index < len(collection.Text.Runes)), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.HAS cannot inspect unknown value kind")
	}
}

func StdTypeCollectionClear(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot clear invalid map")
		}

		if value.Map.IsImmutable {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot modify immutable map")
		}

		value.Map.Entries = make(map[string]Binding)
		return NewVoidValue(), nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot clear invalid array")
		}

		if value.Array.IsImmutable {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot modify immutable array")
		}

		value.Array.Elements = nil
		return NewVoidValue(), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot clear invalid string")
		}

		if value.Text.IsImmutable {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot modify immutable string")
		}

		value.Text.Runes = nil
		return NewVoidValue(), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot clear unknown value kind")
	}
}
