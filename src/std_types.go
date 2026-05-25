package main

func NewStdTypesMap() Value {
	entries := map[string]Binding{
		"ARRAY":      NewImmutableBinding(NewStdTypesArrayMap()),
		"BOOL":       NewImmutableBinding(NewStdTypesBoolMap()),
		"COLLECTION": NewImmutableBinding(NewStdTypesCollectionMap()),
		"NUMBER":     NewImmutableBinding(NewStdTypesNumberMap()),
	}

	return NewMapValue(entries, true)
}
