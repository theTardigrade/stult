package main

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
		if function := p.tryParseFunctionLiteral(openBrace); function != nil {
			return function
		}
	}

	if p.isMapLiteralEntryStart() {
		return p.parseMapLiteral(openBrace)
	}

	return p.parseArrayLiteral(openBrace)
}

func (p *Parser) tryParseFunctionLiteral(openBrace Token) Expression {
	checkpoint := p.checkpoint()

	function := p.parseFunctionLiteral(openBrace)
	if function != nil {
		return function
	}

	p.restore(checkpoint)
	return nil
}

func (p *Parser) parseFunctionLiteral(openBrace Token) Expression {
	parameters, variadicParameter, ok := p.parseFunctionParameters()
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
			if p.isLoopStatementStart() {
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
				Token:             openBrace,
				Parameters:        parameters,
				VariadicParameter: variadicParameter,
				Body:              body,
				Returns:           returns,
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

func (p *Parser) isMapLiteralEntryStart() bool {
	return (p.current.Type == TokenString && p.peek.Type == TokenColon) ||
		p.isLeadingDotMapKeyStart()
}

func (p *Parser) isLeadingDotMapKeyStart() bool {
	if p.current.Type != TokenDot || p.peek.Type != TokenIdentifier {
		return false
	}

	checkpoint := p.checkpoint()
	dot := p.current
	p.advance()

	touchesIdentifier := tokensTouch(dot, p.current)
	p.advance()

	isMapKey := touchesIdentifier && p.current.Type == TokenColon
	p.restore(checkpoint)

	return isMapKey
}

func (p *Parser) parseMapLiteral(openBrace Token) Expression {
	entries := []MapEntry{}

	for {
		key, isDotKey, ok := p.parseMapEntryKey()
		if !ok {
			return nil
		}

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
			Key:      key,
			Value:    value,
			IsDotKey: isDotKey,
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

func (p *Parser) parseMapEntryKey() (Token, bool, bool) {
	if p.current.Type == TokenString {
		key := p.current
		p.advance()
		return key, false, true
	}

	if p.current.Type != TokenDot {
		p.errorAtCurrent("expected string map key or leading-dot map key")
		return Token{}, false, false
	}

	dot := p.current
	p.advance() // consume "."

	if p.current.Type != TokenIdentifier {
		p.errorAtToken(dot, "expected identifier after leading-dot map key")
		return Token{}, false, false
	}

	if !tokensTouch(dot, p.current) {
		p.errorAtToken(p.current, "expected leading-dot map key identifier to touch '.'")
		return Token{}, false, false
	}

	key := p.current
	p.advance()

	return key, true, true
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

	if p.current.Type == TokenColon {
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
	colon := p.current
	p.advance() // consume ":"

	if p.current.Type == TokenNewline {
		p.errorAtCurrent("expected range step expression on the same line as ':'")
		return nil, false
	}

	step := p.parseExpression(precLowest)
	if step == nil {
		p.errorAtToken(colon, "expected range step expression")
		return nil, false
	}

	return step, true
}
