package main

import (
	"fmt"
	"strings"
)

func FormatBytecode(chunk *BytecodeChunk) string {
	var builder strings.Builder

	builder.WriteString("STULT BYTECODE DISASSEMBLY\n")
	builder.WriteString("==========================\n\n")

	if chunk == nil {
		builder.WriteString("<nil chunk>\n")
		return builder.String()
	}

	formatBytecodeChunk(&builder, chunk, "")

	return builder.String()
}

func formatBytecodeChunk(builder *strings.Builder, chunk *BytecodeChunk, indent string) {
	fmt.Fprintf(builder, "%schunk: %s\n\n", indent, chunk.Name)

	formatBytecodeInstructions(builder, chunk, indent)
	formatBytecodeConstants(builder, chunk, indent)
	formatBytecodeLocals(builder, chunk, indent)
	formatBytecodeUpvalues(builder, chunk, indent)
	formatBytecodeFunctions(builder, chunk, indent)
	formatBytecodeSourceSpans(builder, chunk, indent)
}

func formatBytecodeInstructions(builder *strings.Builder, chunk *BytecodeChunk, indent string) {
	fmt.Fprintf(builder, "%sinstructions:\n", indent)

	if len(chunk.Instructions) == 0 {
		fmt.Fprintf(builder, "%s  <none>\n\n", indent)
		return
	}

	for index, instruction := range chunk.Instructions {
		fmt.Fprintf(
			builder,
			"%s%s\n",
			indent,
			formatBytecodeInstruction(index, instruction, chunk),
		)
	}

	builder.WriteByte('\n')
}

func formatBytecodeConstants(builder *strings.Builder, chunk *BytecodeChunk, indent string) {
	fmt.Fprintf(builder, "%sconstants:\n", indent)

	if len(chunk.Constants) == 0 {
		fmt.Fprintf(builder, "%s  <none>\n\n", indent)
		return
	}

	for index, constant := range chunk.Constants {
		fmt.Fprintf(
			builder,
			"%s[%d] %s %s\n",
			indent,
			index,
			bytecodeConstantKindName(constant),
			constant.DebugString(),
		)
	}

	builder.WriteByte('\n')
}

func formatBytecodeLocals(builder *strings.Builder, chunk *BytecodeChunk, indent string) {
	fmt.Fprintf(builder, "%slocals:\n", indent)

	if len(chunk.Locals) == 0 {
		fmt.Fprintf(builder, "%s  <none>\n\n", indent)
		return
	}

	for index, local := range chunk.Locals {
		fmt.Fprintf(
			builder,
			"%s[%d] %s %s scope=%d\n",
			indent,
			index,
			local.Name,
			bytecodeMutabilityName(local.IsImmutable),
			local.ScopeDepth,
		)
	}

	builder.WriteByte('\n')
}

func formatBytecodeUpvalues(builder *strings.Builder, chunk *BytecodeChunk, indent string) {
	fmt.Fprintf(builder, "%supvalues:\n", indent)

	if len(chunk.Upvalues) == 0 {
		fmt.Fprintf(builder, "%s  <none>\n\n", indent)
		return
	}

	for index, upvalue := range chunk.Upvalues {
		source := "upvalue"
		if upvalue.IsLocal {
			source = "local"
		}

		fmt.Fprintf(
			builder,
			"%s[%d] %s %s from=%s[%d]\n",
			indent,
			index,
			upvalue.Name,
			bytecodeMutabilityName(upvalue.IsImmutable),
			source,
			upvalue.Index,
		)
	}

	builder.WriteByte('\n')
}

func formatBytecodeFunctions(builder *strings.Builder, chunk *BytecodeChunk, indent string) {
	fmt.Fprintf(builder, "%sfunctions:\n", indent)

	if len(chunk.Functions) == 0 {
		fmt.Fprintf(builder, "%s  <none>\n\n", indent)
		return
	}

	for index, function := range chunk.Functions {
		formatBytecodeFunction(builder, index, function, indent+"  ")
	}

	builder.WriteByte('\n')
}

func formatBytecodeFunction(
	builder *strings.Builder,
	index int,
	function BytecodeFunction,
	indent string,
) {
	fmt.Fprintf(builder, "%s[%d] %s\n", indent, index, function.Name)

	if len(function.Parameters) == 0 {
		fmt.Fprintf(builder, "%sparameters: <none>\n", indent)
	} else {
		fmt.Fprintf(builder, "%sparameters:\n", indent)

		for parameterIndex, parameter := range function.Parameters {
			fmt.Fprintf(
				builder,
				"%s  [%d] %s %s\n",
				indent,
				parameterIndex,
				parameter.Name,
				bytecodeMutabilityName(parameter.IsImmutable),
			)
		}
	}

	if function.VariadicParameter == nil {
		fmt.Fprintf(builder, "%svariadic: <none>\n", indent)
	} else {
		fmt.Fprintf(
			builder,
			"%svariadic: %s %s\n",
			indent,
			function.VariadicParameter.Name,
			bytecodeMutabilityName(function.VariadicParameter.IsImmutable),
		)
	}

	if len(function.Upvalues) == 0 {
		fmt.Fprintf(builder, "%supvalues: <none>\n\n", indent)
	} else {
		fmt.Fprintf(builder, "%supvalues:\n", indent)

		for upvalueIndex, upvalue := range function.Upvalues {
			source := "upvalue"
			if upvalue.IsLocal {
				source = "local"
			}

			fmt.Fprintf(
				builder,
				"%s  [%d] %s %s from=%s[%d]\n",
				indent,
				upvalueIndex,
				upvalue.Name,
				bytecodeMutabilityName(upvalue.IsImmutable),
				source,
				upvalue.Index,
			)
		}

		builder.WriteByte('\n')
	}

	if function.Chunk == nil {
		fmt.Fprintf(builder, "%s  <nil chunk>\n", indent)
		return
	}

	formatBytecodeChunk(builder, function.Chunk, indent+"  ")
}

func formatBytecodeSourceSpans(builder *strings.Builder, chunk *BytecodeChunk, indent string) {
	fmt.Fprintf(builder, "%ssource spans:\n", indent)

	if len(chunk.SourceSpans) == 0 {
		fmt.Fprintf(builder, "%s  <none>\n", indent)
		return
	}

	for index, sourceSpan := range chunk.SourceSpans {
		fmt.Fprintf(
			builder,
			"%s%04d %s\n",
			indent,
			index,
			formatBytecodeSourceSpan(sourceSpan),
		)
	}
}

func formatBytecodeInstruction(index int, instruction BytecodeInstruction, chunk *BytecodeChunk) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "%04d %-24s", index, instruction.Opcode.String())

	switch instruction.Opcode {
	case BytecodeOpLoadConst,
		BytecodeOpLoadGlobal,
		BytecodeOpStoreGlobalMutable,
		BytecodeOpStoreGlobalImmutable:
		fmt.Fprintf(&builder, " %d", instruction.Operand)

		if instruction.Operand >= 0 && instruction.Operand < len(chunk.Constants) {
			fmt.Fprintf(&builder, "    ; %s", chunk.Constants[instruction.Operand].DebugString())
		}

	case BytecodeOpLoadLocal,
		BytecodeOpStoreLocalMutable,
		BytecodeOpStoreLocalImmutable,
		BytecodeOpIteratorValue,
		BytecodeOpIteratorKey,
		BytecodeOpIteratorCollection,
		BytecodeOpIteratorPosition:
		fmt.Fprintf(&builder, " %d", instruction.Operand)

		if instruction.Operand >= 0 && instruction.Operand < len(chunk.Locals) {
			fmt.Fprintf(&builder, "    ; %s", chunk.Locals[instruction.Operand].Name)
		}

	case BytecodeOpLoadUpvalue,
		BytecodeOpStoreUpvalueMutable,
		BytecodeOpStoreUpvalueImmutable:
		fmt.Fprintf(&builder, " %d", instruction.Operand)

		if instruction.Operand >= 0 && instruction.Operand < len(chunk.Upvalues) {
			fmt.Fprintf(&builder, "    ; %s", chunk.Upvalues[instruction.Operand].Name)
		}

	case BytecodeOpBuildArray,
		BytecodeOpBuildMap,
		BytecodeOpCall:
		fmt.Fprintf(&builder, " %d", instruction.Operand)

	case BytecodeOpBuildRange:
		fmt.Fprintf(&builder, " %d", instruction.Operand)

		if instruction.Operand == 1 {
			builder.WriteString("    ; inclusive")
		} else {
			builder.WriteString("    ; exclusive")
		}

	case BytecodeOpMakeFunction:
		fmt.Fprintf(&builder, " %d", instruction.Operand)

		if instruction.Operand >= 0 && instruction.Operand < len(chunk.Functions) {
			fmt.Fprintf(&builder, "    ; %s", chunk.Functions[instruction.Operand].Name)
		}

	case BytecodeOpResetLocals:
		fmt.Fprintf(
			&builder,
			" %d    ; scope >= %d",
			instruction.Operand,
			instruction.Operand,
		)

	case BytecodeOpJump,
		BytecodeOpJumpIfFalse,
		BytecodeOpJumpIfTrue,
		BytecodeOpJumpIfCollection,
		BytecodeOpIteratorNext:
		fmt.Fprintf(&builder, " %04d", instruction.Operand)

	case BytecodeOpLoadVoid,
		BytecodeOpLoadTrue,
		BytecodeOpLoadFalse,
		BytecodeOpStoreIndex,
		BytecodeOpIndex,
		BytecodeOpIteratorInit,
		BytecodeOpIteratorEnd,
		BytecodeOpDuplicateTopTwo,
		BytecodeOpNegate,
		BytecodeOpNot,
		BytecodeOpAdd,
		BytecodeOpSubtract,
		BytecodeOpMultiply,
		BytecodeOpDivide,
		BytecodeOpEqual,
		BytecodeOpNotEqual,
		BytecodeOpLess,
		BytecodeOpLessEqual,
		BytecodeOpGreater,
		BytecodeOpGreaterEqual,
		BytecodeOpPop,
		BytecodeOpReturn:
		// No operand.

	default:
		if instruction.Operand != 0 {
			fmt.Fprintf(&builder, " %d", instruction.Operand)
		}
	}

	return builder.String()
}

func formatBytecodeSourceSpan(sourceSpan BytecodeSourceSpan) string {
	if sourceSpan.Filename == "" || sourceSpan.Line == 0 || sourceSpan.Column == 0 {
		return "<unknown>"
	}

	return fmt.Sprintf(
		"%s:%d:%d",
		sourceSpan.Filename,
		sourceSpan.Line,
		sourceSpan.Column,
	)
}

func bytecodeMutabilityName(isImmutable bool) string {
	if isImmutable {
		return "immutable"
	}

	return "mutable"
}

func bytecodeConstantKindName(value Value) string {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueVoid:
		return "void"

	case ValueNumber:
		return "number"

	case ValueBool:
		return "bool"

	case ValueString:
		return "string"

	case ValueMap:
		return "map"

	case ValueArray:
		return "array"

	case ValueFunction:
		return "function"

	case ValueBuiltinFunction:
		return "builtin_function"

	default:
		return "unknown"
	}
}
