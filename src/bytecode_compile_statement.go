package main

import "fmt"

func (compiler *BytecodeCompiler) compileStatementList(statements []Statement) error {
	for _, statement := range statements {
		if err := compiler.compileStatement(statement); err != nil {
			return err
		}
	}

	return nil
}

func (compiler *BytecodeCompiler) compileScopedStatementList(statements []Statement) error {
	scopeDepth := compiler.beginLocalScope()

	compiler.chunk.EmitOperand(BytecodeOpResetLocals, scopeDepth)

	if err := compiler.compileStatementList(statements); err != nil {
		compiler.endLocalScope()
		return err
	}

	compiler.endLocalScope()

	return nil
}

func (compiler *BytecodeCompiler) compileStatement(statement Statement) error {
	switch statement := statement.(type) {
	case *AssignmentStatement:
		return compiler.compileAssignmentStatement(statement)

	case *CompoundAssignmentStatement:
		return compiler.compileCompoundAssignmentStatement(statement)

	case *IndexAssignmentStatement:
		return compiler.compileIndexAssignmentStatement(statement)

	case *ExpressionStatement:
		if err := compiler.compileExpression(statement.Expression); err != nil {
			return err
		}

		compiler.chunk.EmitAt(BytecodeOpPop, compiler.sourceSpanForExpression(statement.Expression))
		return nil

	case *BreakStatement:
		return compiler.compileBreakStatement(statement)

	case *ReturnStatement:
		return compiler.compileReturnStatement(statement)

	case *ConditionalStatement:
		return compiler.compileConditionalStatement(statement)

	case *LoopStatement:
		return compiler.compileLoopStatement(statement)

	default:
		return fmt.Errorf(
			"bytecode compiler does not know statement type %T",
			statement,
		)
	}
}

func (compiler *BytecodeCompiler) compileAssignmentStatement(statement *AssignmentStatement) error {
	if err := compiler.compileExpression(statement.Value); err != nil {
		return err
	}

	if statement.Name.Literal == "_" {
		compiler.chunk.EmitAt(BytecodeOpPop, compiler.sourceSpanFromToken(statement.Name))
		return nil
	}

	if statement.IsOuter {
		return compiler.compileOuterNameStore(statement.Name)
	}

	if compiler.shouldStorePlainNameAsLocal() {
		index := compiler.ensureLocal(statement.Name.Literal, statement.IsImmutable)
		opcode := BytecodeOpStoreLocalMutable

		if statement.IsImmutable {
			opcode = BytecodeOpStoreLocalImmutable
		}

		compiler.chunk.EmitOperandAt(
			opcode,
			index,
			compiler.sourceSpanFromToken(statement.Name),
		)

		return nil
	}

	name := compiler.chunk.AddNameConstant(statement.Name.Literal)
	opcode := BytecodeOpStoreGlobalMutable

	if statement.IsImmutable {
		opcode = BytecodeOpStoreGlobalImmutable
	}

	compiler.chunk.EmitOperandAt(
		opcode,
		name,
		compiler.sourceSpanFromToken(statement.Name),
	)

	return nil
}

func (compiler *BytecodeCompiler) compileCompoundAssignmentStatement(statement *CompoundAssignmentStatement) error {
	opcode, ok := bytecodeOpcodeForCompoundAssignmentOperator(statement.Operator.Type)
	if !ok {
		return compiler.compileError(
			statement.Operator,
			fmt.Sprintf("bytecode compiler does not support compound assignment operator %q", statement.Operator.Literal),
		)
	}

	switch target := statement.Target.(type) {
	case *IdentifierExpression:
		if err := compiler.compileExpression(target); err != nil {
			return err
		}

		if err := compiler.compileExpression(statement.Value); err != nil {
			return err
		}

		compiler.chunk.EmitAt(opcode, compiler.sourceSpanFromToken(statement.Operator))

		if target.Name == "_" {
			compiler.chunk.EmitAt(BytecodeOpPop, compiler.sourceSpanFromToken(target.Token))
			return nil
		}

		return compiler.compileIdentifierStore(target)

	case *IndexExpression:
		if err := compiler.compileExpression(target.Object); err != nil {
			return err
		}

		if err := compiler.compileExpression(target.Index); err != nil {
			return err
		}

		compiler.chunk.EmitAt(
			BytecodeOpDuplicateTopTwo,
			compiler.sourceSpanForExpression(target),
		)

		compiler.chunk.EmitAt(
			BytecodeOpIndex,
			compiler.sourceSpanForExpression(target),
		)

		if err := compiler.compileExpression(statement.Value); err != nil {
			return err
		}

		compiler.chunk.EmitAt(opcode, compiler.sourceSpanFromToken(statement.Operator))
		compiler.chunk.EmitAt(BytecodeOpStoreIndex, compiler.sourceSpanForExpression(target))

		return nil

	default:
		return fmt.Errorf(
			"%s: bytecode compiler does not know compound assignment target %T",
			formatBytecodeSourceSpan(compiler.sourceSpanForExpression(statement.Target)),
			statement.Target,
		)
	}
}

func (compiler *BytecodeCompiler) compileIndexAssignmentStatement(statement *IndexAssignmentStatement) error {
	if err := compiler.compileExpression(statement.Target.Object); err != nil {
		return err
	}

	if err := compiler.compileExpression(statement.Target.Index); err != nil {
		return err
	}

	if err := compiler.compileExpression(statement.Value); err != nil {
		return err
	}

	compiler.chunk.EmitAt(
		BytecodeOpStoreIndex,
		compiler.sourceSpanForExpression(statement.Target),
	)

	return nil
}

func (compiler *BytecodeCompiler) compileBreakStatement(statement *BreakStatement) error {
	if len(compiler.loopBreakJumps) == 0 {
		return compiler.compileError(
			statement.Token,
			"bytecode compiler cannot compile break outside a loop",
		)
	}

	jump := compiler.chunk.EmitOperandAt(
		BytecodeOpJump,
		-1,
		compiler.sourceSpanFromToken(statement.Token),
	)

	currentLoop := len(compiler.loopBreakJumps) - 1
	compiler.loopBreakJumps[currentLoop] = append(compiler.loopBreakJumps[currentLoop], jump)

	return nil
}

func (compiler *BytecodeCompiler) compileReturnStatement(statement *ReturnStatement) error {
	if err := compiler.compileExpression(statement.Value); err != nil {
		return err
	}

	compiler.chunk.EmitAt(
		BytecodeOpReturn,
		compiler.sourceSpanFromToken(statement.Token),
	)

	return nil
}

func (compiler *BytecodeCompiler) compileConditionalStatement(statement *ConditionalStatement) error {
	endJumps := []int{}

	for _, branch := range statement.Branches {
		if err := compiler.compileExpression(branch.Condition); err != nil {
			return err
		}

		sourceSpan := compiler.sourceSpanForExpression(branch.Condition)
		falseJump := compiler.chunk.EmitOperandAt(BytecodeOpJumpIfFalse, -1, sourceSpan)

		compiler.chunk.EmitAt(BytecodeOpPop, sourceSpan)

		if err := compiler.compileScopedStatementList(branch.Body); err != nil {
			return err
		}

		endJump := compiler.chunk.EmitOperandAt(BytecodeOpJump, -1, sourceSpan)
		endJumps = append(endJumps, endJump)

		if err := compiler.patchJumpToCurrent(falseJump); err != nil {
			return err
		}

		compiler.chunk.EmitAt(BytecodeOpPop, sourceSpan)
	}

	if err := compiler.compileScopedStatementList(statement.ElseBody); err != nil {
		return err
	}

	for _, endJump := range endJumps {
		if err := compiler.patchJumpToCurrent(endJump); err != nil {
			return err
		}
	}

	return nil
}

func (compiler *BytecodeCompiler) compileLoopStatement(statement *LoopStatement) error {
	if len(statement.RangeParameters) > 0 {
		return compiler.compileCollectionLoopStatement(statement)
	}

	return compiler.compileDynamicLoopStatement(statement)
}

func (compiler *BytecodeCompiler) compileDynamicLoopStatement(statement *LoopStatement) error {
	loopStart := len(compiler.chunk.Instructions)

	if err := compiler.compileExpression(statement.Condition); err != nil {
		return err
	}

	sourceSpan := compiler.sourceSpanForExpression(statement.Condition)
	collectionJump := compiler.chunk.EmitOperandAt(BytecodeOpJumpIfCollection, -1, sourceSpan)
	exitJump := compiler.chunk.EmitOperandAt(BytecodeOpJumpIfFalse, -1, sourceSpan)

	compiler.chunk.EmitAt(BytecodeOpPop, sourceSpan)

	compiler.beginLoop()

	if err := compiler.compileScopedStatementList(statement.Body); err != nil {
		compiler.endLoop()
		return err
	}

	whileBreakJumps := compiler.endLoop()

	compiler.chunk.EmitOperandAt(BytecodeOpJump, loopStart, sourceSpan)

	falseConditionPop := len(compiler.chunk.Instructions)

	if err := compiler.chunk.PatchOperand(exitJump, falseConditionPop); err != nil {
		return err
	}

	compiler.chunk.EmitAt(BytecodeOpPop, sourceSpan)

	skipCollectionJump := compiler.chunk.EmitOperandAt(BytecodeOpJump, -1, sourceSpan)
	collectionStart := len(compiler.chunk.Instructions)

	if err := compiler.chunk.PatchOperand(collectionJump, collectionStart); err != nil {
		return err
	}

	if err := compiler.compileCollectionLoopWithCollectionOnStack(statement, sourceSpan); err != nil {
		return err
	}

	afterLoopStart := len(compiler.chunk.Instructions)

	if err := compiler.chunk.PatchOperand(skipCollectionJump, afterLoopStart); err != nil {
		return err
	}

	for _, breakJump := range whileBreakJumps {
		if err := compiler.chunk.PatchOperand(breakJump, afterLoopStart); err != nil {
			return err
		}
	}

	return compiler.compileScopedStatementList(statement.AfterLoopBody)
}

func (compiler *BytecodeCompiler) compileCollectionLoopStatement(statement *LoopStatement) error {
	if err := compiler.compileExpression(statement.Condition); err != nil {
		return err
	}

	sourceSpan := compiler.sourceSpanForExpression(statement.Condition)

	if err := compiler.compileCollectionLoopWithCollectionOnStack(statement, sourceSpan); err != nil {
		return err
	}

	return compiler.compileScopedStatementList(statement.AfterLoopBody)
}

func (compiler *BytecodeCompiler) compileCollectionLoopWithCollectionOnStack(
	statement *LoopStatement,
	sourceSpan BytecodeSourceSpan,
) error {
	if len(statement.RangeParameters) > 4 {
		return fmt.Errorf(
			"%s: bytecode compiler expected at most four collection loop parameters",
			formatBytecodeSourceSpan(sourceSpan),
		)
	}

	compiler.chunk.EmitAt(BytecodeOpIteratorInit, sourceSpan)

	loopStart := len(compiler.chunk.Instructions)
	nextJump := compiler.chunk.EmitOperandAt(BytecodeOpIteratorNext, -1, sourceSpan)

	compiler.beginLoop()

	scopeDepth := compiler.beginLocalScope()
	compiler.chunk.EmitOperandAt(BytecodeOpResetLocals, scopeDepth, sourceSpan)

	rangeLocals := compiler.defineRangeParameterLocals(statement.RangeParameters)
	compiler.emitRangeParameterBindings(statement.RangeParameters, rangeLocals)

	if err := compiler.compileStatementList(statement.Body); err != nil {
		compiler.endLocalScope()
		compiler.endLoop()
		return err
	}

	compiler.endLocalScope()
	breakJumps := compiler.endLoop()

	compiler.chunk.EmitOperandAt(BytecodeOpJump, loopStart, sourceSpan)

	iteratorEnd := len(compiler.chunk.Instructions)

	if err := compiler.chunk.PatchOperand(nextJump, iteratorEnd); err != nil {
		return err
	}

	for _, breakJump := range breakJumps {
		if err := compiler.chunk.PatchOperand(breakJump, iteratorEnd); err != nil {
			return err
		}
	}

	compiler.chunk.EmitAt(BytecodeOpIteratorEnd, sourceSpan)

	return nil
}

func (compiler *BytecodeCompiler) beginLoop() {
	compiler.loopBreakJumps = append(compiler.loopBreakJumps, []int{})
}

func (compiler *BytecodeCompiler) endLoop() []int {
	if len(compiler.loopBreakJumps) == 0 {
		return nil
	}

	last := len(compiler.loopBreakJumps) - 1
	breakJumps := compiler.loopBreakJumps[last]
	compiler.loopBreakJumps = compiler.loopBreakJumps[:last]

	return breakJumps
}
