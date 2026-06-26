package main

import "fmt"

type collectionFreezeState struct {
	maps    map[*Map]bool
	arrays  map[*Array]bool
	strings map[*String]bool
}

func freezeCollectionValue(value Value) (Value, error) {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("cannot freeze invalid map")
		}

		value.Map.IsFrozen = true
		return value, nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("cannot freeze invalid array")
		}

		value.Array.IsFrozen = true
		return value, nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("cannot freeze invalid string")
		}

		value.Text.IsFrozen = true
		return value, nil

	default:
		return value, nil
	}
}

func newCollectionFreezeState() *collectionFreezeState {
	return &collectionFreezeState{
		maps:    make(map[*Map]bool),
		arrays:  make(map[*Array]bool),
		strings: make(map[*String]bool),
	}
}

func deepFreezeCollectionValue(value Value) (Value, error) {
	return deepFreezeCollectionValueWithState(value, newCollectionFreezeState())
}

func deepFreezeCollectionValueWithState(value Value, state *collectionFreezeState) (Value, error) {
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

		if err := value.Map.ForEach(func(key string, binding Binding) error {
			frozenValue, err := deepFreezeNestedCollectionValue(binding.Value, state)
			if err != nil {
				return err
			}

			binding.Value = frozenValue
			return value.Map.Set(key, binding)
		}); err != nil {
			return Value{}, err
		}

		value.Map.IsFrozen = true
		return value, nil

	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("cannot freeze invalid array")
		}

		if state.arrays[value.Array] {
			return value, nil
		}

		state.arrays[value.Array] = true

		if value.Array.IsFrozen {
			if err := value.Array.ForEach(func(_ *Number, element Value) error {
				_, err := deepFreezeNestedCollectionValue(element, state)
				return err
			}); err != nil {
				return Value{}, err
			}
		} else {
			if err := value.Array.ForEach(func(index *Number, element Value) error {
				frozenValue, err := deepFreezeNestedCollectionValue(element, state)
				if err != nil {
					return err
				}

				return value.Array.Set(index, frozenValue)
			}); err != nil {
				return Value{}, err
			}
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
		return deepFreezeCollectionValueWithState(value, state)

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
