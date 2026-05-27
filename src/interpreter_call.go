package main

import "fmt"

func (i *Interpreter) evalCallExpression(call *CallExpression) (Value, error) {
	callee, err := i.evalExpression(call.Callee)
	if err != nil {
		return Value{}, err
	}

	args := make([]Value, 0, len(call.Arguments))

	for _, argExpr := range call.Arguments {
		arg, err := i.evalExpression(argExpr)
		if err != nil {
			return Value{}, err
		}

		args = append(args, arg)
	}

	switch callee.Kind {
	case ValueBuiltinFunction:
		return callee.BuiltinFunction(i, args)

	case ValueFunction:
		return i.callFunction(callee.Function, args)

	default:
		return Value{}, fmt.Errorf("value is not callable")
	}
}

func (i *Interpreter) callFunction(fn *Function, args []Value) (Value, error) {
	if len(args) != len(fn.Parameters) {
		return Value{}, fmt.Errorf(
			"function expected %d argument(s), got %d",
			len(fn.Parameters),
			len(args),
		)
	}

	callEnv := NewChildEnvironment(fn.Env)

	for index, parameter := range fn.Parameters {
		if parameter.Literal == "_" {
			continue
		}

		if err := callEnv.Set(parameter.Literal, args[index], parameter.IsImmutable); err != nil {
			return Value{}, fmt.Errorf(
				"line %d, column %d: %w",
				parameter.StartOfLine,
				parameter.StartOfColumn,
				err,
			)
		}
	}

	previousEnv := i.Env
	i.Env = callEnv
	defer func() {
		i.Env = previousEnv
	}()

	for _, stmt := range fn.Body {
		if _, err := i.evalStatement(stmt); err != nil {
			if flow, ok := asControlFlow(err); ok {
				switch flow.Kind {
				case controlFlowBreak:
					return Value{}, fmt.Errorf("break used outside loop")
				case controlFlowReturn:
					return flow.Value, nil
				}
			}

			return Value{}, err
		}
	}

	if len(fn.Returns) != 1 {
		return Value{}, fmt.Errorf("functions must return exactly one value for now")
	}

	return i.evalExpression(fn.Returns[0])
}
