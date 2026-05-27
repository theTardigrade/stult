package main

func NewStdMap() Value {
	entries := map[string]Binding{
		"CSV":  NewImmutableBinding(NewStdCSVMap()),
		"DATA": NewImmutableBinding(NewStdDataMap()),
		"FILE": NewImmutableBinding(NewStdFileMap()),
		"IO":   NewImmutableBinding(NewStdIOMap()),
		"JSON": NewImmutableBinding(NewStdJSONMap()),
		"MATH": NewImmutableBinding(NewStdMathMap()),
		"TIME": NewImmutableBinding(NewStdTimeMap()),
		"TYPE": NewImmutableBinding(NewStdTypeMap()),
	}

	return NewMapValue(entries, true)
}
