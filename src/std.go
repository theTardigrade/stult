package main

func NewStdMap(runtime *RuntimeContext) Value {
	if runtime == nil {
		runtime = NewRuntimeContext(nil)
	}

	entries := map[string]Binding{
		"ASSERT": NewImmutableBinding(NewStdAssertMap()),
		"DATA":   NewImmutableBinding(NewStdDataMap()),
		"FILE":   NewImmutableBinding(NewStdFileMap()),
		"IO":     NewImmutableBinding(NewStdIOMap()),
		"MATH":   NewImmutableBinding(NewStdMathMap()),
		"SYSTEM": NewImmutableBinding(NewStdSystemMap(runtime)),
		"TIME":   NewImmutableBinding(NewStdTimeMap()),
		"TYPE":   NewImmutableBinding(NewStdTypeMap()),
	}

	return NewMapValue(entries, true)
}
