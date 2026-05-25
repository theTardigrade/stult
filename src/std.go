package main

func NewStdMap() Value {
	entries := map[string]Binding{
		"BOOL": {
			Value:       NewStdBoolMap(),
			IsImmutable: true,
		},
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
