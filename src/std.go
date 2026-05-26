package main

func NewStdMap() Value {
	entries := map[string]Binding{
		"CSV":   NewImmutableBinding(NewStdCSVMap()),
		"DATA":  NewImmutableBinding(NewStdDataMap()),
		"FILE":  NewImmutableBinding(NewStdFileMap()),
		"IO":    NewImmutableBinding(NewStdIOMap()),
		"JSON":  NewImmutableBinding(NewStdJSONMap()),
		"MATH":  NewImmutableBinding(NewStdMathMap()),
		"TIME":  NewImmutableBinding(NewStdTimeMap()),
		"TYPES": NewImmutableBinding(NewStdTypesMap()),
	}

	return NewMapValue(entries, true)
}
