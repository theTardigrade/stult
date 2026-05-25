package main

import "fmt"

type Parser struct {
	lexer *Lexer

	current Token
	peek    Token

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
		} else if p.current.Type != TokenEOF && p.current.Type != TokenRBrace {
			p.errorAtCurrent("expected comma, newline, or end of file after statement")
			p.synchronize()
		}
	}

	return program
}

func (p *Parser) parseStatement() Statement {
	target := p.parseExpression(precLowest)
	if target == nil {
		return nil
	}

	if p.current.Type != TokenAssign {
		return &ExpressionStatement{Expression: target}
	}

	p.advance() // consume "="

	value := p.parseExpression(precLowest)
	if value == nil {
		p.errorAtCurrent("expected expression after assignment")
		return nil
	}

	switch t := target.(type) {
	case *IdentifierExpression:
		return &AssignmentStatement{
			Name:        t.Token,
			Value:       value,
			IsImmutable: t.IsImmutable,
		}

	case *IndexExpression:
		return &IndexAssignmentStatement{
			Target: t,
			Value:  value,
		}

	default:
		p.errorAtCurrent("invalid assignment target")
		return nil
	}
}

func (p *Parser) parseExpressionStatement() Statement {
	expr := p.parseExpression(precLowest)
	if expr == nil {
		return nil
	}

	return &ExpressionStatement{Expression: expr}
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
		left = &IdentifierExpression{
			Token:       p.current,
			Name:        p.current.Literal,
			IsImmutable: p.current.IsImmutable,
		}
		p.advance()

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
		p.advance()

		inner := p.parseExpression(precLowest)
		if inner == nil {
			return nil
		}

		if !p.expectCurrent(TokenRParen, "expected ')' after expression") {
			return nil
		}

		p.advance()
		left = inner

	case TokenLBrace:
		left = p.parseMapLiteral()
		if left == nil {
			return nil
		}

	default:
		p.errorAtCurrent("expected expression")
		return nil
	}

	for {
		if p.current.Type == TokenLBracket {
			index, ok := p.parseIndexExpression(left)
			if !ok {
				return nil
			}

			left = index
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

func (p *Parser) parseMapLiteral() Expression {
	openBrace := p.current
	p.advance() // consume "{"

	entries := []MapEntry{}

	p.skipSeparators()

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

		if p.current.Type == TokenComma || p.current.Type == TokenNewline {
			p.skipSeparators()
		}

		if p.current.Type == TokenRBrace {
			p.advance()
			break
		}

		if p.current.Type == TokenEOF {
			p.errorAtToken(openBrace, "unterminated map literal")
			return nil
		}

		if p.current.Type != TokenString {
			p.errorAtCurrent("expected comma, newline, or '}' after map entry")
			return nil
		}
	}

	return &MapLiteral{Token: openBrace, Entries: entries}
}

func (p *Parser) parseIndexExpression(object Expression) (Expression, bool) {
	p.advance() // consume "["

	index := p.parseExpression(precLowest)
	if index == nil {
		p.errorAtCurrent("expected index expression")
		return nil, false
	}

	if !p.expectCurrent(TokenRBracket, "expected ']' after index expression") {
		return nil, false
	}

	p.advance()

	return &IndexExpression{
		Object: object,
		Index:  index,
	}, true
}

func (p *Parser) skipSeparators() {
	for p.current.Type == TokenNewline || p.current.Type == TokenComma {
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
			tok.Line,
			tok.Column,
			message,
			tok.Type,
			tok.Literal,
		),
	)
}
