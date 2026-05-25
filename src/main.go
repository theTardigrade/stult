package main

import "fmt"

func main() {
	source := `|| Interpreter test program
PI = 3.14159
radius = 10
area = PI * radius * radius
x = 5, y = 9
z = -(x + y) * 2
is_big = area > 100
`

	lexer := NewLexer(source)
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range parser.Errors() {
			fmt.Println("  -", err)
		}
		return
	}

	interpreter := NewInterpreter()

	if err := interpreter.EvalProgram(program); err != nil {
		fmt.Println("Runtime error:")
		fmt.Println("  -", err)
		return
	}

	fmt.Println("Final environment:")
	interpreter.Env.Dump()
}
