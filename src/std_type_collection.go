package main

import "fmt"

type collectionFreezeState struct {
	maps    map[*Map]bool
	arrays  map[*Array]bool
	strings map[*String]bool
}

type collectionCloneState struct {
	maps    map[*Map]*Map
	arrays  map[*Array]*Array
	strings map[*String]*String
}

func NewStdTypeCollectionMap() Value {
	entries := map[string]Binding{
		"CLEAR":     NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionClear)),
		"CLONE":     NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionClone)),
		"FREEZE":    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionFreeze)),
		"GET":       NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionGet)),
		"HAS":       NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionHas)),
		"IS_EMPTY":  NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionIsEmpty)),
		"IS_FROZEN": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeCollectionIsFrozen)),
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

		return NewNumberValueFromNumber(value.Map.Len()), nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.SIZE cannot determine size of invalid array")
		}

		return NewNumberValueFromNumber(value.Array.Len()), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.SIZE cannot determine size of invalid string")
		}

		return NewNumberValueFromNumber(value.Text.Len()), nil

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

		return NewBoolValue(value.Map.Len().Sign() == 0), nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_EMPTY cannot determine emptiness of invalid array")
		}

		return NewBoolValue(value.Array.Len().Sign() == 0), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.IS_EMPTY cannot determine emptiness of invalid string")
		}

		return NewBoolValue(value.Text.Len().Sign() == 0), nil

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

		keyText, err := mapKeyString(key)
		if err != nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.GET map key must be a string")
		}

		value, exists, err := collection.Map.GetFromString(keyText)
		if err != nil {
			return Value{}, err
		}

		if !exists {
			return fallback, nil
		}

		return value, nil

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

		if key.Kind != ValueNumber {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.GET string index must be a number")
		}

		value, exists, err := collection.Text.Get(key.Number)
		if err != nil {
			return Value{}, err
		}

		if !exists {
			return fallback, nil
		}

		return value, nil

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

		keyText, err := mapKeyString(key)
		if err != nil {
			return NewBoolValue(false), nil
		}

		_, exists, err := collection.Map.GetFromString(keyText)
		if err != nil {
			return NewBoolValue(false), nil
		}

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

		if key.Kind != ValueNumber {
			return NewBoolValue(false), nil
		}

		_, exists, err := collection.Text.Get(key.Number)
		if err != nil {
			return NewBoolValue(false), nil
		}

		return NewBoolValue(exists), nil

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

		if err := value.Map.Clear(); err != nil {
			if value.Map.IsFrozen {
				return Value{}, fmt.Errorf("TYPE.COLLECTION.CLEAR cannot modify frozen map")
			}

			return Value{}, err
		}

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

		if err := value.Text.Clear(); err != nil {
			return Value{}, err
		}

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
		value.Map.IsFrozen = true

		if err := value.Map.ForEach(func(key string, binding Binding) error {
			frozenValue, err := deepFreezeNestedCollectionValue(binding.Value, state)
			if err != nil {
				return err
			}

			binding.Value = frozenValue
			return value.Map.SetBindingUnchecked(key, binding)
		}); err != nil {
			return Value{}, err
		}

		return value, nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze invalid array")
		}

		if state.arrays[value.Array] || value.Array.IsFrozen {
			return value, nil
		}

		state.arrays[value.Array] = true

		if err := value.Array.ForEach(func(index *Number, element Value) error {
			frozenValue, err := deepFreezeNestedCollectionValue(element, state)
			if err != nil {
				return err
			}

			return value.Array.Set(index, frozenValue)
		}); err != nil {
			return Value{}, err
		}

		value.Array.IsFrozen = true
		return value, nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.FREEZE cannot freeze invalid string")
		}

		if state.strings[value.Text] {
			return value, nil
		}

		state.strings[value.Text] = true
		value.Text.IsFrozen = true

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
			Entries:  make(map[string]Binding, len(value.Map.Keys())),
			IsFrozen: false,
		}

		state.maps[value.Map] = clone

		if err := value.Map.ForEach(func(key string, binding Binding) error {
			clonedValue, err := deepCloneValue(binding.Value, state)
			if err != nil {
				return err
			}

			return clone.SetBindingUnchecked(key, Binding{
				Value:       clonedValue,
				IsImmutable: binding.IsImmutable,
			})
		}); err != nil {
			return Value{}, err
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

		clone := NewStringValue(value.Text.String()).Text

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
