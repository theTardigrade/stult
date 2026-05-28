package main

func NewStdMap() Value {
	entries := map[string]Binding{
		"DATA": NewImmutableBinding(NewStdDataMap()),
		"FILE": NewImmutableBinding(NewStdFileMap()),
		"IO":   NewImmutableBinding(NewStdIOMap()),
		"MATH": NewImmutableBinding(NewStdMathMap()),
		"TIME": NewImmutableBinding(NewStdTimeMap()),
		"TYPE": NewImmutableBinding(NewStdTypeMap()),
	}

	return NewMapValue(entries, true)
}
