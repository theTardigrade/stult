package main

import "fmt"

func evalPrefix(operator string, right Value) (Value, error) {
	right = resolveSpecializedValue(right)

	switch operator {
	case "-":
		if right.Kind != ValueNumber {
			return Value{}, fmt.Errorf("unary '-' requires a number")
		}

		out := CloneNumber(right.Number)
		out.Neg(out)

		return Value{Kind: ValueNumber, Number: out}, nil

	case "!":
		if right.Kind != ValueBool {
			return Value{}, fmt.Errorf("unary '!' requires a bool")
		}

		return NewBoolValue(!right.Bool), nil

	default:
		return Value{}, fmt.Errorf("unknown prefix operator %q", operator)
	}
}

func evalBinary(operator string, left Value, right Value) (Value, error) {
	left = resolveSpecializedValue(left)
	right = resolveSpecializedValue(right)

	if operator == "=" || operator == "!" {
		equal, err := valuesEqual(left, right)
		if err != nil {
			return Value{}, err
		}

		if operator == "!" {
			equal = !equal
		}

		return NewBoolValue(equal), nil
	}

	if operator == "+" && left.Kind == ValueString && right.Kind == ValueString {
		return NewStringValue(left.Text.String() + right.Text.String()), nil
	}

	if left.Kind != ValueNumber || right.Kind != ValueNumber {
		return Value{}, fmt.Errorf("operator %q requires numbers", operator)
	}

	switch operator {
	case "+":
		out := newFloat()
		out.Add(left.Number, right.Number)
		return Value{Kind: ValueNumber, Number: out}, nil

	case "-":
		out := newFloat()
		out.Sub(left.Number, right.Number)
		return Value{Kind: ValueNumber, Number: out}, nil

	case "*":
		out := newFloat()
		out.Mul(left.Number, right.Number)
		return Value{Kind: ValueNumber, Number: out}, nil

	case "/":
		if right.Number.Sign() == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}

		out := newFloat()
		out.Quo(left.Number, right.Number)
		return Value{Kind: ValueNumber, Number: out}, nil

	case "<":
		return NewBoolValue(left.Number.Cmp(right.Number) < 0), nil

	case "<=":
		return NewBoolValue(left.Number.Cmp(right.Number) <= 0), nil

	case ">":
		return NewBoolValue(left.Number.Cmp(right.Number) > 0), nil

	case ">=":
		return NewBoolValue(left.Number.Cmp(right.Number) >= 0), nil

	default:
		return Value{}, fmt.Errorf("unknown binary operator %q", operator)
	}
}
