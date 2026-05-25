package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: interpreter <file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	sourceBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read %q: %v\n", filename, err)
		os.Exit(1)
	}

	lexer := NewLexer(string(sourceBytes))
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Parser errors:")
		for _, err := range parser.Errors() {
			fmt.Fprintln(os.Stderr, "  -", err)
		}
		os.Exit(1)
	}

	interpreter := NewInterpreter()

	if err := interpreter.EvalProgram(program); err != nil {
		fmt.Fprintln(os.Stderr, "Runtime error:")
		fmt.Fprintln(os.Stderr, "  -", err)
		os.Exit(1)
	}

	interpreter.Env.Dump()
}
