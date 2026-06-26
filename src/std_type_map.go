package main

import "fmt"

func NewStdTypeMapMap() Value {
	entries := map[string]Binding{
		"ENTRIES": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapEntries)),
		"KEYS":    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapKeys)),
		"MERGE":   NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapMerge)),
		"VALUES":  NewImmutableBinding(NewBuiltinFunctionValue(StdTypeMapValues)),
	}

	return NewMapValue(entries, true)
}

func StdTypeMapMerge(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return Value{}, fmt.Errorf("TYPE.MAP.MERGE expected 2 or 3 arguments, got %d", len(args))
	}

	left := resolveSpecializedValue(args[0])
	if left.Kind != ValueMap {
		return Value{}, fmt.Errorf("TYPE.MAP.MERGE argument 1 must be a map")
	}

	if left.Map == nil {
		return Value{}, fmt.Errorf("TYPE.MAP.MERGE cannot merge invalid map")
	}

	right := resolveSpecializedValue(args[1])
	if right.Kind != ValueMap {
		return Value{}, fmt.Errorf("TYPE.MAP.MERGE argument 2 must be a map")
	}

	if right.Map == nil {
		return Value{}, fmt.Errorf("TYPE.MAP.MERGE cannot merge invalid map")
	}

	deep := false
	if len(args) == 3 {
		deepArgument := resolveSpecializedValue(args[2])
		if deepArgument.Kind != ValueBool {
			return Value{}, fmt.Errorf("TYPE.MAP.MERGE argument 3 expected a boolean")
		}

		deep = deepArgument.Bool
	}

	if deep {
		merged, err := deepMergeMapPair(left.Map, right.Map, newMapDeepMergeState())
		if err != nil {
			return Value{}, err
		}

		return Value{Kind: ValueMap, Map: merged}, nil
	}

	entries := make(map[string]Binding, len(left.Map.Entries)+len(right.Map.Entries))

	for key, binding := range left.Map.Entries {
		entries[key] = binding
	}

	for key, binding := range right.Map.Entries {
		entries[key] = binding
	}

	return NewMapValue(entries, false), nil
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
		return fmt.Errorf("TYPE.MAP.MERGE cannot merge into invalid map")
	}

	if source == nil {
		return fmt.Errorf("TYPE.MAP.MERGE cannot merge invalid map")
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
			return Binding{}, fmt.Errorf("TYPE.MAP.MERGE cannot merge invalid nested map")
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
		return nil, fmt.Errorf("TYPE.MAP.MERGE cannot merge invalid nested map")
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
