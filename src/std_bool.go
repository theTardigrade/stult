package main

func NewStdBoolMap() Value {
	entries := map[string]Binding{
		"TRUE": {
			Value:       NewBoolValue(true),
			IsImmutable: true,
		},
		"FALSE": {
			Value:       NewBoolValue(false),
			IsImmutable: true,
		},
	}

	return NewMapValue(entries, true)
}
