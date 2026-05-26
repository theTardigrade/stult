package main

import "fmt"

func NewStdTypesMap() Value {
	entries := map[string]Binding{
		"ARRAY":      NewImmutableBinding(NewStdTypesArrayMap()),
		"BOOL":       NewImmutableBinding(NewStdTypesBoolMap()),
		"COLLECTION": NewImmutableBinding(NewStdTypesCollectionMap()),
		"MAP":        NewImmutableBinding(NewStdTypesMapMap()),
		"NUMBER":     NewImmutableBinding(NewStdTypesNumberMap()),

		"IS_ARRAY":            NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsArray)),
		"IS_BOOL":             NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsBool)),
		"IS_BUILTIN_FUNCTION": NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsBuiltinFunction)),
		"IS_COLLECTION":       NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsCollection)),
		"IS_EMPTY_COLLECTION": NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsEmptyCollection)),
		"IS_FUNCTION":         NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsFunction)),
		"IS_MAP":              NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsMap)),
		"IS_NUMBER":           NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsNumber)),
		"IS_STRING":           NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsString)),
		"IS_VOID":             NewImmutableBinding(NewBuiltinFunctionValue(stdTypesIsVoid)),
	}

	return NewMapValue(entries, true)
}

func stdTypesIsArray(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_ARRAY", args, func(value Value) bool {
		return value.Kind == ValueArray
	})
}

func stdTypesIsBool(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_BOOL", args, func(value Value) bool {
		return value.Kind == ValueBool
	})
}

func stdTypesIsBuiltinFunction(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_BUILTIN_FUNCTION", args, func(value Value) bool {
		return value.Kind == ValueBuiltinFunction
	})
}

func stdTypesIsCollection(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_COLLECTION", args, func(value Value) bool {
		switch value.Kind {
		case ValueMap, ValueArray, ValueString, ValueEmptyCollection:
			return true
		default:
			return false
		}
	})
}

func stdTypesIsEmptyCollection(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_EMPTY_COLLECTION", args, func(value Value) bool {
		return value.Kind == ValueEmptyCollection
	})
}

func stdTypesIsFunction(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_FUNCTION", args, func(value Value) bool {
		return value.Kind == ValueFunction
	})
}

func stdTypesIsMap(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_MAP", args, func(value Value) bool {
		return value.Kind == ValueMap
	})
}

func stdTypesIsNumber(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_NUMBER", args, func(value Value) bool {
		return value.Kind == ValueNumber
	})
}

func stdTypesIsString(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_STRING", args, func(value Value) bool {
		return value.Kind == ValueString
	})
}

func stdTypesIsVoid(_ *Interpreter, args []Value) (Value, error) {
	return stdTypesPredicate("TYPES.IS_VOID", args, func(value Value) bool {
		return value.Kind == ValueVoid
	})
}

func stdTypesPredicate(name string, args []Value, predicate func(Value) bool) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("%s expected 1 argument, got %d", name, len(args))
	}

	value := resolveSpecializedValue(args[0])

	return NewBoolValue(predicate(value)), nil
}
