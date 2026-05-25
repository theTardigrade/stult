package main

func NewStdMap() Value {
	entries := map[string]Binding{
		"IO":    NewImmutableBinding(NewStdIOMap()),
		"MATH":  NewImmutableBinding(NewStdMathMap()),
		"TYPES": NewImmutableBinding(NewStdTypesMap()),
	}

	return NewMapValue(entries, true)
}
