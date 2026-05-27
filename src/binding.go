package main

type Binding struct {
	Value       Value
	IsImmutable bool
}

func NewImmutableBinding(value Value) Binding {
	return Binding{
		Value:       value,
		IsImmutable: true,
	}
}
