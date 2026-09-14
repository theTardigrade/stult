package main

func NewStdNetMap(runtime *RuntimeContext) Value {
	if runtime == nil {
		runtime = NewRuntimeContext(nil)
	}

	entries := map[string]Binding{
		"HTTP": NewImmutableBinding(NewStdNetHTTPMap(runtime)),
	}

	return NewMapValue(entries, true)
}
