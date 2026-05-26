package main

func NewStdDataMap() Value {
	entries := map[string]Binding{
		"CSV":     NewImmutableBinding(NewStdCSVMap()),
		"JSON":    NewImmutableBinding(NewStdJSONMap()),
		"STULTON": NewImmutableBinding(NewStdDataStultonMap()),
	}

	return NewMapValue(entries, true)
}
