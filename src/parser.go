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
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		if p.current.Type == TokenComma || p.current.Type == TokenNewline {
			p.skipSeparators()
		} else if p.current.Type != TokenEOF {
			p.errorAtCurrent("expected comma, newline, or end of file after statement")
			p.synchronize()
		}
	}

	return program
}

func (p *Parser) parseStatement() Statement {
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
	if p.current.Type != TokenAssign {
		return &ExpressionStatement{Expression: target}
	}

	assignToken := p.current
	p.advance() // consume "="

	value := p.parseExpression(precLowest)
	if value == nil {
		p.errorAtToken(assignToken, "expected expression after assignment")
		return nil
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

	rangeParameters := []Token{}

	if p.looksLikeLoopRangeParameters(openBrace) {
		parameters, ok := p.parseLoopRangeParameters()
		if !ok {
			return nil, nil, Token{}, false
		}

		rangeParameters = parameters

		if p.current.Type != TokenComma &&
			p.current.Type != TokenNewline &&
			p.current.Type != TokenRBrace &&
			p.current.Type != TokenEOF {
			p.errorAtCurrent("expected comma, newline, or '}' after loop range parameters")
			return nil, nil, Token{}, false
		}
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

		p.errorAtCurrent("expected comma, newline, or '}' after statement")
		return nil, nil, Token{}, false
	}
}

func (p *Parser) looksLikeLoopRangeParameters(openBrace Token) bool {
	if p.current.Type != TokenLParen || !tokensOnSameLine(openBrace, p.current) {
		return false
	}

	trial := p.cloneForLookahead()

	parameters, ok := trial.parseLoopRangeParameters()
	if !ok || len(parameters) != 3 {
		return false
	}

	return trial.current.Type == TokenComma ||
		trial.current.Type == TokenNewline ||
		trial.current.Type == TokenRBrace ||
		trial.current.Type == TokenEOF
}

func (p *Parser) parseLoopRangeParameters() ([]Token, bool) {
	openParen := p.current

	if !p.expectCurrent(TokenLParen, "expected '(' before loop range parameters") {
		return nil, false
	}

	p.advance() // consume "("

	parameters := []Token{}

	p.skipNewlines()

	for {
		if p.current.Type == TokenEOF {
			p.errorAtToken(openParen, "unterminated loop range parameter list")
			return nil, false
		}

		if p.current.Type == TokenRParen {
			p.errorAtToken(openParen, "loop range parameter list cannot be empty")
			return nil, false
		}

		if p.current.Type != TokenIdentifier {
			p.errorAtCurrent("expected loop range parameter name")
			return nil, false
		}

		parameters = append(parameters, p.current)
		p.advance()

		if p.current.Type == TokenRParen {
			p.advance()
			break
		}

		if p.current.Type != TokenComma && p.current.Type != TokenNewline {
			p.errorAtCurrent("expected comma, newline, or ')' after loop range parameter")
			return nil, false
		}

		p.skipSeparators()

		if p.current.Type == TokenRParen {
			p.errorAtCurrent("expected loop range parameter name")
			return nil, false
		}
	}

	if len(parameters) != 3 {
		p.errorAtToken(openParen, "map range loop must declare exactly three parameters: index, key, value")
		return nil, false
	}

	if !p.validateBindingNames(parameters, "loop range parameter") {
		return nil, false
	}

	return parameters, true
}

func (p *Parser) cloneForLookahead() *Parser {
	lexerCopy := *p.lexer
	parserCopy := *p
	parserCopy.lexer = &lexerCopy
	parserCopy.errors = nil
	return &parserCopy
}

func (p *Parser) finishConditionalStatement(condition Expression, closeParen Token) (Statement, bool) {
	if !p.expectCurrent(TokenLBrace, "expected '{' after conditional expression") {
		return nil, false
	}

	if !tokensOnSameLine(closeParen, p.current) {
		p.errorAtCurrent("expected conditional block to start on the same line as the condition")
		return nil, false
	}

	body, closeBrace, ok := p.parseStatementBlock("conditional block")
	if !ok {
		return nil, false
	}

	branches := []ConditionalBranch{
		{
			Condition: condition,
			Body:      body,
		},
	}

	var elseBody []Statement

	for p.current.Type == TokenComma {
		comma := p.current

		if p.peek.Type != TokenLParen && p.peek.Type != TokenLBrace {
			break
		}

		if !tokensTouch(closeBrace, comma) || !tokensTouch(comma, p.peek) {
			p.errorAtToken(comma, "expected conditional separator to be written without whitespace as '},(' or '},{'")
			return nil, false
		}

		switch p.peek.Type {
		case TokenLParen:
			p.advance() // consume "," and move to "("

			condition, closeParen, ok := p.parseParenthesizedExpression("else-if expression cannot be empty")
			if !ok {
				return nil, false
			}

			if !p.expectCurrent(TokenLBrace, "expected '{' after else-if expression") {
				return nil, false
			}

			if !tokensOnSameLine(closeParen, p.current) {
				p.errorAtCurrent("expected else-if block to start on the same line as the else-if condition")
				return nil, false
			}

			body, closeBrace, ok = p.parseStatementBlock("else-if block")
			if !ok {
				return nil, false
			}

			branches = append(branches, ConditionalBranch{
				Condition: condition,
				Body:      body,
			})

		case TokenLBrace:
			p.advance() // consume "," and move to "{"

			elseBody, closeBrace, ok = p.parseStatementBlock("else block")
			if !ok {
				return nil, false
			}

			if p.current.Type == TokenComma &&
				tokensTouch(closeBrace, p.current) &&
				tokensTouch(p.current, p.peek) &&
				(p.peek.Type == TokenLParen || p.peek.Type == TokenLBrace) {
				p.errorAtCurrent("else block must be the final conditional branch")
				return nil, false
			}

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

		p.errorAtCurrent("expected comma, newline, or '}' after statement")
		return nil, Token{}, false
	}
}

const (
	precLowest = iota
	precEquality
	precComparison
	precTerm
	precFactor
	precPrefix
)

func precedence(tok TokenType) int {
	switch tok {
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
	var left Expression

	switch p.current.Type {
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
			left = &EmptyLiteral{
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

	case TokenMinus:
		operator := p.current
		p.advance()

		right := p.parseExpression(precPrefix)
		if right == nil {
			p.errorAtToken(operator, "expected expression after unary '-'")
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

	return p.parseExpressionTail(left, parentPrec)
}

func (p *Parser) parseOuterIdentifierExpression() (Expression, bool) {
	at := p.current
	p.advance() // consume "@"

	if p.current.Type != TokenIdentifier {
		p.errorAtToken(at, "expected identifier after '@'")
		return nil, false
	}

	if p.current.Literal == "_" {
		p.errorAtToken(p.current, "'_' is an empty literal, not an outer binding")
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
	for {
		if p.current.Type == TokenLBracket {
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

		right := p.parseExpression(currentPrec)
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

	if p.current.Type == TokenLParen {
		return p.parseFunctionLiteral(openBrace)
	}

	return p.parseMapLiteral(openBrace)
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

				if !p.finishFunctionBodyStatement() {
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

				if !p.finishFunctionBodyStatement() {
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

			if !p.finishFunctionBodyStatement() {
				return nil
			}
		}
	}
}

func (p *Parser) finishFunctionBodyStatement() bool {
	if p.current.Type == TokenComma || p.current.Type == TokenNewline {
		p.skipSeparators()
		return true
	}

	if p.current.Type == TokenLParen || p.current.Type == TokenRBrace || p.current.Type == TokenEOF {
		return true
	}

	p.errorAtCurrent("expected comma, newline, return list, or '}' after function statement")
	return false
}

func (p *Parser) parseFunctionParameters() ([]Token, bool) {
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

			if !p.validateBindingNames(parameters, "function parameter") {
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

			if !p.validateBindingNames(parameters, "function parameter") {
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

	p.skipNewlines()

	if p.current.Type == TokenRBrace {
		p.advance()
		return &MapLiteral{Token: openBrace, Entries: entries}
	}

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

func (p *Parser) synchronize() {
	for p.current.Type != TokenEOF && p.current.Type != TokenNewline && p.current.Type != TokenComma {
		p.advance()
	}

	p.skipSeparators()
}

func (p *Parser) advance() {
	p.previous = p.current
	p.current = p.peek
	p.peek = p.lexer.NextToken()
}

func (p *Parser) expectCurrent(expected TokenType, message string) bool {
	if p.current.Type == expected {
		return true
	}

	p.errorAtCurrent(message)
	return false
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
