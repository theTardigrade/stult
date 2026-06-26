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

	merged := NewMap(make(map[string]Binding, left.Map.EntryCount()+right.Map.EntryCount()), false)

	if err := left.Map.ForEach(func(key string, binding Binding) error {
		return merged.Set(key, binding)
	}); err != nil {
		return Value{}, err
	}

	if err := right.Map.ForEach(func(key string, binding Binding) error {
		return merged.Set(key, binding)
	}); err != nil {
		return Value{}, err
	}

	return Value{Kind: ValueMap, Map: merged}, nil
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

	return source.ForEach(func(key string, incomingBinding Binding) error {
		currentBinding, exists := target.Get(key)
		if !exists {
			return target.Set(key, incomingBinding)
		}

		mergedBinding, err := deepMergeBindings(currentBinding, incomingBinding, state)
		if err != nil {
			return err
		}

		return target.Set(key, mergedBinding)
	})
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

	merged := NewMap(make(map[string]Binding, left.EntryCount()+right.EntryCount()), false)

	state.pairs[pair] = merged

	if err := left.ForEach(func(key string, binding Binding) error {
		return merged.Set(key, binding)
	}); err != nil {
		return nil, err
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

		keys := value.Map.Keys()
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

		keys := value.Map.Keys()
		elements := make([]Value, 0, len(keys))

		for _, key := range keys {
			binding, _ := value.Map.Get(key)
			pair := NewArrayValue([]Value{
				NewStringValue(key),
				binding.Value,
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

		keys := value.Map.Keys()
		elements := make([]Value, 0, len(keys))

		for _, key := range keys {
			binding, _ := value.Map.Get(key)
			elements = append(elements, binding.Value)
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
