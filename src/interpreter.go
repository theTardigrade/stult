package main

type Interpreter struct {
	Env *Environment
}

func NewInterpreter() *Interpreter {
	return NewInterpreterWithArgs(nil)
}

func NewInterpreterWithArgs(args []string) *Interpreter {
	env := NewEnvironment()

	if err := env.Set("STD", NewStdMap(args), true); err != nil {
		panic(err)
	}

	return &Interpreter{Env: env}
}
