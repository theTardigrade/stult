package main

func NewStdMap() Value {
	entries := map[string]Binding{
		"FILE":  NewImmutableBinding(NewStdFileMap()),
		"IO":    NewImmutableBinding(NewStdIOMap()),
		"MATH":  NewImmutableBinding(NewStdMathMap()),
		"TYPES": NewImmutableBinding(NewStdTypesMap()),
	}

	return NewMapValue(entries, true)
}
