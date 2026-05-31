package main

type RuntimeContext struct {
	Args []string
}

func NewRuntimeContext(args []string) *RuntimeContext {
	return &RuntimeContext{
		Args: append([]string{}, args...),
	}
}

type RuntimeMode int

const (
	RuntimeModeBytecode RuntimeMode = iota
	RuntimeModeInterpreter
)
