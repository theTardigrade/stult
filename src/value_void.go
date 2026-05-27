package main

func NewVoidValue() Value {
	return Value{Kind: ValueVoid}
}

func NewEmptyValue() Value {
	return NewVoidValue()
}
