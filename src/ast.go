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
	Name                Token
	Value               Expression
	IsImmutable         bool
	IsOuter             bool
	ContractDeclaration *BindingContractDeclaration
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

type BreakStatement struct {
	Token Token
}

func (s *BreakStatement) statementNode() {}

type ReturnStatement struct {
	Token Token
	Value Expression
}

func (s *ReturnStatement) statementNode() {}

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

type TryCatchStatement struct {
	Token          Token
	TryBody        []Statement
	CatchParameter *Token
	CatchBody      []Statement
}

func (s *TryCatchStatement) statementNode() {}

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

type ConditionalExpression struct {
	Token     Token
	Condition Expression
	WhenTrue  Expression
	WhenFalse Expression
}

func (e *ConditionalExpression) expressionNode() {}

type MatchExpression struct {
	Token   Token
	Target  Expression
	Arms    []MatchArm
	Default Expression
}

func (e *MatchExpression) expressionNode() {}

type FallibleExpression struct {
	Token    Token
	Attempt  Expression
	Fallback Expression
}

func (e *FallibleExpression) expressionNode() {}

type MatchArm struct {
	Pattern MatchPattern
	Value   Expression
}

type MatchPatternKind int

const (
	MatchPatternString MatchPatternKind = iota
	MatchPatternNumber
	MatchPatternBool
)

type MatchPattern struct {
	Token Token
	Kind  MatchPatternKind
}

type StringLiteral struct {
	Token  Token
	Value  string
	Frozen bool
}

func (*StringLiteral) expressionNode() {}

type MapLiteral struct {
	Token   Token
	Entries []MapEntry
	Frozen  bool
}

func (*MapLiteral) expressionNode() {}

type MapEntry struct {
	Key      Token
	Value    Expression
	IsDotKey bool
}

type ArrayLiteral struct {
	Token    Token
	Elements []ArrayElement
	Frozen   bool
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

type RangeIndexExpression struct {
	Object      Expression
	Start       Expression
	End         Expression
	Step        Expression
	IsInclusive bool
}

func (*RangeIndexExpression) expressionNode() {}

type LeadingDotReceiverExpression struct {
	Token Token
}

func (*LeadingDotReceiverExpression) expressionNode() {}

type IndexAssignmentStatement struct {
	Target *IndexExpression
	Value  Expression
}

func (*IndexAssignmentStatement) statementNode() {}

type FunctionParameter struct {
	Token      Token
	IsOptional bool
}

type FunctionLiteral struct {
	Token             Token
	Parameters        []FunctionParameter
	VariadicParameter *Token
	Body              []Statement
	Returns           []Expression
}

func (*FunctionLiteral) expressionNode() {}

type CallArgument struct {
	Expression  Expression
	IsSpread    bool
	SpreadToken Token
}

type CallExpression struct {
	Callee    Expression
	Arguments []CallArgument
}

func (*CallExpression) expressionNode() {}
