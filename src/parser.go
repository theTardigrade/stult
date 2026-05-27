package main

import "fmt"

type Parser struct {
	lexer *Lexer

	previous Token
	current  Token
	peek     Token

	errors []string
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{lexer: lexer}

	p.advance()
	p.advance()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{}

	for p.current.Type != TokenEOF {
		p.skipSeparators()

		if p.current.Type == TokenEOF {
			break
		}

		stmt := p.parseStatement()
		canFollowTightly := statementAllowsTightFollower(stmt) && p.previous.Type == TokenRBrace

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		if p.current.Type == TokenComma || p.current.Type == TokenNewline {
			p.skipSeparators()
		} else if p.current.Type == TokenEOF {
			break
		} else if canFollowTightly {
			continue
		} else {
			p.errorAtCurrent("expected comma, newline, or end of file after statement")
			p.synchronize()
		}
	}

	return program
}

func (p *Parser) parseStatement() Statement {
	if p.current.Type == TokenCaret {
		return p.parseCaretStatement()
	}

	if p.isLoopStart() {
		return p.parseLoopStatement()
	}

	if p.current.Type == TokenLParen {
		expr, closeParen, ok := p.parseParenthesizedExpression("conditional expression cannot be empty")
		if !ok {
			return nil
		}

		if p.current.Type == TokenLBrace {
			stmt, ok := p.finishConditionalStatement(expr, closeParen)
			if !ok {
				return nil
			}

			return stmt
		}

		target := p.parseExpressionTail(expr, precLowest)
		if target == nil {
			return nil
		}

		return p.finishExpressionOrAssignmentStatement(target)
	}

	target := p.parseExpression(precLowest)
	if target == nil {
		return nil
	}

	return p.finishExpressionOrAssignmentStatement(target)
}

func (p *Parser) finishExpressionOrAssignmentStatement(target Expression) Statement {
	if !isAssignmentOperator(p.current.Type) {
		return &ExpressionStatement{Expression: target}
	}

	assignToken := p.current
	p.advance() // consume assignment operator

	value := p.parseExpression(precLowest)
	if value == nil {
		p.errorAtToken(assignToken, "expected expression after assignment")
		return nil
	}

	if assignToken.Type == TokenPlusAssign || assignToken.Type == TokenMinusAssign {
		if !isAssignableExpression(target) {
			p.errorAtToken(assignToken, "invalid assignment target")
			return nil
		}

		return &CompoundAssignmentStatement{
			Target:   target,
			Operator: assignToken,
			Value:    value,
		}
	}

	switch t := target.(type) {
	case *IdentifierExpression:
		return &AssignmentStatement{
			Name:        t.Token,
			Value:       value,
			IsImmutable: t.IsImmutable,
			IsOuter:     t.IsOuter,
		}

	case *IndexExpression:
		return &IndexAssignmentStatement{
			Target: t,
			Value:  value,
		}

	default:
		p.errorAtToken(assignToken, "invalid assignment target")
		return nil
	}
}

func isAssignmentOperator(tokenType TokenType) bool {
	return tokenType == TokenColon ||
		tokenType == TokenPlusAssign ||
		tokenType == TokenMinusAssign
}

func isAssignableExpression(expr Expression) bool {
	switch expr.(type) {
	case *IdentifierExpression, *IndexExpression:
		return true
	default:
		return false
	}
}

func (p *Parser) parseCaretStatement() Statement {
	caret := p.current
	p.advance() // consume "^"

	if p.current.Type != TokenLParen || !tokensTouch(caret, p.current) {
		return &BreakStatement{Token: caret}
	}

	value, _, ok := p.parseParenthesizedExpression("return expression cannot be empty")
	if !ok {
		return nil
	}

	return &ReturnStatement{
		Token: caret,
		Value: value,
	}
}

func statementAllowsTightFollower(stmt Statement) bool {
	switch stmt.(type) {
	case *ConditionalStatement, *LoopStatement:
		return true
	default:
		return false
	}
}

func (p *Parser) isLoopStart() bool {
	return p.current.Type == TokenLParen &&
		p.peek.Type == TokenLParen &&
		tokensTouch(p.current, p.peek)
}

func (p *Parser) parseLoopStatement() Statement {
	outerOpen := p.current

	p.advance() // consume first "(" and move to second "("
	p.advance() // consume second "(" and move to loop expression

	if p.current.Type == TokenRParen {
		p.errorAtToken(outerOpen, "loop expression cannot be empty")
		return nil
	}

	condition := p.parseExpression(precLowest)
	if condition == nil {
		p.errorAtToken(outerOpen, "expected loop expression")
		return nil
	}

	innerClose := p.current
	if !p.expectCurrent(TokenRParen, "expected ')' after loop expression") {
		return nil
	}

	p.advance() // consume inner ")"

	outerClose := p.current
	if !p.expectCurrent(TokenRParen, "expected second ')' after loop expression") {
		return nil
	}

	if !tokensTouch(innerClose, outerClose) {
		p.errorAtToken(outerClose, "expected loop expression to close with touching '))'")
		return nil
	}

	p.advance() // consume outer ")"

	if !p.expectCurrent(TokenLBrace, "expected '{' after loop expression") {
		return nil
	}

	if !tokensOnSameLine(outerClose, p.current) {
		p.errorAtCurrent("expected loop block to start on the same line as the loop expression")
		return nil
	}

	rangeParameters, body, closeBrace, ok := p.parseLoopBodyBlock("loop block")
	if !ok {
		return nil
	}

	var afterLoopBody []Statement

	if p.current.Type == TokenComma && p.peek.Type == TokenLBrace {
		comma := p.current

		if !tokensTouch(closeBrace, comma) || !tokensTouch(comma, p.peek) {
			p.errorAtToken(comma, "expected after-loop separator to be written without whitespace as '},{'")
			return nil
		}

		p.advance() // consume "," and move to "{"

		afterLoopBody, _, ok = p.parseStatementBlock("after-loop block")
		if !ok {
			return nil
		}
	}

	return &LoopStatement{
		Condition:       condition,
		RangeParameters: rangeParameters,
		Body:            body,
		AfterLoopBody:   afterLoopBody,
	}
}

func (p *Parser) parseLoopBodyBlock(name string) ([]Token, []Statement, Token, bool) {
	openBrace := p.current

	if !p.expectCurrent(TokenLBrace, "expected '{' to start "+name) {
		return nil, nil, Token{}, false
	}

	p.advance() // consume "{"
	p.skipSeparators()

	rangeParameters := []Token{}

	if p.current.Type == TokenLParen && !p.isLoopStart() {
		parameters, ok := p.parseBindingParameters("range parameter")
		if !ok {
			return nil, nil, Token{}, false
		}

		rangeParameters = parameters
	}

	body := []Statement{}

	for {
		p.skipSeparators()

		switch p.current.Type {
		case TokenEOF:
			p.errorAtToken(openBrace, "unterminated "+name)
			return nil, nil, Token{}, false

		case TokenRBrace:
			closeBrace := p.current
			p.advance()
			return rangeParameters, body, closeBrace, true
		}

		stmt := p.parseStatement()
		canFollowTightly := statementAllowsTightFollower(stmt) && p.previous.Type == TokenRBrace

		if stmt == nil {
			return nil, nil, Token{}, false
		}

		body = append(body, stmt)

		if p.current.Type == TokenComma || p.current.Type == TokenNewline {
			p.skipSeparators()
			continue
		}

		if p.current.Type == TokenRBrace || p.current.Type == TokenEOF {
			continue
		}

		if canFollowTightly {
			continue
		}

		p.errorAtCurrent("expected comma, newline, or '}' after loop statement")
		return nil, nil, Token{}, false
	}
}

func (p *Parser) finishConditionalStatement(firstCondition Expression, firstCloseParen Token) (Statement, bool) {
	if !p.expectCurrent(TokenLBrace, "expected '{' after conditional expression") {
		return nil, false
	}

	if !tokensOnSameLine(firstCloseParen, p.current) {
		p.errorAtCurrent("expected conditional block to start on the same line as the condition")
		return nil, false
	}

	body, closeBrace, ok := p.parseStatementBlock("conditional block")
	if !ok {
		return nil, false
	}

	branches := []ConditionalBranch{
		{
			Condition: firstCondition,
			Body:      body,
		},
	}

	var elseBody []Statement

	for p.current.Type == TokenComma {
		comma := p.current

		switch p.peek.Type {
		case TokenLParen:
			if elseBody != nil {
				p.errorAtToken(comma, "else-if cannot appear after else")
				return nil, false
			}

			if !tokensTouch(closeBrace, comma) || !tokensTouch(comma, p.peek) {
				p.errorAtToken(comma, "expected else-if separator to be written without whitespace as '},('")
				return nil, false
			}

			p.advance() // consume "," and move to "("

			condition, conditionCloseParen, ok := p.parseParenthesizedExpression("else-if expression cannot be empty")
			if !ok {
				return nil, false
			}

			if !p.expectCurrent(TokenLBrace, "expected '{' after else-if expression") {
				return nil, false
			}

			if !tokensOnSameLine(conditionCloseParen, p.current) {
				p.errorAtCurrent("expected else-if block to start on the same line as the condition")
				return nil, false
			}

			body, nextCloseBrace, ok := p.parseStatementBlock("else-if block")
			if !ok {
				return nil, false
			}

			branches = append(branches, ConditionalBranch{
				Condition: condition,
				Body:      body,
			})

			closeBrace = nextCloseBrace

		case TokenLBrace:
			if !tokensTouch(closeBrace, comma) || !tokensTouch(comma, p.peek) {
				p.errorAtToken(comma, "expected else separator to be written without whitespace as '},{'")
				return nil, false
			}

			p.advance() // consume "," and move to "{"

			body, _, ok := p.parseStatementBlock("else block")
			if !ok {
				return nil, false
			}

			elseBody = body

			return &ConditionalStatement{
				Branches: branches,
				ElseBody: elseBody,
			}, true

		default:
			return &ConditionalStatement{
				Branches: branches,
				ElseBody: elseBody,
			}, true
		}
	}

	return &ConditionalStatement{
		Branches: branches,
		ElseBody: elseBody,
	}, true
}

func (p *Parser) parseParenthesizedExpression(emptyMessage string) (Expression, Token, bool) {
	openParen := p.current

	if !p.expectCurrent(TokenLParen, "expected '(' before expression") {
		return nil, Token{}, false
	}

	p.advance() // consume "("
	p.skipNewlines()

	if p.current.Type == TokenRParen {
		p.errorAtToken(openParen, emptyMessage)
		return nil, Token{}, false
	}

	expr := p.parseExpression(precLowest)
	if expr == nil {
		p.errorAtToken(openParen, "expected expression")
		return nil, Token{}, false
	}

	p.skipNewlines()

	closeParen := p.current
	if !p.expectCurrent(TokenRParen, "expected ')' after expression") {
		return nil, Token{}, false
	}

	p.advance() // consume ")"

	return expr, closeParen, true
}

func (p *Parser) parseStatementBlock(name string) ([]Statement, Token, bool) {
	openBrace := p.current

	if !p.expectCurrent(TokenLBrace, "expected '{' to start "+name) {
		return nil, Token{}, false
	}

	p.advance() // consume "{"

	body := []Statement{}

	for {
		p.skipSeparators()

		switch p.current.Type {
		case TokenEOF:
			p.errorAtToken(openBrace, "unterminated "+name)
			return nil, Token{}, false

		case TokenRBrace:
			closeBrace := p.current
			p.advance()
			return body, closeBrace, true
		}

		stmt := p.parseStatement()
		canFollowTightly := statementAllowsTightFollower(stmt) && p.previous.Type == TokenRBrace

		if stmt == nil {
			return nil, Token{}, false
		}

		body = append(body, stmt)

		if p.current.Type == TokenComma || p.current.Type == TokenNewline {
			p.skipSeparators()
			continue
		}

		if p.current.Type == TokenRBrace || p.current.Type == TokenEOF {
			continue
		}

		if canFollowTightly {
			continue
		}

		p.errorAtCurrent("expected comma, newline, or '}' after statement")
		return nil, Token{}, false
	}
}

const (
	precLowest = iota
	precLogicalOr
	precLogicalAnd
	precEquality
	precComparison
	precTerm
	precFactor
	precPrefix
)

func precedence(tok TokenType) int {
	switch tok {
	case TokenOr:
		return precLogicalOr
	case TokenAnd:
		return precLogicalAnd
	case TokenEqual, TokenNotEqual:
		return precEquality
	case TokenLess, TokenLessEqual, TokenGreater, TokenGreaterEqual:
		return precComparison
	case TokenPlus, TokenMinus:
		return precTerm
	case TokenStar, TokenSlash:
		return precFactor
	default:
		return precLowest
	}
}

func (p *Parser) parseExpression(parentPrec int) Expression {
	return p.parseExpressionWithOptions(parentPrec, false)
}

func (p *Parser) parseRangeEndExpression() Expression {
	return p.parseExpressionWithOptions(precLowest, true)
}

func (p *Parser) parseExpressionWithOptions(parentPrec int, stopBeforeTouchingIndex bool) Expression {
	var left Expression

	switch p.current.Type {
	case TokenBool:
		left = &BoolLiteral{
			Token: p.current,
			Value: p.current.Literal == "\\/",
		}
		p.advance()

	case TokenNumber:
		left = &NumberLiteral{
			Token: p.current,
			Value: p.current.Literal,
		}
		p.advance()

	case TokenString:
		left = &StringLiteral{
			Token: p.current,
			Value: p.current.Literal,
		}
		p.advance()

	case TokenIdentifier:
		if p.current.Literal == "_" {
			left = &VoidLiteral{
				Token: p.current,
			}
			p.advance()
		} else {
			left = &IdentifierExpression{
				Token:       p.current,
				Name:        p.current.Literal,
				IsImmutable: p.current.IsImmutable,
			}
			p.advance()
		}

	case TokenAt:
		outer, ok := p.parseOuterIdentifierExpression()
		if !ok {
			return nil
		}

		left = outer

	case TokenMinus, TokenNotEqual:
		operator := p.current
		p.advance()

		right := p.parseExpressionWithOptions(precPrefix, stopBeforeTouchingIndex)
		if right == nil {
			p.errorAtToken(operator, "expected expression after unary "+strconvQuote(operator.Literal))
			return nil
		}

		left = &PrefixExpression{
			Token:    operator,
			Operator: operator.Literal,
			Right:    right,
		}

	case TokenLParen:
		inner, _, ok := p.parseParenthesizedExpression("grouped expression cannot be empty")
		if !ok {
			return nil
		}

		left = inner

	case TokenLBrace:
		left = p.parseBraceLiteral()
		if left == nil {
			return nil
		}

	default:
		p.errorAtCurrent("expected expression")
		return nil
	}

	return p.parseExpressionTailWithOptions(left, parentPrec, stopBeforeTouchingIndex)
}

func (p *Parser) parseOuterIdentifierExpression() (Expression, bool) {
	at := p.current
	p.advance() // consume "@"

	if p.current.Type != TokenIdentifier {
		p.errorAtToken(at, "expected identifier after '@'")
		return nil, false
	}

	if p.current.Literal == "_" {
		p.errorAtToken(p.current, "'_' is a void literal, not an outer binding")
		return nil, false
	}

	if !tokensTouch(at, p.current) {
		p.errorAtToken(p.current, "expected '@' to touch outer identifier")
		return nil, false
	}

	identifier := p.current
	p.advance()

	return &IdentifierExpression{
		Token:       identifier,
		Name:        identifier.Literal,
		IsImmutable: identifier.IsImmutable,
		IsOuter:     true,
	}, true
}

func (p *Parser) parseExpressionTail(left Expression, parentPrec int) Expression {
	return p.parseExpressionTailWithOptions(left, parentPrec, false)
}

func (p *Parser) parseExpressionTailWithOptions(left Expression, parentPrec int, stopBeforeTouchingIndex bool) Expression {
	for {
		if p.current.Type == TokenLBracket {
			if stopBeforeTouchingIndex && tokensTouch(p.previous, p.current) {
				break
			}

			if !tokensTouch(p.previous, p.current) {
				p.errorAtCurrent("expected '[' to touch indexed expression")
				return nil
			}

			index, ok := p.parseIndexExpression(left)
			if !ok {
				return nil
			}

			left = index
			continue
		}

		if p.current.Type == TokenLParen && tokensTouch(p.previous, p.current) {
			call, ok := p.parseCallExpression(left)
			if !ok {
				return nil
			}

			left = call
			continue
		}

		currentPrec := precedence(p.current.Type)
		if currentPrec == precLowest || currentPrec <= parentPrec {
			break
		}

		operator := p.current
		p.advance()

		right := p.parseExpressionWithOptions(currentPrec, stopBeforeTouchingIndex)
		if right == nil {
			p.errorAtToken(operator, "expected expression after operator")
			return nil
		}

		left = &BinaryExpression{
			Left:     left,
			Operator: operator.Literal,
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parseBraceLiteral() Expression {
	openBrace := p.current
	p.advance() // consume "{"

	p.skipNewlines()

	if p.current.Type == TokenRBrace {
		p.advance()
		return &ArrayLiteral{
			Token:    openBrace,
			Elements: []ArrayElement{},
		}
	}

	if p.current.Type == TokenColon {
		p.advance()
		p.skipNewlines()

		if !p.expectCurrent(TokenRBrace, "expected '}' after ':' in empty map literal") {
			return nil
		}

		p.advance()

		return &MapLiteral{
			Token:   openBrace,
			Entries: []MapEntry{},
		}
	}

	if p.current.Type == TokenLParen {
		return p.parseFunctionLiteral(openBrace)
	}

	if p.current.Type == TokenString && p.peek.Type == TokenColon {
		return p.parseMapLiteral(openBrace)
	}

	return p.parseArrayLiteral(openBrace)
}

func (p *Parser) parseFunctionLiteral(openBrace Token) Expression {
	parameters, ok := p.parseFunctionParameters()
	if !ok {
		return nil
	}

	body := []Statement{}

	for {
		p.skipSeparators()

		switch p.current.Type {
		case TokenEOF:
			p.errorAtToken(openBrace, "unterminated function literal")
			return nil

		case TokenRBrace:
			p.errorAtCurrent("expected return list before '}'")
			return nil

		case TokenLParen:
			if p.isLoopStart() {
				stmt := p.parseLoopStatement()
				if stmt == nil {
					return nil
				}

				body = append(body, stmt)

				if !p.finishFunctionBodyStatement(stmt) {
					return nil
				}

				continue
			}

			expr, closeParen, ok := p.parseParenthesizedExpression("return list must contain exactly one expression")
			if !ok {
				return nil
			}

			if p.current.Type == TokenLBrace {
				stmt, ok := p.finishConditionalStatement(expr, closeParen)
				if !ok {
					return nil
				}

				body = append(body, stmt)

				if !p.finishFunctionBodyStatement(stmt) {
					return nil
				}

				continue
			}

			returns := []Expression{expr}

			p.skipSeparators()

			if !p.expectCurrent(TokenRBrace, "expected '}' after function literal") {
				return nil
			}

			p.advance()

			return &FunctionLiteral{
				Token:      openBrace,
				Parameters: parameters,
				Body:       body,
				Returns:    returns,
			}

		default:
			stmt := p.parseStatement()
			if stmt != nil {
				body = append(body, stmt)
			}

			if !p.finishFunctionBodyStatement(stmt) {
				return nil
			}
		}
	}
}

func (p *Parser) finishFunctionBodyStatement(stmt Statement) bool {
	if p.current.Type == TokenComma || p.current.Type == TokenNewline {
		p.skipSeparators()
		return true
	}

	if p.current.Type == TokenLParen || p.current.Type == TokenRBrace || p.current.Type == TokenEOF {
		return true
	}

	if statementAllowsTightFollower(stmt) && p.previous.Type == TokenRBrace {
		return true
	}

	p.errorAtCurrent("expected comma, newline, return list, or '}' after function statement")
	return false
}

func (p *Parser) parseFunctionParameters() ([]Token, bool) {
	return p.parseBindingParameters("function parameter")
}

func (p *Parser) parseBindingParameters(name string) ([]Token, bool) {
	openParen := p.current
	p.advance() // consume "("

	parameters := []Token{}

	p.skipNewlines()

	if p.current.Type == TokenRParen {
		p.advance()
		return parameters, true
	}

	for {
		if p.current.Type == TokenEOF {
			p.errorAtToken(openParen, "unterminated parameter list")
			return nil, false
		}

		if p.current.Type != TokenIdentifier {
			p.errorAtCurrent("expected parameter name")
			return nil, false
		}

		parameters = append(parameters, p.current)
		p.advance()

		if p.current.Type == TokenRParen {
			p.advance()

			if !p.validateBindingNames(parameters, name) {
				return nil, false
			}

			return parameters, true
		}

		if p.current.Type != TokenComma && p.current.Type != TokenNewline {
			p.errorAtCurrent("expected comma, newline, or ')' after parameter")
			return nil, false
		}

		p.skipSeparators()

		if p.current.Type == TokenRParen {
			p.advance()

			if !p.validateBindingNames(parameters, name) {
				return nil, false
			}

			return parameters, true
		}
	}
}

func (p *Parser) validateBindingNames(parameters []Token, name string) bool {
	seen := map[string]bool{}

	for _, parameter := range parameters {
		bindingName := parameter.Literal

		if bindingName == "_" {
			continue
		}

		if seen[bindingName] {
			p.errorAtToken(parameter, "duplicate "+name+" "+strconvQuote(bindingName))
			return false
		}

		seen[bindingName] = true
	}

	return true
}

func (p *Parser) parseMapLiteral(openBrace Token) Expression {
	entries := []MapEntry{}

	for {
		if p.current.Type != TokenString {
			p.errorAtCurrent("expected string map key")
			return nil
		}

		key := p.current
		p.advance()

		if !p.expectCurrent(TokenColon, "expected ':' after map key") {
			return nil
		}

		p.advance()

		value := p.parseExpression(precLowest)
		if value == nil {
			p.errorAtToken(key, "expected expression after map key")
			return nil
		}

		entries = append(entries, MapEntry{
			Key:   key,
			Value: value,
		})

		if p.current.Type == TokenRBrace {
			p.advance()
			break
		}

		if p.current.Type == TokenEOF {
			p.errorAtToken(openBrace, "unterminated map literal")
			return nil
		}

		if p.current.Type != TokenComma && p.current.Type != TokenNewline {
			p.errorAtCurrent("expected comma, newline, or '}' after map entry")
			return nil
		}

		p.skipSeparators()

		if p.current.Type == TokenRBrace {
			p.advance()
			break
		}
	}

	return &MapLiteral{Token: openBrace, Entries: entries}
}

func (p *Parser) parseArrayLiteral(openBrace Token) Expression {
	elements := []ArrayElement{}

	for {
		element, ok := p.parseArrayElement(openBrace)
		if !ok {
			return nil
		}

		elements = append(elements, element)

		if p.current.Type == TokenRBrace {
			p.advance()
			break
		}

		if p.current.Type == TokenEOF {
			p.errorAtToken(openBrace, "unterminated array literal")
			return nil
		}

		if p.current.Type != TokenComma && p.current.Type != TokenNewline {
			p.errorAtCurrent("expected comma, newline, or '}' after array element")
			return nil
		}

		p.skipSeparators()

		if p.current.Type == TokenRBrace {
			p.advance()
			break
		}
	}

	return &ArrayLiteral{Token: openBrace, Elements: elements}
}

func (p *Parser) parseArrayElement(openBrace Token) (ArrayElement, bool) {
	start := p.parseExpression(precLowest)
	if start == nil {
		p.errorAtToken(openBrace, "expected array element")
		return nil, false
	}

	if p.current.Type != TokenRangeInclusive && p.current.Type != TokenRangeExclusive {
		return &ExpressionArrayElement{Expression: start}, true
	}

	rangeToken := p.current
	p.advance()

	end := p.parseRangeEndExpression()
	if end == nil {
		p.errorAtToken(rangeToken, "expected range end expression")
		return nil, false
	}

	var step Expression

	if p.current.Type == TokenLBracket {
		if !tokensTouch(p.previous, p.current) {
			p.errorAtCurrent("expected '[' to touch range end expression")
			return nil, false
		}

		var ok bool
		step, ok = p.parseRangeStepExpression()
		if !ok {
			return nil, false
		}
	}

	return &RangeArrayElement{
		Start:       start,
		End:         end,
		Step:        step,
		IsInclusive: rangeToken.Type == TokenRangeInclusive,
	}, true
}

func (p *Parser) parseRangeStepExpression() (Expression, bool) {
	p.advance() // consume "["

	p.skipNewlines()

	step := p.parseExpression(precLowest)
	if step == nil {
		p.errorAtCurrent("expected range step expression")
		return nil, false
	}

	p.skipNewlines()

	if !p.expectCurrent(TokenRBracket, "expected ']' after range step expression") {
		return nil, false
	}

	p.advance()

	return step, true
}

func (p *Parser) parseIndexExpression(object Expression) (Expression, bool) {
	p.advance() // consume "["

	p.skipNewlines()

	index := p.parseExpression(precLowest)
	if index == nil {
		p.errorAtCurrent("expected index expression")
		return nil, false
	}

	p.skipNewlines()

	if !p.expectCurrent(TokenRBracket, "expected ']' after index expression") {
		return nil, false
	}

	p.advance()

	return &IndexExpression{
		Object: object,
		Index:  index,
	}, true
}

func (p *Parser) parseCallExpression(callee Expression) (Expression, bool) {
	p.advance() // consume "("

	args := []Expression{}

	p.skipNewlines()

	if p.current.Type == TokenRParen {
		p.advance()
		return &CallExpression{
			Callee:    callee,
			Arguments: args,
		}, true
	}

	for {
		arg := p.parseExpression(precLowest)
		if arg == nil {
			p.errorAtCurrent("expected function argument")
			return nil, false
		}

		args = append(args, arg)

		if p.current.Type == TokenRParen {
			p.advance()
			break
		}

		if p.current.Type != TokenComma && p.current.Type != TokenNewline {
			p.errorAtCurrent("expected comma, newline, or ')' after function argument")
			return nil, false
		}

		p.skipSeparators()

		if p.current.Type == TokenRParen {
			p.advance()
			break
		}
	}

	return &CallExpression{
		Callee:    callee,
		Arguments: args,
	}, true
}

func (p *Parser) skipSeparators() {
	for p.current.Type == TokenNewline || p.current.Type == TokenComma {
		p.advance()
	}
}

func (p *Parser) skipNewlines() {
	for p.current.Type == TokenNewline {
		p.advance()
	}
}

func (p *Parser) advance() {
	p.previous = p.current
	p.current = p.peek
	p.peek = p.lexer.NextToken()
}

func (p *Parser) expectCurrent(tokenType TokenType, message string) bool {
	if p.current.Type == tokenType {
		return true
	}

	p.errorAtCurrent(message)
	return false
}

func (p *Parser) synchronize() {
	for p.current.Type != TokenEOF {
		if p.current.Type == TokenComma || p.current.Type == TokenNewline {
			p.skipSeparators()
			return
		}

		p.advance()
	}
}

func (p *Parser) errorAtCurrent(message string) {
	p.errorAtToken(p.current, message)
}

func (p *Parser) errorAtToken(tok Token, message string) {
	p.errors = append(
		p.errors,
		fmt.Sprintf(
			"line %d, column %d: %s, got %s %q",
			tok.StartOfLine,
			tok.StartOfColumn,
			message,
			tok.Type,
			tok.Literal,
		),
	)
}

func tokensTouch(left Token, right Token) bool {
	return left.EndOfLine == right.StartOfLine && left.EndOfColumn == right.StartOfColumn
}

func tokensOnSameLine(left Token, right Token) bool {
	return left.EndOfLine == right.StartOfLine
}

func strconvQuote(text string) string {
	return fmt.Sprintf("%q", text)
}
