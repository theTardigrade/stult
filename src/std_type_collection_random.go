package main

import "fmt"

func StdTypeCollectionChoice(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE cannot choose from invalid array")
		}

		if value.Array.Len().Sign() == 0 {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE cannot choose from empty array")
		}

		length, err := stdMathRandArrayLength(value.Array, "TYPE.COLLECTION.CHOICE")
		if err != nil {
			return Value{}, err
		}

		index, err := stdMathRandIndex(length)
		if err != nil {
			return Value{}, err
		}

		chosen, ok, err := value.Array.Get(NewSmallNumber(int64(index)))
		if err != nil {
			return Value{}, err
		}

		if !ok {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE generated array index out of bounds")
		}

		return chosen, nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE cannot choose from invalid string")
		}

		if len(value.Text.Runes) == 0 {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE cannot choose from empty string")
		}

		index, err := stdMathRandIndex(len(value.Text.Runes))
		if err != nil {
			return Value{}, err
		}

		return NewStringValue(string(value.Text.Runes[index])), nil

	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE cannot choose from invalid map")
		}

		if value.Map.IsEmpty() {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE cannot choose from empty map")
		}

		keys := value.Map.Keys()

		index, err := stdMathRandIndex(len(keys))
		if err != nil {
			return Value{}, err
		}

		binding, _ := value.Map.Get(keys[index])
		return binding.Value, nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE expected an array, string or map")

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.CHOICE cannot choose from unknown value kind")
	}
}

func StdTypeCollectionShuffle(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.COLLECTION.SHUFFLE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])

	switch value.Kind {
	case ValueArray:
		if value.Array == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.SHUFFLE cannot shuffle invalid array")
		}

		if _, err := stdMathRandArrayLength(value.Array, "TYPE.COLLECTION.SHUFFLE"); err != nil {
			return Value{}, err
		}

		elements := make([]Value, 0, len(value.Array.Ordinary))
		if err := value.Array.ForEach(func(_ *Number, element Value) error {
			elements = append(elements, element)
			return nil
		}); err != nil {
			return Value{}, err
		}

		if err := stdMathRandShuffleValues(elements); err != nil {
			return Value{}, err
		}

		return NewArrayValue(elements, false), nil

	case ValueString:
		if value.Text == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.SHUFFLE cannot shuffle invalid string")
		}

		runes := append([]rune{}, value.Text.Runes...)

		if err := stdMathRandShuffleRunes(runes); err != nil {
			return Value{}, err
		}

		return NewStringValue(string(runes)), nil

	case ValueMap:
		if value.Map == nil {
			return Value{}, fmt.Errorf("TYPE.COLLECTION.SHUFFLE cannot shuffle invalid map")
		}

		keys := value.Map.Keys()
		values := make([]Value, 0, len(keys))

		for _, key := range keys {
			binding, _ := value.Map.Get(key)
			values = append(values, binding.Value)
		}

		if err := stdMathRandShuffleValues(values); err != nil {
			return Value{}, err
		}

		entries := make(map[string]Binding, len(keys))

		for index, key := range keys {
			originalBinding, _ := value.Map.Get(key)

			entries[key] = Binding{
				Value:       values[index],
				IsImmutable: originalBinding.IsImmutable,
			}
		}

		return NewMapValue(entries, false), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueFunction,
		ValueBuiltinFunction:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.SHUFFLE expected an array, string or map")

	default:
		return Value{}, fmt.Errorf("TYPE.COLLECTION.SHUFFLE cannot shuffle unknown value kind")
	}
}
