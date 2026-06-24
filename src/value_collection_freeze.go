package main

import "fmt"

type collectionFreezeState struct {
	maps    map[*Map]bool
	arrays  map[*Array]bool
	strings map[*String]bool
}

func freezeCollectionValue(value Value) (Value, error) {
	return deepFreezeCollectionValue(value, newCollectionFreezeState())
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
			return Value{}, fmt.Errorf("cannot freeze invalid map")
		}

		if state.maps[value.Map] {
			return value, nil
		}

		state.maps[value.Map] = true
		value.Map.IsFrozen = true

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
			return Value{}, fmt.Errorf("cannot freeze invalid array")
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
			return Value{}, fmt.Errorf("cannot freeze invalid string")
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
		return Value{}, fmt.Errorf("cannot freeze unknown value kind")
	}
}
