package main

import "fmt"

func valuesEqual(left Value, right Value) (bool, error) {
	left = resolveSpecializedValue(left)
	right = resolveSpecializedValue(right)

	if left.Kind != right.Kind {
		return false, nil
	}

	switch left.Kind {
	case ValueVoid:
		return true, nil

	case ValueNumber:
		return numberCompare(left.Number, right.Number) == 0, nil

	case ValueBool:
		return left.Bool == right.Bool, nil

	case ValueString:
		return left.Text.String() == right.Text.String(), nil

	case ValueMap:
		return left.Map == right.Map, nil

	case ValueArray:
		return left.Array == right.Array, nil

	case ValueFunction:
		return left.Function == right.Function, nil

	case ValueBuiltinFunction:
		return false, fmt.Errorf("cannot compare builtin functions")

	case ValueContract:
		if left.Contract == nil || right.Contract == nil {
			return left.Contract == right.Contract, nil
		}
		return left.Contract.SourceString() == right.Contract.SourceString(), nil

	default:
		return false, fmt.Errorf("cannot compare unknown value kind")
	}
}
