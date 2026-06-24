package main

import (
	"fmt"
	"math/big"
	"sort"
)

func NewStdTypeArrayMap() Value {
	entries := map[string]Binding{
		"APPEND":  NewImmutableBinding(NewBuiltinFunctionValue(StdTypeArrayAppend)),
		"REVERSE": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeArrayReverse)),
		"SORT":    NewImmutableBinding(NewBuiltinFunctionValue(StdTypeArraySort)),
	}

	return NewMapValue(entries, true)
}

func StdTypeArrayAppend(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 2 {
		return Value{}, fmt.Errorf("TYPE.ARRAY.APPEND expected at least 2 arguments, got %d", len(args))
	}

	target := resolveSpecializedValue(args[0])

	if target.Kind != ValueArray {
		return Value{}, fmt.Errorf("TYPE.ARRAY.APPEND expected an array")
	}

	for _, value := range args[1:] {
		if err := appendArrayValue(target, value); err != nil {
			return Value{}, err
		}
	}

	return NewVoidValue(), nil
}

func StdTypeArrayReverse(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.ARRAY.REVERSE expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])
	if value.Kind != ValueArray || value.Array == nil {
		return Value{}, fmt.Errorf("TYPE.ARRAY.REVERSE argument 1 expected an array")
	}

	result := NewArrayValue(nil, false)
	length := value.Array.lengthInteger()
	if length.Sign() == 0 {
		return result, nil
	}

	for index := new(big.Int).Sub(length, big.NewInt(1)); index.Sign() >= 0; index.Sub(index, big.NewInt(1)) {
		element, ok, err := value.Array.Get(NewBigIntNumber(index))
		if err != nil {
			return Value{}, err
		}
		if !ok {
			return Value{}, fmt.Errorf("TYPE.ARRAY.REVERSE encountered invalid array index")
		}

		if err := result.Array.Append(element); err != nil {
			return Value{}, err
		}
	}

	return result, nil
}

func StdTypeArraySort(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TYPE.ARRAY.SORT expected 1 argument, got %d", len(args))
	}

	value := resolveSpecializedValue(args[0])
	if value.Kind != ValueArray || value.Array == nil {
		return Value{}, fmt.Errorf("TYPE.ARRAY.SORT argument 1 expected an array")
	}

	elements := make([]Value, 0)
	if err := value.Array.ForEach(func(_ *Number, element Value) error {
		elements = append(elements, element)
		return nil
	}); err != nil {
		return Value{}, err
	}

	sort.SliceStable(elements, func(leftIndex int, rightIndex int) bool {
		return compareValuesForArraySort(elements[leftIndex], elements[rightIndex]) < 0
	})

	return NewArrayValue(elements, false), nil
}

func compareValuesForArraySort(left Value, right Value) int {
	left = resolveSpecializedValue(left)
	right = resolveSpecializedValue(right)

	leftRank := valueSortRank(left)
	rightRank := valueSortRank(right)
	if leftRank != rightRank {
		return compareInt(leftRank, rightRank)
	}

	switch left.Kind {
	case ValueVoid:
		return 0

	case ValueBool:
		return compareBool(left.Bool, right.Bool)

	case ValueNumber:
		return numberCompare(left.Number, right.Number)

	case ValueString:
		return compareStringValue(left, right)

	case ValueArray,
		ValueMap,
		ValueFunction,
		ValueBuiltinFunction:
		return 0

	default:
		return 0
	}
}

func valueSortRank(value Value) int {
	switch value.Kind {
	case ValueVoid:
		return 0
	case ValueBool:
		return 1
	case ValueNumber:
		return 2
	case ValueString:
		return 3
	case ValueArray:
		return 4
	case ValueMap:
		return 5
	case ValueFunction:
		return 6
	case ValueBuiltinFunction:
		return 7
	default:
		return 8
	}
}

func compareBool(left bool, right bool) int {
	if left == right {
		return 0
	}
	if !left {
		return -1
	}
	return 1
}

func compareStringValue(left Value, right Value) int {
	leftText := ""
	if left.Text != nil {
		leftText = left.Text.String()
	}

	rightText := ""
	if right.Text != nil {
		rightText = right.Text.String()
	}

	if leftText < rightText {
		return -1
	}
	if leftText > rightText {
		return 1
	}
	return 0
}

func compareInt(left int, right int) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func appendArrayValue(target Value, value Value) error {
	if target.Kind != ValueArray {
		return fmt.Errorf("TYPE.ARRAY.APPEND expected an array")
	}

	if target.Array == nil {
		return fmt.Errorf("TYPE.ARRAY.APPEND cannot append to invalid array")
	}

	if target.Array.IsFrozen {
		return fmt.Errorf("cannot modify frozen array")
	}

	return target.Array.Append(value)
}
