package main

import (
	"fmt"
	"math/big"
)

type collectionCloneState struct {
	maps    map[*Map]*Map
	arrays  map[*Array]*Array
	strings map[*String]*String
}

func NewStdTypeCollectionMap() Value {
	entries := map[string]Binding{
		"CHOICE":    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionChoice)),
		"CLEAR":     NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionClear)),
		"CLONE":     NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionClone)),
		"FREEZE":    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionFreeze)),
		"GET":       NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionGet)),
		"HAS":       NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionHas)),
		"IS_EMPTY":  NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionIsEmpty)),
		"IS_FROZEN": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionIsFrozen)),
		"SHUFFLE":   NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionShuffle)),
		"SIZE":      NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionSize)),
	}

	return NewMapValue(entries, true)
}

func StdTypeCollectionClone(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.CLONE expected 1 argument, got %d", len(args))
	}

	return deepCloneValue(args[0], newCollectionCloneState())
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

		return NewNumberValueFromNumber(value.Array.Len()), nil

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

		return NewBoolValue(value.Array.Len().Sign() == 0), nil

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

func StdTypeCollectionGet(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 2 && len(args) != 3 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.GET expected 2 or 3 arguments, got %d", len(args))
	}

	collection := resolveSpecializedValue(args[0])
	key := resolveSpecializedValue(args[1])
	fallback := collectionGetFallback(args)

	switch collection.Kind {
	case ValueMap:
		if collection.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.GET cannot inspect invalid map")
		}

		if key.Kind != ValueString || key.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.GET map key must be a string")
		}

		binding, exists := collection.Map.Entries[key.Text.String()]
		if !exists {
			return fallback, nil
		}

		return binding.Value, nil

	case ValueArray:
		if collection.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.GET cannot inspect invalid array")
		}

		if key.Kind != ValueNumber {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.GET array index must be a number")
		}

		value, exists, err := collection.Array.Get(key.Number)
		if err != nil {
			return Value{}, err
		}

		if !exists {
			return fallback, nil
		}

		return value, nil

	case ValueString:
		if collection.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.GET cannot inspect invalid string")
		}

		index, valid, err := collectionGetIndex(key, "string")
		if err != nil {
			return Value{}, err
		}

		if !valid || index >= len(collection.Text.Runes) {
			return fallback, nil
		}

		return NewStringValue(string(collection.Text.Runes[index])), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.GET first argument must be a collection")

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.GET cannot inspect unknown value kind")
	}
}

func collectionGetFallback(args []Value) Value {
	if len(args) == 3 {
		return args[2]
	}

	return NewVoidValue()
}

func collectionGetIndex(key Value, collectionName string) (int, bool, error) {
	key = resolveSpecializedValue(key)

	if key.Kind != ValueNumber {
		return 0, false, fmt.Errorf("TYPE.COLLECTION.GET %s index must be a number", collectionName)
	}

	integer, accuracy := key.Number.Int(nil)
	if accuracy != big.Exact {
		return 0, false, fmt.Errorf("TYPE.COLLECTION.GET %s index must be an integer", collectionName)
	}

	if integer.Sign() < 0 {
		return 0, false, nil
	}

	if !integer.IsInt64() {
		return 0, false, nil
	}

	index64 := integer.Int64()
	index := int(index64)
	if int64(index) != index64 {
		return 0, false, nil
	}

	return index, true, nil
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

		if key.Kind != ValueNumber {
			return NewBoolValue(false), nil
		}

		_, exists, err := collection.Array.Get(key.Number)
		if err != nil {
			return NewBoolValue(false), nil
		}

		return NewBoolValue(exists), nil

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

		if value.Map.IsFrozen {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot modify frozen map")
		}

		value.Map.Entries = make(map[string]Binding)
		return NewVoidValue(), nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot clear invalid array")
		}

		if value.Array.IsFrozen {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot modify frozen array")
		}

		if err := value.Array.Clear(); err != nil {
			return Value{}, err
		}

		return NewVoidValue(), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot clear invalid string")
		}

		if value.Text.IsFrozen {
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
	if len(args) < 1 || len(args) > 2 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE expected 1 or 2 arguments, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	deep := false
	if len(args) == 2 {
		deepArgument := resolveSpecializedValue(args[1])
		if deepArgument.Kind != ValueBool {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE argument 2 expected a boolean")
		}

		deep = deepArgument.Bool
	}

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

	if deep {
		return deepFreezeCollectionValue(value)
	}

	return freezeCollectionValue(value)
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

		return NewBoolValue(value.Map.IsFrozen), nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_FROZEN cannot inspect invalid array")
		}

		return NewBoolValue(value.Array.IsFrozen), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_FROZEN cannot inspect invalid string")
		}

		return NewBoolValue(value.Text.IsFrozen), nil

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

func newCollectionCloneState() *collectionCloneState {
	return &collectionCloneState{
		maps:    make(map[*Map]*Map),
		arrays:  make(map[*Array]*Array),
		strings: make(map[*String]*String),
	}
}

func deepCloneValue(value Value, state *collectionCloneState) (Value, error) {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLONE cannot clone invalid map")
		}

		if clone, exists := state.maps[value.Map]; exists {
			return Value{Kind: ValueMap, Map: clone}, nil
		}

		clone := &Map{
			Entries:  make(map[string]Binding, len(value.Map.Entries)),
			IsFrozen: false,
		}

		state.maps[value.Map] = clone

		for key, binding := range value.Map.Entries {
			clonedValue, err := deepCloneValue(binding.Value, state)
			if err != nil {
				return Value{}, err
			}

			clone.Entries[key] = Binding{
				Value:       clonedValue,
				IsImmutable: binding.IsImmutable,
			}
		}

		return Value{Kind: ValueMap, Map: clone}, nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLONE cannot clone invalid array")
		}

		if clone, exists := state.arrays[value.Array]; exists {
			return Value{Kind: ValueArray, Array: clone}, nil
		}

		clone := &Array{
			Ordinary: make([]Value, 0, len(value.Array.Ordinary)),
			Length:   NewSmallNumber(0),
			IsFrozen: false,
		}

		state.arrays[value.Array] = clone

		if err := value.Array.ForEach(func(_ *Number, element Value) error {
			clonedValue, err := deepCloneValue(element, state)
			if err != nil {
				return err
			}

			if err := clone.Append(clonedValue); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return Value{}, err
		}

		return Value{Kind: ValueArray, Array: clone}, nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLONE cannot clone invalid string")
		}

		if clone, exists := state.strings[value.Text]; exists {
			return Value{Kind: ValueString, Text: clone}, nil
		}

		clone := &String{
			Runes:    append([]rune(nil), value.Text.Runes...),
			IsFrozen: false,
		}

		state.strings[value.Text] = clone

		return Value{Kind: ValueString, Text: clone}, nil

	case ValueNumber:
		if value.Number == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CLONE cannot clone invalid number")
		}

		return NewNumberValueFromNumber(CloneNumber(value.Number)), nil

	case ValueVoid,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return value, nil

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.CLONE cannot clone unknown value kind")
	}
}
