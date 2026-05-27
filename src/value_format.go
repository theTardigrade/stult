package main

import "strconv"

func (v Value) String() string {
	return v.Format(DefaultFractionDigits)
}

func (v Value) PrintString() string {
	v = resolveSpecializedValue(v)

	switch v.Kind {
	case ValueString:
		return v.Text.String()
	default:
		return v.String()
	}
}

func (v Value) Format(fractionDigits int) string {
	v = resolveSpecializedValue(v)

	switch v.Kind {
	case ValueVoid:
		return "_"

	case ValueNumber:
		return formatNumber(v.Number, fractionDigits)

	case ValueBool:
		if v.Bool {
			return "true"
		}
		return "false"

	case ValueString:
		return strconv.Quote(v.Text.String())

	case ValueMap:
		return formatMap(v.Map, fractionDigits)

	case ValueArray:
		return formatArray(v.Array, fractionDigits)

	case ValueBuiltinFunction:
		return "<builtin function>"

	case ValueFunction:
		return "<function>"

	default:
		return "<unknown>"
	}
}

func (v Value) DebugString() string {
	v = resolveSpecializedValue(v)

	switch v.Kind {
	case ValueVoid:
		return "_"

	case ValueNumber:
		return formatNumber(v.Number, DefaultFractionDigits)

	case ValueBool:
		return v.String()

	case ValueString:
		return strconv.Quote(v.Text.String())

	case ValueMap:
		return formatMap(v.Map, DefaultFractionDigits)

	case ValueArray:
		return formatArray(v.Array, DefaultFractionDigits)

	case ValueFunction:
		return "<function>"

	case ValueBuiltinFunction:
		return "<builtin function>"

	default:
		return "<unknown>"
	}
}
