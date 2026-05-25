package main

import "fmt"

func NewStdTypesCollectionMap() Value {
	entries := map[string]Binding{
		"CLEAR":    NewImmutableBinding(NewBuiltinFunctionValue(stdTypesCollectionClear)),
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

	case ValueEmpty,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewEmptyValue(), nil

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

	case ValueEmpty,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewEmptyValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPES.COLLECTION.IS_EMPTY cannot determine emptiness of unknown value kind")
	}
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
		return NewEmptyValue(), nil

	case ValueArray:
		if value.Array.IsImmutable {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify immutable array")
		}

		value.Array.Elements = nil
		return NewEmptyValue(), nil

	case ValueEmptyCollection:
		if value.EmptyCollection == nil {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify invalid empty collection")
		}

		if value.EmptyCollection.IsImmutable {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify immutable empty collection")
		}

		value.EmptyCollection.Specialized = nil
		return NewEmptyValue(), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify invalid string")
		}

		if value.Text.IsImmutable {
			return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify immutable string")
		}

		value.Text.Runes = nil
		return NewEmptyValue(), nil

	case ValueEmpty,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewEmptyValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPES.COLLECTION.CLEAR cannot modify unknown value kind")
	}
}
