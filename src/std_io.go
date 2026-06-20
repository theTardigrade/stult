package main

func NewStdIOMap() Value {
	entries := map[string]Binding{
		"INPUT":  NewImmutableBinding(NewStdIOInputMap()),
		"OUTPUT": NewImmutableBinding(NewStdIOOutputMap()),
	}

	return NewMapValue(entries, true)
}
