package main

func NewStdTypeFunctionMap() Value {
	return NewMapValue(map[string]Binding{}, true)
}
