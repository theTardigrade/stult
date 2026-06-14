package main

import "fmt"

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
