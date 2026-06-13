package main

import "fmt"

type collectionFreezeState struct {
	maps    map[*Map]bool
	arrays  map[*Array]bool
	strings map[*String]bool
}

func NewStdTypeCollectionMap() Value {
	entries := map[string]Binding{
		"CLEAR":     NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionClear)),
		"FREEZE":    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionFreeze)),
		"HAS":       NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionHas)),
		"IS_EMPTY":  NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionIsEmpty)),
		"IS_FROZEN": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionIsFrozen)),
		"SIZE":      NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionSize)),
	}

	return NewMapValue(entries, true)
}

func StdTypeCollectionSize(_ *RuntimeContext, args []Value) (Value, error) {
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

func StdTypeCollectionIsEmpty(_ *RuntimeContext, args []Value) (Value, error) {
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

func StdTypeCollectionHas(_ *RuntimeContext, args []Value) (Value, error) {
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

func StdTypeCollectionClear(_ *RuntimeContext, args []Value) (Value, error) {
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
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot modify frozen map")
		}

		value.Map.Entries = make(map[string]Binding)
		return NewVoidValue(), nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot clear invalid array")
		}

		if value.Array.IsImmutable {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot modify frozen array")
		}

		value.Array.Elements = nil
		return NewVoidValue(), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot clear invalid string")
		}

		if value.Text.IsImmutable {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot modify frozen string")
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

func StdTypeCollectionFreeze(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze invalid map")
		}

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze invalid array")
		}

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze invalid string")
		}

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze unknown value kind")
	}

	return deepFreezeCollectionValue(value, newCollectionFreezeState())
}

func StdTypeCollectionIsFrozen(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_FROZEN expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_FROZEN cannot inspect invalid map")
		}

		return NewBoolValue(value.Map.IsImmutable), nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_FROZEN cannot inspect invalid array")
		}

		return NewBoolValue(value.Array.IsImmutable), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_FROZEN cannot inspect invalid string")
		}

		return NewBoolValue(value.Text.IsImmutable), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return NewBoolValue(false), nil

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_FROZEN cannot inspect unknown value kind")
	}
}

func newCollectionFreezeState() *collectionFreezeState {
	return &collectionFreezeState{
		maps:    make(map[*Map]bool),
		arrays:  make(map[*Array]bool),
		strings: make(map[*String]bool),
	}
}

func deepFreezeCollectionValue(value Value, state *collectionFreezeState) (Value, error) {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze invalid map")
		}

		if state.maps[value.Map] {
			return value, nil
		}

		state.maps[value.Map] = true
		value.Map.IsImmutable = true

		for key, binding := range value.Map.Entries {
			frozenValue, err := deepFreezeNestedCollectionValue(binding.Value, state)
			if err != nil {
				return Value{}, err
			}

			binding.Value = frozenValue
			value.Map.Entries[key] = binding
		}

		return value, nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze invalid array")
		}

		if state.arrays[value.Array] {
			return value, nil
		}

		state.arrays[value.Array] = true
		value.Array.IsImmutable = true

		for index, element := range value.Array.Elements {
			frozenValue, err := deepFreezeNestedCollectionValue(element, state)
			if err != nil {
				return Value{}, err
			}

			value.Array.Elements[index] = frozenValue
		}

		return value, nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze invalid string")
		}

		if state.strings[value.Text] {
			return value, nil
		}

		state.strings[value.Text] = true
		value.Text.IsImmutable = true

		return value, nil

	default:
		return value, nil
	}
}

func deepFreezeNestedCollectionValue(value Value, state *collectionFreezeState) (Value, error) {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueMap,
		ValueArray,
		ValueString:
		return deepFreezeCollectionValue(value, state)

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return value, nil

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze unknown value kind")
	}
}
