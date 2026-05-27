package main

import "fmt"

func NewStdTypeMap() Value {
	entries := map[string]Binding{
		"ARRAY":      NewImmutableBinding(NewStdTypeArrayMap()),
		"BOOL":       NewImmutableBinding(NewStdTypeBoolMap()),
		"COLLECTION": NewImmutableBinding(NewStdTypeCollectionMap()),
		"MAP":        NewImmutableBinding(NewStdTypeMapMap()),
		"NUMBER":     NewImmutableBinding(NewStdTypeNumberMap()),
		"STRING":     NewImmutableBinding(NewStdTypeStringMap()),

		"IS_ARRAY":            NewImmutableBinding(NewBuiltinFunctionValue(StdTypeIsArray)),
		"IS_BOOL":             NewImmutableBinding(NewBuiltinFunctionValue(StdTypeIsBool)),
		"IS_BUILTIN_FUNCTION": NewImmutableBinding(NewBuiltinFunctionValue(StdTypeIsBuiltinFunction)),
		"IS_COLLECTION":       NewImmutableBinding(NewBuiltinFunctionValue(StdTypeIsCollection)),
		"IS_FUNCTION":         NewImmutableBinding(NewBuiltinFunctionValue(StdTypeIsFunction)),
		"IS_MAP":              NewImmutableBinding(NewBuiltinFunctionValue(StdTypeIsMap)),
		"IS_NUMBER":           NewImmutableBinding(NewBuiltinFunctionValue(StdTypeIsNumber)),
		"IS_STRING":           NewImmutableBinding(NewBuiltinFunctionValue(StdTypeIsString)),
		"IS_VOID":             NewImmutableBinding(NewBuiltinFunctionValue(StdTypeIsVoid)),
	}

	return NewMapValue(entries, true)
}

func StdTypeIsArray(_ *Interpreter, args []Value) (Value, error) {
	return StdTypePredicate("TYPE.IS_ARRAY", args, func(value Value) bool {
		return value.Kind == ValueArray
	})
}

func StdTypeIsBool(_ *Interpreter, args []Value) (Value, error) {
	return StdTypePredicate("TYPE.IS_BOOL", args, func(value Value) bool {
		return value.Kind == ValueBool
	})
}

func StdTypeIsBuiltinFunction(_ *Interpreter, args []Value) (Value, error) {
	return StdTypePredicate("TYPE.IS_BUILTIN_FUNCTION", args, func(value Value) bool {
		return value.Kind == ValueBuiltinFunction
	})
}

func StdTypeIsCollection(_ *Interpreter, args []Value) (Value, error) {
	return StdTypePredicate("TYPE.IS_COLLECTION", args, func(value Value) bool {
		switch value.Kind {
		case ValueMap, ValueArray, ValueString:
			return true
		default:
			return false
		}
	})
}

func StdTypeIsFunction(_ *Interpreter, args []Value) (Value, error) {
	return StdTypePredicate("TYPE.IS_FUNCTION", args, func(value Value) bool {
		return value.Kind == ValueFunction
	})
}

func StdTypeIsMap(_ *Interpreter, args []Value) (Value, error) {
	return StdTypePredicate("TYPE.IS_MAP", args, func(value Value) bool {
		return value.Kind == ValueMap
	})
}

func StdTypeIsNumber(_ *Interpreter, args []Value) (Value, error) {
	return StdTypePredicate("TYPE.IS_NUMBER", args, func(value Value) bool {
		return value.Kind == ValueNumber
	})
}

func StdTypeIsString(_ *Interpreter, args []Value) (Value, error) {
	return StdTypePredicate("TYPE.IS_STRING", args, func(value Value) bool {
		return value.Kind == ValueString
	})
}

func StdTypeIsVoid(_ *Interpreter, args []Value) (Value, error) {
	return StdTypePredicate("TYPE.IS_VOID", args, func(value Value) bool {
		return value.Kind == ValueVoid
	})
}

func StdTypePredicate(name string, args []Value, predicate func(Value) bool) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("%s expected 1 argument, got %d", name, len(args))
	}

	value := resolveSpecializedValue(args[0])

	return NewBoolValue(predicate(value)), nil
}
