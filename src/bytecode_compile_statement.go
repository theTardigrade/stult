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

	case *TryCatchStatement:
		return compiler.compileTryCatchStatement(statement)

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

	contract := bindingContractKindFromTokenPointer(statement.ContractToken)

	if compiler.shouldStorePlainNameAsLocal() {
		index := compiler.ensureLocalWithContract(statement.Name.Literal, statement.IsImmutable, BindingContract{})
		opcode := bytecodeLocalStoreOpcode(statement.IsImmutable, contract)

		compiler.chunk.EmitOperandAt(
			opcode,
			index,
			compiler.sourceSpanFromToken(statement.Name),
		)

		return nil
	}

	name := compiler.chunk.AddNameConstant(statement.Name.Literal)
	opcode := bytecodeGlobalStoreOpcode(statement.IsImmutable, contract)

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

func (compiler *BytecodeCompiler) compileTryCatchStatement(statement *TryCatchStatement) error {
	sourceSpan := compiler.sourceSpanFromToken(statement.Token)

	tryStart := compiler.chunk.EmitOperandAt(BytecodeOpTryStart, -1, sourceSpan)

	compiler.tryDepth++
	if err := compiler.compileScopedStatementList(statement.TryBody); err != nil {
		compiler.tryDepth--
		return err
	}
	compiler.tryDepth--

	compiler.chunk.EmitAt(BytecodeOpTryEnd, sourceSpan)
	afterCatchJump := compiler.chunk.EmitOperandAt(BytecodeOpJump, -1, sourceSpan)
	catchStart := len(compiler.chunk.Instructions)

	if err := compiler.chunk.PatchOperand(tryStart, catchStart); err != nil {
		return err
	}

	scopeDepth := compiler.beginLocalScope()
	compiler.chunk.EmitOperandAt(BytecodeOpResetLocals, scopeDepth, sourceSpan)

	if statement.CatchParameter == nil || statement.CatchParameter.Literal == "_" {
		compiler.chunk.EmitAt(BytecodeOpPop, sourceSpan)
	} else {
		localIndex := compiler.ensureLocal(
			statement.CatchParameter.Literal,
			statement.CatchParameter.IsImmutable,
		)

		opcode := BytecodeOpStoreLocalMutable
		if statement.CatchParameter.IsImmutable {
			opcode = BytecodeOpStoreLocalImmutable
		}

		compiler.chunk.EmitOperandAt(
			opcode,
			localIndex,
			compiler.sourceSpanFromToken(*statement.CatchParameter),
		)
	}

	if err := compiler.compileStatementList(statement.CatchBody); err != nil {
		compiler.endLocalScope()
		return err
	}

	compiler.endLocalScope()

	return compiler.patchJumpToCurrent(afterCatchJump)
}

func (compiler *BytecodeCompiler) compileBreakStatement(statement *BreakStatement) error {
	if len(compiler.loopBreakJumps) == 0 {
		return compiler.compileError(
			statement.Token,
			"bytecode compiler cannot compile break outside a loop",
		)
	}

	compiler.emitTryEndForBreak()

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
	if !compiler.isFunction {
		return compiler.compileError(
			statement.Token,
			"return used outside function",
		)
	}

	if err := compiler.compileExpression(statement.Value); err != nil {
		return err
	}

	compiler.emitTryEndForReturn()

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
	if directRange, ok := directRangeLoopElement(statement.Condition); ok && len(statement.RangeParameters) <= 2 {
		return compiler.compileDirectRangeLoopStatement(statement, directRange)
	}

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

	compiler.chunk.EmitOperandAt(
		BytecodeOpIteratorInit,
		len(statement.RangeParameters),
		sourceSpan,
	)

	return compiler.compileIteratorLoopBody(statement, sourceSpan)
}

func (compiler *BytecodeCompiler) compileDirectRangeLoopStatement(
	statement *LoopStatement,
	rangeElement *RangeArrayElement,
) error {
	if len(statement.RangeParameters) > 2 {
		return fmt.Errorf(
			"%s: bytecode compiler expected at most two direct range loop parameters",
			formatBytecodeSourceSpan(compiler.sourceSpanForExpression(statement.Condition)),
		)
	}

	if err := compiler.compileExpression(rangeElement.Start); err != nil {
		return err
	}

	if err := compiler.compileExpression(rangeElement.End); err != nil {
		return err
	}

	if rangeElement.Step == nil {
		compiler.chunk.EmitAt(
			BytecodeOpLoadVoid,
			compiler.sourceSpanForExpression(rangeElement.End),
		)
	} else {
		if err := compiler.compileExpression(rangeElement.Step); err != nil {
			return err
		}
	}

	sourceSpan := compiler.sourceSpanForExpression(statement.Condition)
	compiler.chunk.EmitOperandAt(
		BytecodeOpIteratorRangeInit,
		encodeIteratorRangeInitOperand(len(statement.RangeParameters), rangeElement.IsInclusive),
		sourceSpan,
	)

	if err := compiler.compileIteratorLoopBody(statement, sourceSpan); err != nil {
		return err
	}

	return compiler.compileScopedStatementList(statement.AfterLoopBody)
}

func (compiler *BytecodeCompiler) compileIteratorLoopBody(
	statement *LoopStatement,
	sourceSpan BytecodeSourceSpan,
) error {
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
	compiler.loopTryDepths = append(compiler.loopTryDepths, compiler.tryDepth)
}

func (compiler *BytecodeCompiler) endLoop() []int {
	if len(compiler.loopBreakJumps) == 0 {
		return nil
	}

	last := len(compiler.loopBreakJumps) - 1
	breakJumps := compiler.loopBreakJumps[last]
	compiler.loopBreakJumps = compiler.loopBreakJumps[:last]
	compiler.loopTryDepths = compiler.loopTryDepths[:last]

	return breakJumps
}

func (compiler *BytecodeCompiler) emitTryEndForBreak() {
	if len(compiler.loopTryDepths) == 0 {
		return
	}

	loopTryDepth := compiler.loopTryDepths[len(compiler.loopTryDepths)-1]
	for depth := compiler.tryDepth; depth > loopTryDepth; depth-- {
		compiler.chunk.Emit(BytecodeOpTryEnd)
	}
}

func (compiler *BytecodeCompiler) emitTryEndForReturn() {
	for depth := compiler.tryDepth; depth > 0; depth-- {
		compiler.chunk.Emit(BytecodeOpTryEnd)
	}
}

func bindingContractKindFromTokenPointer(token *Token) BindingContractKind {
	if token == nil || token.Type == TokenContractAny {
		return BindingContractAny
	}

	return BindingContractSameKind
}

func bytecodeGlobalStoreOpcode(
	isImmutable bool,
	contract BindingContractKind,
) BytecodeOpcode {
	if contract == BindingContractSameKind {
		if isImmutable {
			return BytecodeOpStoreGlobalImmutableSameKind
		}

		return BytecodeOpStoreGlobalMutableSameKind
	}

	if isImmutable {
		return BytecodeOpStoreGlobalImmutable
	}

	return BytecodeOpStoreGlobalMutable
}

func bytecodeLocalStoreOpcode(
	isImmutable bool,
	contract BindingContractKind,
) BytecodeOpcode {
	if contract == BindingContractSameKind {
		if isImmutable {
			return BytecodeOpStoreLocalImmutableSameKind
		}

		return BytecodeOpStoreLocalMutableSameKind
	}

	if isImmutable {
		return BytecodeOpStoreLocalImmutable
	}

	return BytecodeOpStoreLocalMutable
}
