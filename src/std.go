package main

func NewStdMap(runtime *RuntimeContext) Value {
	if runtime == nil {
		runtime = NewRuntimeContext(nil)
	}

	entries := map[string]Binding{
		"DATA":   NewImmutableBinding(NewStdDataMap()),
		"FILE":   NewImmutableBinding(NewStdFileMap()),
		"IO":     NewImmutableBinding(NewStdIOMap()),
		"MATH":   NewImmutableBinding(NewStdMathMap()),
		"PATH":   NewImmutableBinding(NewStdPathMap()),
		"SYSTEM": NewImmutableBinding(NewStdSystemMap(runtime)),
		"TIME":   NewImmutableBinding(NewStdTimeMap()),
		"TYPE":   NewImmutableBinding(NewStdTypeMap()),
	}

	return NewMapValue(entries, true)
}
