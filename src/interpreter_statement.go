package main

import "fmt"

func (i *Interpreter) evalStatement(stmt Statement) (Value, error) {
	switch s := stmt.(type) {
	case *AssignmentStatement:
		value, err := i.evalExpression(s.Value)
		if err != nil {
			return Value{}, err
		}

		if s.IsOuter {
			if err := i.Env.SetOuter(s.Name.Literal, value); err != nil {
				return Value{}, fmt.Errorf("line %d, column %d: %w", s.Name.StartOfLine, s.Name.StartOfColumn, err)
			}
		} else {
			contractDeclaration := s.ContractDeclaration
			if contractDeclaration != nil {
				contract, err := i.resolveBindingContractAliases(contractDeclaration.Contract)
				if err != nil {
					return Value{}, fmt.Errorf("line %d, column %d: %w", s.Name.StartOfLine, s.Name.StartOfColumn, err)
				}
				contractDeclaration = &BindingContractDeclaration{Token: contractDeclaration.Token, Contract: contract}
			}

			if err := i.Env.SetWithContract(s.Name.Literal, value, s.IsImmutable, contractDeclaration); err != nil {
				return Value{}, fmt.Errorf("line %d, column %d: %w", s.Name.StartOfLine, s.Name.StartOfColumn, err)
			}
		}

		return value, nil

	case *CompoundAssignmentStatement:
		return i.evalCompoundAssignmentStatement(s)

	case *IndexAssignmentStatement:
		return i.evalIndexAssignmentStatement(s)

	case *ExpressionStatement:
		return i.evalExpression(s.Expression)

	case *BreakStatement:
		return NewVoidValue(), &controlFlow{Kind: controlFlowBreak}

	case *ReturnStatement:
		value, err := i.evalExpression(s.Value)
		if err != nil {
			return Value{}, err
		}

		return value, &controlFlow{
			Kind:  controlFlowReturn,
			Value: value,
		}

	case *ConditionalStatement:
		return i.evalConditionalStatement(s)

	case *LoopStatement:
		return i.evalLoopStatement(s)

	case *TryCatchStatement:
		return i.evalTryCatchStatement(s)

	default:
		return Value{}, fmt.Errorf("unknown statement type %T", stmt)
	}
}

func (i *Interpreter) evalConditionalStatement(stmt *ConditionalStatement) (Value, error) {
	for _, branch := range stmt.Branches {
		condition, err := i.evalExpression(branch.Condition)
		if err != nil {
			return Value{}, err
		}

		if condition.Kind != ValueBool {
			return Value{}, fmt.Errorf("conditional expression must evaluate to a bool")
		}

		if condition.Bool {
			return i.evalStatementBlock(branch.Body)
		}
	}

	if stmt.ElseBody != nil {
		return i.evalStatementBlock(stmt.ElseBody)
	}

	return NewVoidValue(), nil
}

func (i *Interpreter) evalStatementBlock(statements []Statement) (Value, error) {
	return i.evalStatementBlockWithBindings(statements, nil)
}

func (i *Interpreter) evalStatementBlockWithBindings(statements []Statement, initialBindings map[string]Binding) (Value, error) {
	blockEnv := NewChildEnvironment(i.Env)

	for name, binding := range initialBindings {
		blockEnv.values[name] = binding
	}

	previousEnv := i.Env
	i.Env = blockEnv
	defer func() {
		i.Env = previousEnv
	}()

	result := NewVoidValue()

	for _, stmt := range statements {
		value, err := i.evalStatement(stmt)
		if err != nil {
			return Value{}, err
		}

		result = value
	}

	return result, nil
}

func (i *Interpreter) evalTryCatchStatement(stmt *TryCatchStatement) (Value, error) {
	value, err := i.evalStatementBlock(stmt.TryBody)
	if err == nil {
		return value, nil
	}

	if _, ok := asControlFlow(err); ok {
		return Value{}, err
	}

	bindings := map[string]Binding{}
	if stmt.CatchParameter != nil && stmt.CatchParameter.Literal != "_" {
		bindings[stmt.CatchParameter.Literal] = Binding{
			Value:       NewStringValue(err.Error()),
			IsImmutable: stmt.CatchParameter.IsImmutable,
		}
	}

	return i.evalStatementBlockWithBindings(stmt.CatchBody, bindings)
}
