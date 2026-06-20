package main

import "strconv"

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

	case TokenDot:
		leadingDot, ok := p.parseLeadingDotAccessExpression()
		if !ok {
			return nil
		}

		left = leadingDot

	case TokenMinus, TokenNotEqual:
		operator := p.current
		p.advance()

		right := p.parseExpressionWithOptions(precPrefix, stopBeforeTouchingIndex)
		if right == nil {
			p.errorAtToken(operator, "expected expression after unary "+strconv.Quote(operator.Literal))
			return nil
		}

		left = &PrefixExpression{
			Token:    operator,
			Operator: operator.Literal,
			Right:    right,
		}

	case TokenLParen:
		inner, closeParen, ok := p.parseParenthesizedExpression("grouped expression cannot be empty")
		if !ok {
			return nil
		}

		if p.current.Type == TokenQuestion {
			questionExpression, ok := p.parseQuestionExpression(inner, closeParen)
			if !ok {
				return nil
			}

			left = questionExpression
		} else {
			left = inner
		}

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

		if p.current.Type == TokenDot {
			if !tokensTouch(p.previous, p.current) {
				p.errorAtCurrent("expected '.' to touch dot-accessed expression")
				return nil
			}

			index, ok := p.parseDotAccessExpression(left)
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

func (p *Parser) parseLeadingDotAccessExpression() (Expression, bool) {
	dot := p.current
	p.advance() // consume "."

	if p.current.Type != TokenIdentifier {
		p.errorAtCurrent("expected identifier after '.'")
		return nil, false
	}

	if !tokensTouch(dot, p.current) {
		p.errorAtCurrent("expected dot-access identifier to touch '.'")
		return nil, false
	}

	identifier := p.current
	p.advance()

	return &IndexExpression{
		Object: &LeadingDotMapExpression{Token: dot},
		Index: &StringLiteral{
			Token: identifier,
			Value: identifier.Literal,
		},
	}, true
}

func (p *Parser) parseDotAccessExpression(object Expression) (Expression, bool) {
	dot := p.current
	p.advance() // consume "."

	if p.current.Type != TokenIdentifier {
		p.errorAtCurrent("expected identifier after '.'")
		return nil, false
	}

	if !tokensTouch(dot, p.current) {
		p.errorAtCurrent("expected dot-access identifier to touch '.'")
		return nil, false
	}

	identifier := p.current
	p.advance()

	return &IndexExpression{
		Object: object,
		Index: &StringLiteral{
			Token: identifier,
			Value: identifier.Literal,
		},
	}, true
}

func (p *Parser) parseQuestionExpression(target Expression, closeParen Token) (Expression, bool) {
	question := p.current

	if !tokensTouch(closeParen, question) {
		p.errorAtToken(question, "expected '?' to touch parenthesized expression")
		return nil, false
	}

	p.advance() // consume "?"

	switch p.current.Type {
	case TokenLParen:
		return p.parseConditionalExpressionAfterQuestion(target, question)

	case TokenLBrace:
		return p.parseMatchExpressionAfterQuestion(target, question)

	default:
		p.errorAtCurrent("expected '(' or '{' after '?'")
		return nil, false
	}
}

func (p *Parser) parseConditionalExpressionAfterQuestion(condition Expression, question Token) (Expression, bool) {
	if !tokensTouch(question, p.current) {
		p.errorAtCurrent("expected '(' to touch '?' in conditional expression")
		return nil, false
	}

	p.advance() // consume "("
	p.skipNewlines()

	if p.current.Type == TokenRParen {
		p.errorAtToken(question, "conditional expression expected true and false branch expressions")
		return nil, false
	}

	whenTrue := p.parseExpression(precLowest)
	if whenTrue == nil {
		p.errorAtToken(question, "expected true branch expression in conditional expression")
		return nil, false
	}

	if p.current.Type != TokenComma && p.current.Type != TokenNewline {
		p.errorAtCurrent("expected comma or newline after true branch expression")
		return nil, false
	}

	p.skipSeparators()

	if p.current.Type == TokenRParen {
		p.errorAtToken(question, "conditional expression expected false branch expression")
		return nil, false
	}

	whenFalse := p.parseExpression(precLowest)
	if whenFalse == nil {
		p.errorAtToken(question, "expected false branch expression in conditional expression")
		return nil, false
	}

	p.skipNewlines()

	if p.current.Type == TokenComma {
		p.advance()
		p.skipNewlines()
	}

	if !p.expectCurrent(TokenRParen, "expected ')' after conditional expression branches") {
		return nil, false
	}

	p.advance() // consume ")"

	return &ConditionalExpression{
		Token:     question,
		Condition: condition,
		WhenTrue:  whenTrue,
		WhenFalse: whenFalse,
	}, true
}

func (p *Parser) parseMatchExpressionAfterQuestion(target Expression, question Token) (Expression, bool) {
	if !tokensTouch(question, p.current) {
		p.errorAtCurrent("expected '{' to touch '?' in match expression")
		return nil, false
	}

	p.advance() // consume "{"
	p.skipNewlines()

	arms := []MatchArm{}
	var defaultExpression Expression
	seenDefault := false

	seenStringPatterns := map[string]Token{}
	seenNumberPatterns := map[string]Token{}
	seenBoolPatterns := map[string]Token{}

	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		if p.current.Type == TokenIdentifier && p.current.Literal == "_" {
			defaultToken := p.current

			if seenDefault {
				p.errorAtToken(defaultToken, "duplicate default match arm")
				return nil, false
			}

			seenDefault = true
			p.advance()

			if !p.expectCurrent(TokenColon, "expected ':' after match default") {
				return nil, false
			}

			p.advance()

			defaultExpression = p.parseExpression(precLowest)
			if defaultExpression == nil {
				p.errorAtToken(defaultToken, "expected expression after match default")
				return nil, false
			}
		} else {
			pattern, ok := p.parseMatchPattern(
				seenStringPatterns,
				seenNumberPatterns,
				seenBoolPatterns,
			)
			if !ok {
				return nil, false
			}

			if !p.expectCurrent(TokenColon, "expected ':' after match pattern") {
				return nil, false
			}

			p.advance()

			value := p.parseExpression(precLowest)
			if value == nil {
				p.errorAtToken(pattern.Token, "expected expression after match pattern")
				return nil, false
			}

			arms = append(arms, MatchArm{
				Pattern: pattern,
				Value:   value,
			})
		}

		if p.current.Type == TokenRBrace {
			break
		}

		if p.current.Type == TokenEOF {
			p.errorAtToken(question, "unterminated match expression")
			return nil, false
		}

		if p.current.Type != TokenComma && p.current.Type != TokenNewline {
			p.errorAtCurrent("expected comma, newline, or '}' after match arm")
			return nil, false
		}

		p.skipSeparators()
	}

	if !p.expectCurrent(TokenRBrace, "expected '}' after match expression") {
		return nil, false
	}

	p.advance() // consume "}"

	return &MatchExpression{
		Token:   question,
		Target:  target,
		Arms:    arms,
		Default: defaultExpression,
	}, true
}

func (p *Parser) parseMatchPattern(
	seenStringPatterns map[string]Token,
	seenNumberPatterns map[string]Token,
	seenBoolPatterns map[string]Token,
) (MatchPattern, bool) {
	token := p.current

	switch token.Type {
	case TokenString:
		if _, ok := seenStringPatterns[token.Literal]; ok {
			p.errorAtToken(token, "duplicate string match pattern "+strconv.Quote(token.Literal))
			return MatchPattern{}, false
		}

		seenStringPatterns[token.Literal] = token
		p.advance()

		return MatchPattern{
			Token: token,
			Kind:  MatchPatternString,
		}, true

	case TokenNumber:
		if _, ok := seenNumberPatterns[token.Literal]; ok {
			p.errorAtToken(token, "duplicate number match pattern "+strconv.Quote(token.Literal))
			return MatchPattern{}, false
		}

		seenNumberPatterns[token.Literal] = token
		p.advance()

		return MatchPattern{
			Token: token,
			Kind:  MatchPatternNumber,
		}, true

	case TokenBool:
		if _, ok := seenBoolPatterns[token.Literal]; ok {
			p.errorAtToken(token, "duplicate boolean match pattern "+strconv.Quote(token.Literal))
			return MatchPattern{}, false
		}

		seenBoolPatterns[token.Literal] = token
		p.advance()

		return MatchPattern{
			Token: token,
			Kind:  MatchPatternBool,
		}, true

	default:
		p.errorAtCurrent("expected string, number, boolean, or '_' match pattern")
		return MatchPattern{}, false
	}
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
