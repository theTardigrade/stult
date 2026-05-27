package main

type Interpreter struct {
	Env *Environment
}

func NewInterpreter() *Interpreter {
	env := NewEnvironment()

	if err := env.Set("STD", NewStdMap(), true); err != nil {
		panic(err)
	}

	return &Interpreter{Env: env}
}
