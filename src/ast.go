package main

type Program struct {
	Statements []Statement
}

type Statement interface {
	statementNode()
}

type Expression interface {
	expressionNode()
}

type AssignmentStatement struct {
	Name        Token
	Value       Expression
	IsImmutable bool
}

func (s *AssignmentStatement) statementNode() {}

type ExpressionStatement struct {
	Expression Expression
}

func (s *ExpressionStatement) statementNode() {}

type NumberLiteral struct {
	Token Token
	Value string
}

func (e *NumberLiteral) expressionNode() {}

type IdentifierExpression struct {
	Token       Token
	Name        string
	IsImmutable bool
}

func (e *IdentifierExpression) expressionNode() {}

type PrefixExpression struct {
	Token    Token
	Operator string
	Right    Expression
}

func (e *PrefixExpression) expressionNode() {}

type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryExpression) expressionNode() {}

type GroupedExpression struct {
	Expression Expression
}

func (e *GroupedExpression) expressionNode() {}

type StringLiteral struct {
	Token Token
	Value string
}

func (*StringLiteral) expressionNode() {}
