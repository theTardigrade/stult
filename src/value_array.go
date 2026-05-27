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

func formatArray(a *Array, fractionDigits int) string {
	if a == nil {
		return "{}"
	}

	parts := make([]string, 0, len(a.Elements))

	for _, element := range a.Elements {
		parts = append(parts, element.Format(fractionDigits))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}
