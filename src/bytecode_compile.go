package main

import "fmt"

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

func CompileBytecode(program *Program, name string) (*BytecodeChunk, error) {
	compiler := NewBytecodeCompiler(name, name, false, nil)

	if err := compiler.compileProgram(program); err != nil {
		return nil, err
	}

	return compiler.chunk, nil
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

func bytecodeParametersFromFunctionParameters(
	functionParameters []FunctionParameter,
) []BytecodeParameter {
	parameters := make([]BytecodeParameter, 0, len(functionParameters))

	for _, functionParameter := range functionParameters {
		parameters = append(parameters, bytecodeParameterFromFunctionParameter(functionParameter))
	}

	return parameters
}

func bytecodeParameterFromFunctionParameter(
	functionParameter FunctionParameter,
) BytecodeParameter {
	return BytecodeParameter{
		Name:        functionParameter.Token.Literal,
		IsImmutable: functionParameter.Token.IsImmutable,
		IsOptional:  functionParameter.IsOptional,
	}
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
		IsOptional:  false,
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

	case *ConditionalExpression:
		return compiler.sourceSpanFromToken(expression.Token)

	case *MatchExpression:
		return compiler.sourceSpanFromToken(expression.Token)

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
