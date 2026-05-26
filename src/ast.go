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
	IsOuter     bool
}

func (s *AssignmentStatement) statementNode() {}

type CompoundAssignmentStatement struct {
	Target   Expression
	Operator Token
	Value    Expression
}

func (s *CompoundAssignmentStatement) statementNode() {}

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
	Condition       Expression
	RangeParameters []Token
	Body            []Statement
	AfterLoopBody   []Statement
}

func (s *LoopStatement) statementNode() {}

type VoidLiteral struct {
	Token Token
}

func (e *VoidLiteral) expressionNode() {}

type BoolLiteral struct {
	Token Token
	Value bool
}

func (e *BoolLiteral) expressionNode() {}

type NumberLiteral struct {
	Token Token
	Value string
}

func (e *NumberLiteral) expressionNode() {}

type IdentifierExpression struct {
	Token       Token
	Name        string
	IsImmutable bool
	IsOuter     bool
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
	Key   Token
	Value Expression
}

type ArrayLiteral struct {
	Token    Token
	Elements []ArrayElement
}

func (*ArrayLiteral) expressionNode() {}

type ArrayElement interface {
	arrayElementNode()
}

type ExpressionArrayElement struct {
	Expression Expression
}

func (*ExpressionArrayElement) arrayElementNode() {}

type RangeArrayElement struct {
	Start       Expression
	End         Expression
	Step        Expression
	IsInclusive bool
}

func (*RangeArrayElement) arrayElementNode() {}

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
