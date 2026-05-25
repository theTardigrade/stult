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

type ConditionalStatement struct {
	Branches []ConditionalBranch
	ElseBody []Statement
}

func (s *ConditionalStatement) statementNode() {}

type ConditionalBranch struct {
	Condition Expression
	Body      []Statement
}

type LoopStatement struct {
	Condition     Expression
	Body          []Statement
	AfterLoopBody []Statement
}

func (s *LoopStatement) statementNode() {}

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

type StringLiteral struct {
	Token Token
	Value string
}

func (*StringLiteral) expressionNode() {}

type MapLiteral struct {
	Token   Token
	Entries []MapEntry
}

func (*MapLiteral) expressionNode() {}

type MapEntry struct {
	Key   Token // Type == TokenString
	Value Expression
}

type IndexExpression struct {
	Object Expression
	Index  Expression
}

func (*IndexExpression) expressionNode() {}

type IndexAssignmentStatement struct {
	Target *IndexExpression
	Value  Expression
}

func (*IndexAssignmentStatement) statementNode() {}

type FunctionLiteral struct {
	Token      Token
	Parameters []Token
	Body       []Statement
	Returns    []Expression
}

func (*FunctionLiteral) expressionNode() {}

type CallExpression struct {
	Callee    Expression
	Arguments []Expression
}

func (*CallExpression) expressionNode() {}
