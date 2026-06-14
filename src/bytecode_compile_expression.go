package main

import "fmt"

func (compiler *BytecodeCompiler) compileExpression(expression Expression) error {
	switch expression := expression.(type) {
	case *VoidLiteral:
		compiler.chunk.EmitAt(BytecodeOpLoadVoid, compiler.sourceSpanFromToken(expression.Token))
		return nil

	case *BoolLiteral:
		if expression.Value {
			compiler.chunk.EmitAt(BytecodeOpLoadTrue, compiler.sourceSpanFromToken(expression.Token))
		} else {
			compiler.chunk.EmitAt(BytecodeOpLoadFalse, compiler.sourceSpanFromToken(expression.Token))
		}

		return nil

	case *NumberLiteral:
		value, err := NewNumberValueFromString(expression.Value)
		if err != nil {
			return compiler.compileError(expression.Token, err.Error())
		}

		constant := compiler.chunk.AddConstant(value)
		compiler.chunk.EmitOperandAt(
			BytecodeOpLoadConst,
			constant,
			compiler.sourceSpanFromToken(expression.Token),
		)
		return nil

	case *StringLiteral:
		constant := compiler.chunk.AddConstant(NewStringValue(expression.Value))
		compiler.chunk.EmitOperandAt(
			BytecodeOpLoadConst,
			constant,
			compiler.sourceSpanFromToken(expression.Token),
		)
		return nil

	case *IdentifierExpression:
		return compiler.compileIdentifierExpression(expression)

	case *ArrayLiteral:
		return compiler.compileArrayLiteral(expression)

	case *MapLiteral:
		return compiler.compileMapLiteral(expression)

	case *FunctionLiteral:
		return compiler.compileFunctionLiteral(expression)

	case *IndexExpression:
		if err := compiler.compileExpression(expression.Object); err != nil {
			return err
		}

		if err := compiler.compileExpression(expression.Index); err != nil {
			return err
		}

		compiler.chunk.EmitAt(BytecodeOpIndex, compiler.sourceSpanForExpression(expression))
		return nil

	case *CallExpression:
		if err := compiler.compileExpression(expression.Callee); err != nil {
			return err
		}

		for _, argument := range expression.Arguments {
			if err := compiler.compileExpression(argument); err != nil {
				return err
			}
		}

		compiler.chunk.EmitOperandAt(
			BytecodeOpCall,
			len(expression.Arguments),
			compiler.sourceSpanForExpression(expression),
		)
		return nil

	case *PrefixExpression:
		if err := compiler.compileExpression(expression.Right); err != nil {
			return err
		}

		opcode, ok := bytecodeOpcodeForPrefixOperator(expression.Operator)
		if !ok {
			return compiler.compileError(
				expression.Token,
				fmt.Sprintf("bytecode compiler does not support prefix operator %q", expression.Operator),
			)
		}

		compiler.chunk.EmitAt(opcode, compiler.sourceSpanFromToken(expression.Token))
		return nil

	case *BinaryExpression:
		if expression.Operator == "&" || expression.Operator == "|" {
			return compiler.compileLogicalBinaryExpression(expression)
		}

		opcode, ok := bytecodeOpcodeForBinaryOperator(expression.Operator)
		if !ok {
			return fmt.Errorf(
				"%s: bytecode compiler does not support binary operator %q",
				formatBytecodeSourceSpan(compiler.sourceSpanForExpression(expression)),
				expression.Operator,
			)
		}

		if err := compiler.compileExpression(expression.Left); err != nil {
			return err
		}

		if err := compiler.compileExpression(expression.Right); err != nil {
			return err
		}

		compiler.chunk.EmitAt(opcode, compiler.sourceSpanForExpression(expression))
		return nil

	default:
		return fmt.Errorf(
			"%s: bytecode compiler does not know expression type %T",
			formatBytecodeSourceSpan(compiler.sourceSpanForExpression(expression)),
			expression,
		)
	}
}

func (compiler *BytecodeCompiler) compileIdentifierExpression(expression *IdentifierExpression) error {
	if expression.Name == "_" {
		compiler.chunk.EmitAt(BytecodeOpLoadVoid, compiler.sourceSpanFromToken(expression.Token))
		return nil
	}

	if expression.IsOuter {
		return compiler.compileOuterIdentifierExpression(expression)
	}

	if index, ok := compiler.localIndex(expression.Name); ok {
		compiler.chunk.EmitOperandAt(
			BytecodeOpLoadLocal,
			index,
			compiler.sourceSpanFromToken(expression.Token),
		)

		return nil
	}

	if index, ok := compiler.resolveUpvalue(expression.Name); ok {
		compiler.chunk.EmitOperandAt(
			BytecodeOpLoadUpvalue,
			index,
			compiler.sourceSpanFromToken(expression.Token),
		)

		return nil
	}

	name := compiler.chunk.AddNameConstant(expression.Name)
	compiler.chunk.EmitOperandAt(
		BytecodeOpLoadGlobal,
		name,
		compiler.sourceSpanFromToken(expression.Token),
	)

	return nil
}

func (compiler *BytecodeCompiler) compileOuterIdentifierExpression(
	expression *IdentifierExpression,
) error {
	if index, ok := compiler.outerLocalIndex(expression.Name); ok {
		compiler.chunk.EmitOperandAt(
			BytecodeOpLoadLocal,
			index,
			compiler.sourceSpanFromToken(expression.Token),
		)

		return nil
	}

	if index, ok := compiler.resolveUpvalue(expression.Name); ok {
		compiler.chunk.EmitOperandAt(
			BytecodeOpLoadUpvalue,
			index,
			compiler.sourceSpanFromToken(expression.Token),
		)

		return nil
	}

	if compiler.hasOuterContext() {
		name := compiler.chunk.AddNameConstant(expression.Name)
		compiler.chunk.EmitOperandAt(
			BytecodeOpLoadGlobal,
			name,
			compiler.sourceSpanFromToken(expression.Token),
		)

		return nil
	}

	return compiler.compileError(
		expression.Token,
		fmt.Sprintf("bytecode compiler could not resolve outer identifier %q", expression.Name),
	)
}

func (compiler *BytecodeCompiler) compileIdentifierStore(expression *IdentifierExpression) error {
	if expression.IsOuter {
		return compiler.compileOuterNameStore(expression.Token)
	}

	if compiler.shouldStorePlainNameAsLocal() {
		index := compiler.ensureLocal(expression.Name, expression.IsImmutable)
		opcode := BytecodeOpStoreLocalMutable

		if expression.IsImmutable {
			opcode = BytecodeOpStoreLocalImmutable
		}

		compiler.chunk.EmitOperandAt(
			opcode,
			index,
			compiler.sourceSpanFromToken(expression.Token),
		)

		return nil
	}

	name := compiler.chunk.AddNameConstant(expression.Name)
	opcode := BytecodeOpStoreGlobalMutable

	if expression.IsImmutable {
		opcode = BytecodeOpStoreGlobalImmutable
	}

	compiler.chunk.EmitOperandAt(
		opcode,
		name,
		compiler.sourceSpanFromToken(expression.Token),
	)

	return nil
}

func (compiler *BytecodeCompiler) compileOuterNameStore(token Token) error {
	if index, local, ok := compiler.outerLocalWithIndex(token.Literal); ok {
		opcode := BytecodeOpStoreLocalMutable

		if local.IsImmutable {
			opcode = BytecodeOpStoreLocalImmutable
		}

		compiler.chunk.EmitOperandAt(
			opcode,
			index,
			compiler.sourceSpanFromToken(token),
		)

		return nil
	}

	if index, upvalue, ok := compiler.resolveUpvalueWithValue(token.Literal); ok {
		opcode := BytecodeOpStoreUpvalueMutable

		if upvalue.IsImmutable {
			opcode = BytecodeOpStoreUpvalueImmutable
		}

		compiler.chunk.EmitOperandAt(
			opcode,
			index,
			compiler.sourceSpanFromToken(token),
		)

		return nil
	}

	if compiler.hasOuterContext() {
		name := compiler.chunk.AddNameConstant(token.Literal)
		opcode := BytecodeOpStoreGlobalMutable

		if token.IsImmutable {
			opcode = BytecodeOpStoreGlobalImmutable
		}

		compiler.chunk.EmitOperandAt(
			opcode,
			name,
			compiler.sourceSpanFromToken(token),
		)

		return nil
	}

	return compiler.compileError(
		token,
		fmt.Sprintf("bytecode compiler could not resolve outer assignment target %q", token.Literal),
	)
}

func (compiler *BytecodeCompiler) compileArrayLiteral(expression *ArrayLiteral) error {
	for _, element := range expression.Elements {
		switch element := element.(type) {
		case *ExpressionArrayElement:
			if err := compiler.compileExpression(element.Expression); err != nil {
				return err
			}

		case *RangeArrayElement:
			if err := compiler.compileRangeArrayElement(element); err != nil {
				return err
			}

		default:
			return fmt.Errorf(
				"%s: bytecode compiler does not know array element type %T",
				formatBytecodeSourceSpan(compiler.sourceSpanFromToken(expression.Token)),
				element,
			)
		}
	}

	compiler.chunk.EmitOperandAt(
		BytecodeOpBuildArray,
		len(expression.Elements),
		compiler.sourceSpanFromToken(expression.Token),
	)

	return nil
}

func (compiler *BytecodeCompiler) compileRangeArrayElement(element *RangeArrayElement) error {
	if err := compiler.compileExpression(element.Start); err != nil {
		return err
	}

	if err := compiler.compileExpression(element.End); err != nil {
		return err
	}

	if element.Step == nil {
		compiler.chunk.EmitAt(
			BytecodeOpLoadVoid,
			compiler.sourceSpanForExpression(element.End),
		)
	} else {
		if err := compiler.compileExpression(element.Step); err != nil {
			return err
		}
	}

	operand := 0
	if element.IsInclusive {
		operand = 1
	}

	compiler.chunk.EmitOperandAt(
		BytecodeOpBuildRange,
		operand,
		compiler.sourceSpanForExpression(element.Start),
	)

	return nil
}

func (compiler *BytecodeCompiler) compileMapLiteral(expression *MapLiteral) error {
	for _, entry := range expression.Entries {
		key := compiler.chunk.AddConstant(NewStringValue(entry.Key.Literal))
		compiler.chunk.EmitOperandAt(
			BytecodeOpLoadConst,
			key,
			compiler.sourceSpanFromToken(entry.Key),
		)

		if err := compiler.compileExpression(entry.Value); err != nil {
			return err
		}
	}

	compiler.chunk.EmitOperandAt(
		BytecodeOpBuildMap,
		len(expression.Entries),
		compiler.sourceSpanFromToken(expression.Token),
	)

	return nil
}

func (compiler *BytecodeCompiler) compileFunctionLiteral(expression *FunctionLiteral) error {
	functionIndex := len(compiler.chunk.Functions)
	functionName := fmt.Sprintf("%s:<function:%d>", compiler.functionPrefix, functionIndex)

	functionCompiler := NewBytecodeCompiler(functionName, compiler.filename, true, compiler)

	for _, parameter := range expression.Parameters {
		functionCompiler.defineParameterLocal(parameter)
	}

	if expression.VariadicParameter != nil {
		functionCompiler.defineParameterLocal(*expression.VariadicParameter)
	}

	if err := functionCompiler.compileStatementList(expression.Body); err != nil {
		return err
	}

	if len(expression.Returns) == 0 {
		functionCompiler.chunk.EmitAt(
			BytecodeOpLoadVoid,
			functionCompiler.sourceSpanFromToken(expression.Token),
		)
	} else {
		if len(expression.Returns) > 1 {
			return compiler.compileError(
				expression.Token,
				"bytecode compiler expected function literal to have exactly one return expression",
			)
		}

		if err := functionCompiler.compileExpression(expression.Returns[0]); err != nil {
			return err
		}
	}

	functionCompiler.chunk.EmitAt(
		BytecodeOpReturn,
		functionCompiler.sourceSpanFromToken(expression.Token),
	)

	function := BytecodeFunction{
		Name:              functionName,
		Parameters:        bytecodeParametersFromTokens(expression.Parameters),
		VariadicParameter: bytecodeVariadicParameterFromToken(expression.VariadicParameter),
		Upvalues:          append([]BytecodeUpvalue{}, functionCompiler.chunk.Upvalues...),
		Chunk:             functionCompiler.chunk,
	}

	index := compiler.chunk.AddFunction(function)

	compiler.chunk.EmitOperandAt(
		BytecodeOpMakeFunction,
		index,
		compiler.sourceSpanFromToken(expression.Token),
	)

	return nil
}

func (compiler *BytecodeCompiler) compileLogicalBinaryExpression(expression *BinaryExpression) error {
	if err := compiler.compileExpression(expression.Left); err != nil {
		return err
	}

	sourceSpan := compiler.sourceSpanForExpression(expression)

	switch expression.Operator {
	case "&":
		falseJump := compiler.chunk.EmitOperandAt(BytecodeOpJumpIfFalse, -1, sourceSpan)
		compiler.chunk.EmitAt(BytecodeOpPop, sourceSpan)

		if err := compiler.compileExpression(expression.Right); err != nil {
			return err
		}

		return compiler.patchJumpToCurrent(falseJump)

	case "|":
		trueJump := compiler.chunk.EmitOperandAt(BytecodeOpJumpIfTrue, -1, sourceSpan)
		compiler.chunk.EmitAt(BytecodeOpPop, sourceSpan)

		if err := compiler.compileExpression(expression.Right); err != nil {
			return err
		}

		return compiler.patchJumpToCurrent(trueJump)

	default:
		return fmt.Errorf(
			"%s: bytecode compiler does not support logical operator %q",
			formatBytecodeSourceSpan(sourceSpan),
			expression.Operator,
		)
	}
}
