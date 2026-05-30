package main

func NewStdMap(args []string) Value {
	entries := map[string]Binding{
		"DATA":   NewImmutableBinding(NewStdDataMap()),
		"FILE":   NewImmutableBinding(NewStdFileMap()),
		"IO":     NewImmutableBinding(NewStdIOMap()),
		"MATH":   NewImmutableBinding(NewStdMathMap()),
		"PATH":   NewImmutableBinding(NewStdPathMap()),
		"SYSTEM": NewImmutableBinding(NewStdSystemMap(args)),
		"TIME":   NewImmutableBinding(NewStdTimeMap()),
		"TYPE":   NewImmutableBinding(NewStdTypeMap()),
	}

	return NewMapValue(entries, true)
}
