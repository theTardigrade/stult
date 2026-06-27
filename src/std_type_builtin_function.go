package main

func NewStdTypeBuiltinFunctionMap() Value {
	return NewMapValue(map[string]Binding{}, true)
}
