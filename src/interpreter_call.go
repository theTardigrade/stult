package main

import "fmt"

func (i *Interpreter) evalCallExpression(call *CallExpression) (Value, error) {
	callee, err := i.evalExpression(call.Callee)
	if err != nil {
		return Value{}, err
	}

	args := make([]Value, 0, len(call.Arguments))

	for _, argument := range call.Arguments {
		arg, err := i.evalExpression(argument.Expression)
		if err != nil {
			return Value{}, err
		}

		if argument.IsSpread {
			spreadValues, err := callSpreadArgumentValues(arg)
			if err != nil {
				return Value{}, err
			}

			args = append(args, spreadValues...)
			continue
		}

		args = append(args, arg)
	}

	switch callee.Kind {
	case ValueBuiltinFunction:
		return callee.BuiltinFunction(i.Runtime, args)

	case ValueFunction:
		return i.callFunction(callee.Function, args)

	default:
		return Value{}, fmt.Errorf("value is not callable")
	}
}

func (i *Interpreter) callFunction(fn *Function, args []Value) (Value, error) {
	requiredCount := requiredFunctionParameterCount(fn.Parameters)
	maxCount := len(fn.Parameters)

	if len(args) < requiredCount {
		if fn.VariadicParameter == nil && requiredCount == maxCount {
			return Value{}, fmt.Errorf(
				"function expected %d argument(s), got %d",
				requiredCount,
				len(args),
			)
		}

		return Value{}, fmt.Errorf(
			"function expected at least %d argument(s), got %d",
			requiredCount,
			len(args),
		)
	}

	if fn.VariadicParameter == nil && len(args) > maxCount {
		if requiredCount == maxCount {
			return Value{}, fmt.Errorf(
				"function expected %d argument(s), got %d",
				maxCount,
				len(args),
			)
		}

		return Value{}, fmt.Errorf(
			"function expected at most %d argument(s), got %d",
			maxCount,
			len(args),
		)
	}

	callEnv := NewChildEnvironment(fn.Env)

	for index, parameter := range fn.Parameters {
		parameterToken := parameter.Token

		if parameterToken.Literal == "_" {
			continue
		}

		value := NewVoidValue()
		if index < len(args) {
			value = args[index]
		}

		if err := callEnv.Set(parameterToken.Literal, value, parameterToken.IsImmutable); err != nil {
			return Value{}, fmt.Errorf(
				"line %d, column %d: %w",
				parameterToken.StartOfLine,
				parameterToken.StartOfColumn,
				err,
			)
		}
	}

	if fn.VariadicParameter != nil && fn.VariadicParameter.Literal != "_" {
		variadicStart := len(fn.Parameters)
		if len(args) < variadicStart {
			variadicStart = len(args)
		}

		variadicValues := append([]Value{}, args[variadicStart:]...)

		if err := callEnv.Set(
			fn.VariadicParameter.Literal,
			NewArrayValue(variadicValues, false),
			fn.VariadicParameter.IsImmutable,
		); err != nil {
			return Value{}, fmt.Errorf(
				"line %d, column %d: %w",
				fn.VariadicParameter.StartOfLine,
				fn.VariadicParameter.StartOfColumn,
				err,
			)
		}
	}

	previousEnv := i.Env
	previousDotMap := i.currentDotMap
	i.Env = callEnv
	i.currentDotMap = fn.DotMap
	defer func() {
		i.Env = previousEnv
		i.currentDotMap = previousDotMap
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

func requiredFunctionParameterCount(parameters []FunctionParameter) int {
	count := 0

	for _, parameter := range parameters {
		if parameter.IsOptional {
			continue
		}

		count++
	}

	return count
}

func callSpreadArgumentValues(value Value) ([]Value, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueArray || value.Array == nil {
		return nil, fmt.Errorf("spread argument expected an array")
	}

	values := []Value{}

	if err := value.Array.ForEach(func(_ *Number, element Value) error {
		values = append(values, element)
		return nil
	}); err != nil {
		return nil, err
	}

	return values, nil
}
