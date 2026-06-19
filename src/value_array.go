package main

import "strings"

type Array struct {
	Elements    []Value
	IsImmutable bool
}

func NewArrayValue(elements []Value, isImmutable bool) Value {
	return Value{
		Kind: ValueArray,
		Array: &Array{
			Elements:    elements,
			IsImmutable: isImmutable,
		},
	}
}

func (state *valueFormatState) formatArray(a *Array) string {
	if a == nil {
		return "{}"
	}

	if state.arrays[a] {
		return "<cyclical array>"
	}

	state.arrays[a] = true
	defer delete(state.arrays, a)

	parts := make([]string, 0, len(a.Elements))

	for _, element := range a.Elements {
		parts = append(parts, state.formatValue(element))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}
