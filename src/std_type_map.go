package main

import "fmt"

func NewStdTypeMapMap() Value {
	entries := map[string]Binding{
		"ENTRIES":       NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapEntries)),
		"KEYS":          NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapKeys)),
		"MERGE_SHALLOW": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapShallowMerge)),
		"MERGE_DEEP":    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapDeepMerge)),
		"VALUES":        NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapValues)),
	}

	return NewMapValue(entries, true)
}

func StdTypeMapShallowMerge(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return Value{}, fmt.Errorf("TYPE.MAP.MERGE_SHALLOW expected at least 1 argument, got 0")
	}

	entries := make(map[string]Binding)

	for index, arg := range args {
		value := resolveSpecializedValue(arg)
		if value.Kind != ValueMap {
			return Value{}, fmt.Errorf("TYPE.MAP.MERGE_SHALLOW argument %d must be a map", index+1)
		}

		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.MAP.MERGE_SHALLOW cannot merge invalid map")
		}

		for key, binding := range value.Map.Entries {
			entries[key] = binding
		}
	}

	return NewMapValue(entries, false), nil
}

func StdTypeMapDeepMerge(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return Value{}, fmt.Errorf("TYPE.MAP.MERGE_DEEP expected at least 1 argument, got 0")
	}

	result := &Map{
		Entries:  make(map[string]Binding),
		IsFrozen: false,
	}

	state := newMapDeepMergeState()

	for index, arg := range args {
		value := resolveSpecializedValue(arg)
		if value.Kind != ValueMap {
			return Value{}, fmt.Errorf("TYPE.MAP.MERGE_DEEP argument %d must be a map", index+1)
		}

		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.MAP.MERGE_DEEP cannot merge invalid map")
		}

		if err := deepMergeMapInto(result, value.Map, state); err != nil {
			return Value{}, err
		}
	}

	return Value{Kind: ValueMap, Map: result}, nil
}

type mapDeepMergePair struct {
	left  *Map
	right *Map
}

type mapDeepMergeState struct {
	pairs map[mapDeepMergePair]*Map
}

func newMapDeepMergeState() *mapDeepMergeState {
	return &mapDeepMergeState{
		pairs: make(map[mapDeepMergePair]*Map),
	}
}

func deepMergeMapInto(target *Map, source *Map, state *mapDeepMergeState) error {
	if target == nil {
		return fmt.Errorf("TYPE.MAP.MERGE_DEEP cannot merge into invalid map")
	}

	if source == nil {
		return fmt.Errorf("TYPE.MAP.MERGE_DEEP cannot merge invalid map")
	}

	for key, incomingBinding := range source.Entries {
		currentBinding, exists := target.Entries[key]
		if !exists {
			target.Entries[key] = incomingBinding
			continue
		}

		mergedBinding, err := deepMergeBindings(currentBinding, incomingBinding, state)
		if err != nil {
			return err
		}

		target.Entries[key] = mergedBinding
	}

	return nil
}

func deepMergeBindings(current Binding, incoming Binding, state *mapDeepMergeState) (Binding, error) {
	currentValue := resolveSpecializedValue(current.Value)
	incomingValue := resolveSpecializedValue(incoming.Value)

	if currentValue.Kind == ValueMap && incomingValue.Kind == ValueMap {
		if currentValue.Map == nil || incomingValue.Map == nil {
			return Binding{}, fmt.Errorf("TYPE.MAP.MERGE_DEEP cannot merge invalid nested map")
		}

		merged, err := deepMergeMapPair(currentValue.Map, incomingValue.Map, state)
		if err != nil {
			return Binding{}, err
		}

		return Binding{
			Value:       Value{Kind: ValueMap, Map: merged},
			IsImmutable: incoming.IsImmutable,
		}, nil
	}

	return incoming, nil
}

func deepMergeMapPair(left *Map, right *Map, state *mapDeepMergeState) (*Map, error) {
	if left == nil || right == nil {
		return nil, fmt.Errorf("TYPE.MAP.MERGE_DEEP cannot merge invalid nested map")
	}

	pair := mapDeepMergePair{left: left, right: right}
	if existing, ok := state.pairs[pair]; ok {
		return existing, nil
	}

	merged := &Map{
		Entries:  make(map[string]Binding, len(left.Entries)+len(right.Entries)),
		IsFrozen: false,
	}

	state.pairs[pair] = merged

	for key, binding := range left.Entries {
		merged.Entries[key] = binding
	}

	if err := deepMergeMapInto(merged, right, state); err != nil {
		return nil, err
	}

	return merged, nil
}

func StdTypeMapKeys(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.MAP.KEYS expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.MAP.KEYS cannot inspect invalid map")
		}

		keys := sortedMapKeys(value.Map)
		elements := make([]Value, 0, len(keys))

		for _, key := range keys {
			elements = append(elements, NewStringValue(key))
		}

		return NewArrayValue(elements, false), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueString,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.MAP.KEYS cannot inspect unknown value kind")
	}
}

func StdTypeMapEntries(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.MAP.ENTRIES expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.MAP.ENTRIES cannot inspect invalid map")
		}

		keys := sortedMapKeys(value.Map)
		elements := make([]Value, 0, len(keys))

		for _, key := range keys {
			pair := NewArrayValue([]Value{
				NewStringValue(key),
				value.Map.Entries[key].Value,
			}, false)

			elements = append(elements, pair)
		}

		return NewArrayValue(elements, false), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueString,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.MAP.ENTRIES cannot inspect unknown value kind")
	}
}

func StdTypeMapValues(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.MAP.VALUES expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.MAP.VALUES cannot inspect invalid map")
		}

		keys := sortedMapKeys(value.Map)
		elements := make([]Value, 0, len(keys))

		for _, key := range keys {
			elements = append(elements, value.Map.Entries[key].Value)
		}

		return NewArrayValue(elements, false), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueString,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return NewVoidValue(), nil

	default:
		return Value{}, fmt.Errorf("TYPE.MAP.VALUES cannot inspect unknown value kind")
	}
}
