package main

import (
	"fmt"
	"math/big"
)

func NewStdTypeArrayMap() Value {
	entries := map[string]Binding{
		"APPEND":  NewImmutableBinding(NewBuiltinFunctionValue(StdTypeArrayAppend)),
		"REVERSE": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeArrayReverse)),
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
