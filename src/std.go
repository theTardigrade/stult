package main

func NewStdMap() Value {
	entries := map[string]Binding{
		"IO": {
			Value:       NewStdIOMap(),
			IsImmutable: true,
		},
		"MATH": {
			Value:       NewStdMathMap(),
			IsImmutable: true,
		},
	}

	return NewMapValue(entries, true)
}
