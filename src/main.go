package main

import (
	"fmt"
	"strings"
)

func main() {
	source := `|| Parser test program
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

	PrintProgram(program)
}

func PrintProgram(program *Program) {
	fmt.Println("Program")
	for _, stmt := range program.Statements {
		printStatement(stmt, 1)
	}
}

func printStatement(stmt Statement, depth int) {
	indent := strings.Repeat("  ", depth)

	switch s := stmt.(type) {
	case *AssignmentStatement:
		mutability := "mutable"
		if s.IsImmutable {
			mutability = "immutable"
		}

		fmt.Printf("%sAssignment name=%q %s\n", indent, s.Name.Literal, mutability)
		printExpression(s.Value, depth+1)

	case *ExpressionStatement:
		fmt.Printf("%sExpressionStatement\n", indent)
		printExpression(s.Expression, depth+1)

	default:
		fmt.Printf("%sUnknown statement %T\n", indent, stmt)
	}
}

func printExpression(expr Expression, depth int) {
	indent := strings.Repeat("  ", depth)

	switch e := expr.(type) {
	case *NumberLiteral:
		fmt.Printf("%sNumber %q\n", indent, e.Value)

	case *IdentifierExpression:
		mutability := "mutable"
		if e.IsImmutable {
			mutability = "immutable"
		}

		fmt.Printf("%sIdentifier %q %s\n", indent, e.Name, mutability)

	case *PrefixExpression:
		fmt.Printf("%sPrefix %q\n", indent, e.Operator)
		printExpression(e.Right, depth+1)

	case *BinaryExpression:
		fmt.Printf("%sBinary %q\n", indent, e.Operator)
		printExpression(e.Left, depth+1)
		printExpression(e.Right, depth+1)

	default:
		fmt.Printf("%sUnknown expression %T\n", indent, expr)
	}
}
