package main

type ValueKind int

const (
	ValueVoid ValueKind = iota
	ValueNumber
	ValueBool
	ValueString
	ValueMap
	ValueArray
	ValueFunction
	ValueBuiltinFunction
	ValueContract
)

type Value struct {
	Kind            ValueKind
	Number          *Number
	Bool            bool
	Text            *String
	Map             *Map
	Array           *Array
	Function        *Function
	BuiltinFunction BuiltinFunction
	Contract        *BindingContract
}

// resolveSpecializedValue normalizes wrapper/specialized values before
// ordinary runtime operations. It is currently a no-op.
func resolveSpecializedValue(value Value) Value {
	return value
}
