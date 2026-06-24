package main

import (
	"fmt"
	"math/big"
)

func rangeIndexValue(object Value, start Value, end Value, step Value, isInclusive bool) (Value, error) {
	object = resolveSpecializedValue(object)

	switch object.Kind {
	case ValueArray:
		return rangeIndexArrayValue(object, start, end, step, isInclusive)

	case ValueString:
		return rangeIndexStringValue(object, start, end, step, isInclusive)

	case ValueMap:
		return Value{}, fmt.Errorf("map does not support range indexing")

	default:
		return Value{}, fmt.Errorf("cannot range-index non-array or non-string value")
	}
}

func rangeIndexArrayValue(object Value, start Value, end Value, step Value, isInclusive bool) (Value, error) {
	if object.Array == nil {
		return Value{}, fmt.Errorf("invalid array")
	}

	indexes, err := stultRangeValues(start, end, step, isInclusive)
	if err != nil {
		return Value{}, err
	}

	if len(indexes) == 0 {
		if err := validateEmptyRangeIndexBounds(start, end, object.Array.lengthInteger(), "array"); err != nil {
			return Value{}, err
		}
	}

	values := make([]Value, 0, len(indexes))
	for _, index := range indexes {
		index = resolveSpecializedValue(index)
		if index.Kind != ValueNumber {
			return Value{}, fmt.Errorf("array range index must be a number")
		}

		value, ok, err := object.Array.Get(index.Number)
		if err != nil {
			return Value{}, err
		}

		if !ok {
			return Value{}, fmt.Errorf("array range index %s out of bounds", formatArrayIndex(index.Number))
		}

		values = append(values, value)
	}

	return NewArrayValue(values, false), nil
}

func rangeIndexStringValue(object Value, start Value, end Value, step Value, isInclusive bool) (Value, error) {
	if object.Text == nil {
		return Value{}, fmt.Errorf("invalid string")
	}

	indexes, err := stultRangeValues(start, end, step, isInclusive)
	if err != nil {
		return Value{}, err
	}

	if len(indexes) == 0 {
		if err := validateEmptyRangeIndexBounds(start, end, big.NewInt(int64(len(object.Text.Runes))), "string"); err != nil {
			return Value{}, err
		}
	}

	runes := make([]rune, 0, len(indexes))
	for _, index := range indexes {
		stringIndex, err := numberToArrayIndex(index)
		if err != nil {
			return Value{}, err
		}

		if stringIndex < 0 || stringIndex >= len(object.Text.Runes) {
			return Value{}, fmt.Errorf("string range index %d out of bounds", stringIndex)
		}

		runes = append(runes, object.Text.Runes[stringIndex])
	}

	return NewStringValue(string(runes)), nil
}

func validateEmptyRangeIndexBounds(start Value, end Value, length *big.Int, collectionName string) error {
	startInteger, err := numberToExactInteger(start, collectionName+" range start")
	if err != nil {
		return err
	}

	endInteger, err := numberToExactInteger(end, collectionName+" range end")
	if err != nil {
		return err
	}

	if length == nil {
		length = big.NewInt(0)
	}

	if startInteger.Sign() < 0 || startInteger.Cmp(length) > 0 {
		return fmt.Errorf("%s range index boundary %s out of bounds", collectionName, startInteger.String())
	}

	if endInteger.Sign() < 0 || endInteger.Cmp(length) > 0 {
		return fmt.Errorf("%s range index boundary %s out of bounds", collectionName, endInteger.String())
	}

	return nil
}
