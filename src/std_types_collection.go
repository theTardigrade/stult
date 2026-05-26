package main

import "fmt"

func NewStdTypesCollectionMap() Value {
	entries := map[string]Binding{
		"CLEAR":    NewImmutableBinding(NewBuiltinFunctionValue(stdTypesCollectionClear)),
		"HAS":      NewImmutableBinding(NewBuiltinFunctionValue(stdTypesCollectionHas)),
		"IS_EMPTY": NewImmutableBinding(NewBuiltinFunctionValue(stdTypesCollectionIsEmpty)),
		"SIZE":     NewImmutableBinding(NewBuiltinFunctionValue(stdTypesCollectionSize)),
	}

	return NewMapValue(entries, true)
}

func stdTypesCollectionSize(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.COLLECTION.SIZE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		return NewNumberValueFromInt(len(value.Map.Entries)), nil

	case ValueArray:
		return NewNumberValueFromInt(len(value.Array.Elements)), nil

	case ValueEmptyCollection:
		return NewNumberValueFromInt(0), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.SIZE cannot determine size of invalid string")
		}

		return NewNumberValueFromInt(len(value.Text.Runes)), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPES.COLLECTION.SIZE cannot determine size of unknown value kind")
	}
}

func stdTypesCollectionIsEmpty(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.COLLECTION.IS_EMPTY expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		return NewBoolValue(len(value.Map.Entries) == 0), nil

	case ValueArray:
		return NewBoolValue(len(value.Array.Elements) == 0), nil

	case ValueEmptyCollection:
		return NewBoolValue(true), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.IS_EMPTY cannot determine emptiness of invalid string")
		}

		return NewBoolValue(len(value.Text.Runes) == 0), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPES.COLLECTION.IS_EMPTY cannot determine emptiness of unknown value kind")
	}
}

func stdTypesCollectionHas(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 2 {
		return Value{}, fmt.Errorf("TYPES.COLLECTION.HAS expected 2 arguments, got %d", len(args))
	}

	collection := resolveSpecializedValue(args[0])
	key := resolveSpecializedValue(args[1])

	switch collection.Kind {
	case ValueMap:
		if key.Kind != ValueString {
			return NewBoolValue(false), nil
		}

		_, exists := collection.Map.Entries[key.Text.String()]
		return NewBoolValue(exists), nil

	case ValueArray:
		index, ok := collectionIndex(key)
		if !ok {
			return NewBoolValue(false), nil
		}

		return NewBoolValue(index >= 0 && index < len(collection.Array.Elements)), nil

	case ValueString:
		if collection.Text == nil {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.HAS cannot inspect invalid string")
		}

		index, ok := collectionIndex(key)
		if !ok {
			return NewBoolValue(false), nil
		}

		return NewBoolValue(index >= 0 && index < len(collection.Text.Runes)), nil

	case ValueEmptyCollection:
		return NewBoolValue(false), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPES.COLLECTION.HAS cannot inspect unknown value kind")
	}
}

func collectionIndex(value Value) (int, bool) {
	index, err := numberToArrayIndex(value)
	if err != nil {
		return 0, false
	}

	return index, true
}

func stdTypesCollectionClear(_ *Interpreter, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map.IsImmutable {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify immutable map")
		}

		value.Map.Entries = make(map[string]Binding)
		return NewVoidValue(), nil

	case ValueArray:
		if value.Array.IsImmutable {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify immutable array")
		}

		value.Array.Elements = nil
		return NewVoidValue(), nil

	case ValueEmptyCollection:
		if value.EmptyCollection == nil {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify invalid empty collection")
		}

		if value.EmptyCollection.IsImmutable {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify immutable empty collection")
		}

		value.EmptyCollection.Specialized = nil
		return NewVoidValue(), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify invalid string")
		}

		if value.Text.IsImmutable {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify immutable string")
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
		return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify unknown value kind")
	}
}
