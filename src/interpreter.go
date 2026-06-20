package main

type Interpreter struct {
	Env     *Environment
	Runtime *RuntimeContext
	dotMaps []*Map
}

func NewInterpreter() *Interpreter {
	return NewInterpreterWithArgs(nil)
}

func NewInterpreterWithArgs(args []string) *Interpreter {
	return NewInterpreterWithRuntime(NewRuntimeContext(args))
}

func NewInterpreterWithRuntime(runtime *RuntimeContext) *Interpreter {
	if runtime == nil {
		runtime = NewRuntimeContext(nil)
	}

	env := NewEnvironment()

	if err := env.Set("STD", NewStdMap(runtime), true); err != nil {
		panic(err)
	}

	return &Interpreter{
		Env:     env,
		Runtime: runtime,
	}
}
