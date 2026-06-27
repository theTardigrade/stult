package main

import (
	"fmt"
	"sort"
)

type Environment struct {
	values map[string]Binding
	parent *Environment
}

func NewEnvironment() *Environment {
	return NewChildEnvironment(nil)
}

func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		values: make(map[string]Binding),
		parent: parent,
	}
}

func (e *Environment) Get(name string) (Binding, bool) {
	for env := e; env != nil; env = env.parent {
		if binding, ok := env.values[name]; ok {
			return binding, true
		}
	}

	return Binding{}, false
}

func (e *Environment) GetOuter(name string) (Binding, bool) {
	for env := e.parent; env != nil; env = env.parent {
		if binding, ok := env.values[name]; ok {
			return binding, true
		}
	}

	return Binding{}, false
}

func (e *Environment) Set(name string, value Value, isImmutable bool) error {
	return e.SetWithContract(name, value, isImmutable, nil)
}

func (e *Environment) SetWithContract(
	name string,
	value Value,
	isImmutable bool,
	contractToken *Token,
) error {
	if existing, ok := e.values[name]; ok {
		if existing.IsImmutable {
			return fmt.Errorf("cannot reassign immutable constant %q", name)
		}

		if contractToken != nil {
			return fmt.Errorf("binding contract for %q can only be declared when the binding is created", name)
		}

		if err := existing.Contract.Check(name, value); err != nil {
			return err
		}

		e.values[name] = Binding{
			Value:       value,
			IsImmutable: existing.IsImmutable,
			Contract:    existing.Contract,
		}
		return nil
	}

	e.values[name] = Binding{
		Value:       value,
		IsImmutable: isImmutable,
		Contract:    bindingContractFromTokenPointer(contractToken, value),
	}

	return nil
}

func (e *Environment) SetOuter(name string, value Value) error {
	for env := e.parent; env != nil; env = env.parent {
		existing, ok := env.values[name]
		if !ok {
			continue
		}

		if existing.IsImmutable {
			return fmt.Errorf("cannot reassign immutable outer constant %q", name)
		}

		if err := existing.Contract.Check(name, value); err != nil {
			return err
		}

		env.values[name] = Binding{
			Value:       value,
			IsImmutable: existing.IsImmutable,
			Contract:    existing.Contract,
		}

		return nil
	}

	return fmt.Errorf("no outer binding named %q", name)
}

func (e *Environment) Dump() {
	keys := make([]string, 0, len(e.values))

	for name := range e.values {
		keys = append(keys, name)
	}

	sort.Strings(keys)

	for _, name := range keys {
		binding := e.values[name]

		mutability := "mutable"
		if binding.IsImmutable {
			mutability = "immutable"
		}

		fmt.Printf("%s = %s (%s)\n", name, binding.Value.String(), mutability)
	}
}

func bindingContractFromTokenPointer(token *Token, initialValue Value) BindingContract {
	if token == nil {
		return BindingContract{}
	}

	return BindingContractFromToken(*token, initialValue)
}
