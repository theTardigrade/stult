package main

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
		inner, closeParen, ok := p.parseParenthesizedExpression("grouped expression cannot be empty")
		if !ok {
			return nil
		}

		if p.current.Type == TokenQuestion {
			conditional, ok := p.parseConditionalExpression(inner, closeParen)
			if !ok {
				return nil
			}

			left = conditional
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

func (p *Parser) parseConditionalExpression(condition Expression, closeParen Token) (Expression, bool) {
	question := p.current

	if !tokensTouch(closeParen, question) {
		p.errorAtToken(question, "expected '?' to touch parenthesized condition")
		return nil, false
	}

	p.advance() // consume "?"

	if !p.expectCurrent(TokenLParen, "expected '(' after '?' in conditional expression") {
		return nil, false
	}

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
