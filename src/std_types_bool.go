package main

func NewStdTypesBoolMap() Value {
	entries := map[string]Binding{
		"TRUE":  NewImmutableBinding(NewBoolValue(true)),
		"FALSE": NewImmutableBinding(NewBoolValue(false)),
	}

	return NewMapValue(entries, true)
}
