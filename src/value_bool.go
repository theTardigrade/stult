package main

func NewBoolValue(value bool) Value {
	return Value{Kind: ValueBool, Bool: value}
}
