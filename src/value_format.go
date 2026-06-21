package main

import "strconv"

func (v Value) String() string {
	return v.Format(DefaultDecimalPlacesToDisplay)
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
	state := newValueFormatState(fractionDigits)
	return state.formatValue(v)
}

func (v Value) DebugString() string {
	state := newValueFormatState(DefaultDecimalPlacesToDisplay)
	return state.formatValue(v)
}

type valueFormatState struct {
	fractionDigits int
	arrays         map[*Array]bool
	maps           map[*Map]bool
}

func newValueFormatState(fractionDigits int) *valueFormatState {
	return &valueFormatState{
		fractionDigits: fractionDigits,
		arrays:         make(map[*Array]bool),
		maps:           make(map[*Map]bool),
	}
}

func (state *valueFormatState) formatValue(v Value) string {
	v = resolveSpecializedValue(v)

	switch v.Kind {
	case ValueVoid:
		return "_"

	case ValueNumber:
		return formatNumber(v.Number, state.fractionDigits)

	case ValueBool:
		if v.Bool {
			return "+"
		}
		return "-"

	case ValueString:
		return strconv.Quote(v.Text.String())

	case ValueMap:
		return state.formatMap(v.Map)

	case ValueArray:
		return state.formatArray(v.Array)

	case ValueBuiltinFunction:
		return "<builtin function>"

	case ValueFunction:
		return "<function>"

	default:
		return "<unknown>"
	}
}
