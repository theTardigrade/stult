package main

func NewStdTypeContractMap() Value {
	return NewMapValue(map[string]Binding{}, true)
}
