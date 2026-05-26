package main

func NewStdDataMap() Value {
	entries := map[string]Binding{
		"CSV":  NewImmutableBinding(NewStdCSVMap()),
		"JSON": NewImmutableBinding(NewStdJSONMap()),
	}

	return NewMapValue(entries, true)
}
