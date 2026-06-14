package main

import (
	"fmt"
)

type BytecodeOpcode int

const (
	BytecodeOpLoadVoid BytecodeOpcode = iota
	BytecodeOpLoadConst
	BytecodeOpLoadTrue
	BytecodeOpLoadFalse
	BytecodeOpLoadGlobal
	BytecodeOpLoadLocal
	BytecodeOpLoadUpvalue

	BytecodeOpStoreGlobalMutable
	BytecodeOpStoreGlobalImmutable
	BytecodeOpStoreLocalMutable
	BytecodeOpStoreLocalImmutable
	BytecodeOpStoreUpvalueMutable
	BytecodeOpStoreUpvalueImmutable
	BytecodeOpStoreIndex

	BytecodeOpBuildArray
	BytecodeOpBuildMap
	BytecodeOpBuildRange
	BytecodeOpMakeFunction
	BytecodeOpIndex
	BytecodeOpCall

	BytecodeOpIteratorInit
	BytecodeOpIteratorNext
	BytecodeOpIteratorValue
	BytecodeOpIteratorKey
	BytecodeOpIteratorCollection
	BytecodeOpIteratorPosition
	BytecodeOpIteratorEnd

	BytecodeOpResetLocals

	BytecodeOpJump
	BytecodeOpJumpIfFalse
	BytecodeOpJumpIfTrue
	BytecodeOpJumpIfCollection

	BytecodeOpDuplicateTopTwo

	BytecodeOpNegate
	BytecodeOpNot

	BytecodeOpAdd
	BytecodeOpSubtract
	BytecodeOpMultiply
	BytecodeOpDivide

	BytecodeOpEqual
	BytecodeOpNotEqual
	BytecodeOpLess
	BytecodeOpLessEqual
	BytecodeOpGreater
	BytecodeOpGreaterEqual

	BytecodeOpPop
	BytecodeOpReturn
)

type BytecodeInstruction struct {
	Opcode  BytecodeOpcode
	Operand int
}

type BytecodeSourceSpan struct {
	Filename string
	Line     int
	Column   int
}

type BytecodeParameter struct {
	Name        string
	IsImmutable bool
}

type BytecodeLocal struct {
	Name        string
	IsImmutable bool
	ScopeDepth  int
}

type BytecodeUpvalue struct {
	Name        string
	IsImmutable bool
	IsLocal     bool
	Index       int
}

type BytecodeFunction struct {
	Name              string
	Parameters        []BytecodeParameter
	VariadicParameter *BytecodeParameter
	Upvalues          []BytecodeUpvalue
	Chunk             *BytecodeChunk
}

type BytecodeChunk struct {
	Name         string
	Instructions []BytecodeInstruction
	Constants    []Value
	Locals       []BytecodeLocal
	Upvalues     []BytecodeUpvalue
	Functions    []BytecodeFunction
	SourceSpans  []BytecodeSourceSpan
}

func NewBytecodeChunk(name string) *BytecodeChunk {
	return &BytecodeChunk{
		Name:         name,
		Instructions: []BytecodeInstruction{},
		Constants:    []Value{},
		Locals:       []BytecodeLocal{},
		Upvalues:     []BytecodeUpvalue{},
		Functions:    []BytecodeFunction{},
		SourceSpans:  []BytecodeSourceSpan{},
	}
}

func (chunk *BytecodeChunk) AddConstant(value Value) int {
	chunk.Constants = append(chunk.Constants, value)

	return len(chunk.Constants) - 1
}

func (chunk *BytecodeChunk) AddNameConstant(name string) int {
	return chunk.AddConstant(NewStringValue(name))
}

func (chunk *BytecodeChunk) AddLocal(name string, isImmutable bool, scopeDepth int) int {
	chunk.Locals = append(chunk.Locals, BytecodeLocal{
		Name:        name,
		IsImmutable: isImmutable,
		ScopeDepth:  scopeDepth,
	})

	return len(chunk.Locals) - 1
}

func (chunk *BytecodeChunk) AddUpvalue(
	name string,
	isImmutable bool,
	isLocal bool,
	index int,
) int {
	chunk.Upvalues = append(chunk.Upvalues, BytecodeUpvalue{
		Name:        name,
		IsImmutable: isImmutable,
		IsLocal:     isLocal,
		Index:       index,
	})

	return len(chunk.Upvalues) - 1
}

func (chunk *BytecodeChunk) AddFunction(function BytecodeFunction) int {
	chunk.Functions = append(chunk.Functions, function)

	return len(chunk.Functions) - 1
}

func (chunk *BytecodeChunk) Emit(opcode BytecodeOpcode) int {
	return chunk.EmitAt(opcode, EmptyBytecodeSourceSpan())
}

func (chunk *BytecodeChunk) EmitOperand(opcode BytecodeOpcode, operand int) int {
	return chunk.EmitOperandAt(opcode, operand, EmptyBytecodeSourceSpan())
}

func (chunk *BytecodeChunk) EmitAt(opcode BytecodeOpcode, sourceSpan BytecodeSourceSpan) int {
	return chunk.EmitOperandAt(opcode, 0, sourceSpan)
}

func (chunk *BytecodeChunk) EmitOperandAt(
	opcode BytecodeOpcode,
	operand int,
	sourceSpan BytecodeSourceSpan,
) int {
	chunk.Instructions = append(chunk.Instructions, BytecodeInstruction{
		Opcode:  opcode,
		Operand: operand,
	})

	chunk.SourceSpans = append(chunk.SourceSpans, sourceSpan)

	return len(chunk.Instructions) - 1
}

func (chunk *BytecodeChunk) PatchOperand(instructionIndex int, operand int) error {
	if instructionIndex < 0 || instructionIndex >= len(chunk.Instructions) {
		return fmt.Errorf("cannot patch bytecode instruction %d", instructionIndex)
	}

	chunk.Instructions[instructionIndex].Operand = operand

	return nil
}

func EmptyBytecodeSourceSpan() BytecodeSourceSpan {
	return BytecodeSourceSpan{}
}

func (opcode BytecodeOpcode) String() string {
	switch opcode {
	case BytecodeOpLoadVoid:
		return "LOAD_VOID"

	case BytecodeOpLoadConst:
		return "LOAD_CONST"

	case BytecodeOpLoadTrue:
		return "LOAD_TRUE"

	case BytecodeOpLoadFalse:
		return "LOAD_FALSE"

	case BytecodeOpLoadGlobal:
		return "LOAD_GLOBAL"

	case BytecodeOpLoadLocal:
		return "LOAD_LOCAL"

	case BytecodeOpLoadUpvalue:
		return "LOAD_UPVALUE"

	case BytecodeOpStoreGlobalMutable:
		return "STORE_GLOBAL_MUTABLE"

	case BytecodeOpStoreGlobalImmutable:
		return "STORE_GLOBAL_IMMUTABLE"

	case BytecodeOpStoreLocalMutable:
		return "STORE_LOCAL_MUTABLE"

	case BytecodeOpStoreLocalImmutable:
		return "STORE_LOCAL_IMMUTABLE"

	case BytecodeOpStoreUpvalueMutable:
		return "STORE_UPVALUE_MUTABLE"

	case BytecodeOpStoreUpvalueImmutable:
		return "STORE_UPVALUE_IMMUTABLE"

	case BytecodeOpStoreIndex:
		return "STORE_INDEX"

	case BytecodeOpBuildArray:
		return "BUILD_ARRAY"

	case BytecodeOpBuildMap:
		return "BUILD_MAP"

	case BytecodeOpBuildRange:
		return "BUILD_RANGE"

	case BytecodeOpMakeFunction:
		return "MAKE_FUNCTION"

	case BytecodeOpIndex:
		return "INDEX"

	case BytecodeOpCall:
		return "CALL"

	case BytecodeOpIteratorInit:
		return "ITERATOR_INIT"

	case BytecodeOpIteratorNext:
		return "ITERATOR_NEXT"

	case BytecodeOpIteratorValue:
		return "ITERATOR_VALUE"

	case BytecodeOpIteratorKey:
		return "ITERATOR_KEY"

	case BytecodeOpIteratorCollection:
		return "ITERATOR_COLLECTION"

	case BytecodeOpIteratorPosition:
		return "ITERATOR_POSITION"

	case BytecodeOpIteratorEnd:
		return "ITERATOR_END"

	case BytecodeOpResetLocals:
		return "RESET_LOCALS"

	case BytecodeOpJump:
		return "JUMP"

	case BytecodeOpJumpIfFalse:
		return "JUMP_IF_FALSE"

	case BytecodeOpJumpIfTrue:
		return "JUMP_IF_TRUE"

	case BytecodeOpJumpIfCollection:
		return "JUMP_IF_COLLECTION"

	case BytecodeOpDuplicateTopTwo:
		return "DUPLICATE_TOP_TWO"

	case BytecodeOpNegate:
		return "NEGATE"

	case BytecodeOpNot:
		return "NOT"

	case BytecodeOpAdd:
		return "ADD"

	case BytecodeOpSubtract:
		return "SUBTRACT"

	case BytecodeOpMultiply:
		return "MULTIPLY"

	case BytecodeOpDivide:
		return "DIVIDE"

	case BytecodeOpEqual:
		return "EQUAL"

	case BytecodeOpNotEqual:
		return "NOT_EQUAL"

	case BytecodeOpLess:
		return "LESS"

	case BytecodeOpLessEqual:
		return "LESS_EQUAL"

	case BytecodeOpGreater:
		return "GREATER"

	case BytecodeOpGreaterEqual:
		return "GREATER_EQUAL"

	case BytecodeOpPop:
		return "POP"

	case BytecodeOpReturn:
		return "RETURN"

	default:
		return "UNKNOWN"
	}
}

func CompileBytecode(program *Program, name string) (*BytecodeChunk, error) {
	compiler := NewBytecodeCompiler(name, name, false, nil)

	if err := compiler.compileProgram(program); err != nil {
		return nil, err
	}

	return compiler.chunk, nil
}

type BytecodeCompiler struct {
	chunk          *BytecodeChunk
	filename       string
	functionPrefix string
	isFunction     bool
	parent         *BytecodeCompiler
	localScopes    []map[string]int
	upvalueIndexes map[string]int
	loopBreakJumps [][]int
}

func NewBytecodeCompiler(
	name string,
	filename string,
	isFunction bool,
	parent *BytecodeCompiler,
) *BytecodeCompiler {
	return &BytecodeCompiler{
		chunk:          NewBytecodeChunk(name),
		filename:       filename,
		functionPrefix: name,
		isFunction:     isFunction,
		parent:         parent,
		localScopes:    []map[string]int{{}},
		upvalueIndexes: map[string]int{},
		loopBreakJumps: [][]int{},
	}
}

func (compiler *BytecodeCompiler) compileProgram(program *Program) error {
	if program == nil {
		return fmt.Errorf("bytecode compiler cannot compile a nil program")
	}

	if err := compiler.compileStatementList(program.Statements); err != nil {
		return err
	}

	compiler.chunk.Emit(BytecodeOpLoadVoid)
	compiler.chunk.Emit(BytecodeOpReturn)

	return nil
}

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

func (compiler *BytecodeCompiler) patchJumpToCurrent(instructionIndex int) error {
	return compiler.chunk.PatchOperand(instructionIndex, len(compiler.chunk.Instructions))
}

func (compiler *BytecodeCompiler) defineParameterLocal(token Token) {
	if token.Literal == "_" {
		return
	}

	compiler.ensureLocal(token.Literal, token.IsImmutable)
}

func (compiler *BytecodeCompiler) defineRangeParameterLocals(parameters []Token) []int {
	locals := make([]int, len(parameters))

	for index, parameter := range parameters {
		if parameter.Literal == "_" {
			locals[index] = -1
			continue
		}

		locals[index] = compiler.ensureLocal(parameter.Literal, parameter.IsImmutable)
	}

	return locals
}

func (compiler *BytecodeCompiler) emitRangeParameterBindings(parameters []Token, locals []int) {
	for index, local := range locals {
		if local < 0 {
			continue
		}

		opcode := BytecodeOpIteratorValue

		switch index {
		case 0:
			opcode = BytecodeOpIteratorValue

		case 1:
			opcode = BytecodeOpIteratorKey

		case 2:
			opcode = BytecodeOpIteratorCollection

		case 3:
			opcode = BytecodeOpIteratorPosition
		}

		compiler.chunk.EmitOperandAt(
			opcode,
			local,
			compiler.sourceSpanFromToken(parameters[index]),
		)
	}
}

func (compiler *BytecodeCompiler) beginLocalScope() int {
	compiler.localScopes = append(compiler.localScopes, map[string]int{})

	return len(compiler.localScopes) - 1
}

func (compiler *BytecodeCompiler) endLocalScope() {
	if len(compiler.localScopes) <= 1 {
		return
	}

	compiler.localScopes = compiler.localScopes[:len(compiler.localScopes)-1]
}

func (compiler *BytecodeCompiler) hasOuterContext() bool {
	return compiler.parent != nil || len(compiler.localScopes) > 1
}

func (compiler *BytecodeCompiler) shouldStorePlainNameAsLocal() bool {
	return compiler.isFunction || len(compiler.localScopes) > 1
}

func (compiler *BytecodeCompiler) ensureLocal(name string, isImmutable bool) int {
	if index, ok := compiler.currentLocalIndex(name); ok {
		return index
	}

	scope := compiler.currentLocalScope()
	scopeDepth := len(compiler.localScopes) - 1
	index := compiler.chunk.AddLocal(name, isImmutable, scopeDepth)

	scope[name] = index

	return index
}

func (compiler *BytecodeCompiler) currentLocalScope() map[string]int {
	if len(compiler.localScopes) == 0 {
		compiler.localScopes = append(compiler.localScopes, map[string]int{})
	}

	return compiler.localScopes[len(compiler.localScopes)-1]
}

func (compiler *BytecodeCompiler) currentLocalIndex(name string) (int, bool) {
	scope := compiler.currentLocalScope()
	index, ok := scope[name]

	return index, ok
}

func (compiler *BytecodeCompiler) localIndex(name string) (int, bool) {
	for index := len(compiler.localScopes) - 1; index >= 0; index-- {
		localIndex, ok := compiler.localScopes[index][name]
		if ok {
			return localIndex, true
		}
	}

	return 0, false
}

func (compiler *BytecodeCompiler) outerLocalIndex(name string) (int, bool) {
	for index := len(compiler.localScopes) - 2; index >= 0; index-- {
		localIndex, ok := compiler.localScopes[index][name]
		if ok {
			return localIndex, true
		}
	}

	return 0, false
}

func (compiler *BytecodeCompiler) outerLocalWithIndex(name string) (int, BytecodeLocal, bool) {
	index, ok := compiler.outerLocalIndex(name)
	if !ok {
		return 0, BytecodeLocal{}, false
	}

	if index < 0 || index >= len(compiler.chunk.Locals) {
		return 0, BytecodeLocal{}, false
	}

	return index, compiler.chunk.Locals[index], true
}

func (compiler *BytecodeCompiler) resolveUpvalue(name string) (int, bool) {
	index, _, ok := compiler.resolveUpvalueWithValue(name)

	return index, ok
}

func (compiler *BytecodeCompiler) resolveUpvalueWithValue(name string) (int, BytecodeUpvalue, bool) {
	if compiler.parent == nil {
		return 0, BytecodeUpvalue{}, false
	}

	if index, ok := compiler.upvalueIndex(name); ok {
		return index, compiler.chunk.Upvalues[index], true
	}

	if localIndex, ok := compiler.parent.localIndex(name); ok {
		local := compiler.parent.chunk.Locals[localIndex]
		index := compiler.addUpvalue(name, local.IsImmutable, true, localIndex)

		return index, compiler.chunk.Upvalues[index], true
	}

	if parentUpvalueIndex, parentUpvalue, ok := compiler.parent.resolveUpvalueWithValue(name); ok {
		index := compiler.addUpvalue(name, parentUpvalue.IsImmutable, false, parentUpvalueIndex)

		return index, compiler.chunk.Upvalues[index], true
	}

	return 0, BytecodeUpvalue{}, false
}

func (compiler *BytecodeCompiler) addUpvalue(
	name string,
	isImmutable bool,
	isLocal bool,
	index int,
) int {
	if existingIndex, ok := compiler.upvalueIndex(name); ok {
		return existingIndex
	}

	upvalueIndex := compiler.chunk.AddUpvalue(name, isImmutable, isLocal, index)
	compiler.upvalueIndexes[name] = upvalueIndex

	return upvalueIndex
}

func (compiler *BytecodeCompiler) upvalueIndex(name string) (int, bool) {
	index, ok := compiler.upvalueIndexes[name]

	return index, ok
}

func bytecodeOpcodeForPrefixOperator(operator string) (BytecodeOpcode, bool) {
	switch operator {
	case "-":
		return BytecodeOpNegate, true

	case "!":
		return BytecodeOpNot, true

	default:
		return BytecodeOpLoadVoid, false
	}
}

func bytecodeOpcodeForBinaryOperator(operator string) (BytecodeOpcode, bool) {
	switch operator {
	case "+":
		return BytecodeOpAdd, true

	case "-":
		return BytecodeOpSubtract, true

	case "*":
		return BytecodeOpMultiply, true

	case "/":
		return BytecodeOpDivide, true

	case "=":
		return BytecodeOpEqual, true

	case "!":
		return BytecodeOpNotEqual, true

	case "<":
		return BytecodeOpLess, true

	case "<=":
		return BytecodeOpLessEqual, true

	case ">":
		return BytecodeOpGreater, true

	case ">=":
		return BytecodeOpGreaterEqual, true

	default:
		return BytecodeOpLoadVoid, false
	}
}

func bytecodeOpcodeForCompoundAssignmentOperator(tokenType TokenType) (BytecodeOpcode, bool) {
	switch tokenType {
	case TokenPlusAssign:
		return BytecodeOpAdd, true

	case TokenMinusAssign:
		return BytecodeOpSubtract, true

	case TokenStarAssign:
		return BytecodeOpMultiply, true

	case TokenSlashAssign:
		return BytecodeOpDivide, true

	default:
		return BytecodeOpLoadVoid, false
	}
}

func bytecodeParametersFromTokens(tokens []Token) []BytecodeParameter {
	parameters := make([]BytecodeParameter, 0, len(tokens))

	for _, token := range tokens {
		parameters = append(parameters, bytecodeParameterFromToken(token))
	}

	return parameters
}

func bytecodeVariadicParameterFromToken(token *Token) *BytecodeParameter {
	if token == nil {
		return nil
	}

	parameter := bytecodeParameterFromToken(*token)

	return &parameter
}

func bytecodeParameterFromToken(token Token) BytecodeParameter {
	return BytecodeParameter{
		Name:        token.Literal,
		IsImmutable: token.IsImmutable,
	}
}

func (compiler *BytecodeCompiler) sourceSpanFromToken(token Token) BytecodeSourceSpan {
	return BytecodeSourceSpan{
		Filename: compiler.filename,
		Line:     token.StartOfLine,
		Column:   token.StartOfColumn,
	}
}

func (compiler *BytecodeCompiler) sourceSpanForExpression(expression Expression) BytecodeSourceSpan {
	switch expression := expression.(type) {
	case *VoidLiteral:
		return compiler.sourceSpanFromToken(expression.Token)

	case *BoolLiteral:
		return compiler.sourceSpanFromToken(expression.Token)

	case *NumberLiteral:
		return compiler.sourceSpanFromToken(expression.Token)

	case *StringLiteral:
		return compiler.sourceSpanFromToken(expression.Token)

	case *IdentifierExpression:
		return compiler.sourceSpanFromToken(expression.Token)

	case *PrefixExpression:
		return compiler.sourceSpanFromToken(expression.Token)

	case *BinaryExpression:
		return compiler.sourceSpanForExpression(expression.Left)

	case *MapLiteral:
		return compiler.sourceSpanFromToken(expression.Token)

	case *ArrayLiteral:
		return compiler.sourceSpanFromToken(expression.Token)

	case *IndexExpression:
		return compiler.sourceSpanForExpression(expression.Object)

	case *FunctionLiteral:
		return compiler.sourceSpanFromToken(expression.Token)

	case *CallExpression:
		return compiler.sourceSpanForExpression(expression.Callee)

	default:
		return EmptyBytecodeSourceSpan()
	}
}

func (compiler *BytecodeCompiler) compileError(token Token, message string) error {
	return fmt.Errorf(
		"%s:%d:%d: %s",
		compiler.filename,
		token.StartOfLine,
		token.StartOfColumn,
		message,
	)
}
