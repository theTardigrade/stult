package main

func NewStdMap() Value {
	entries := map[string]Binding{
		"BOOL":  NewImmutableBinding(NewStdBoolMap()),
		"IO":    NewImmutableBinding(NewStdIOMap()),
		"MATH":  NewImmutableBinding(NewStdMathMap()),
		"TYPES": NewImmutableBinding(NewStdTypesMap()),
	}

	return NewMapValue(entries, true)
}
