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
	IsOptional  bool
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
