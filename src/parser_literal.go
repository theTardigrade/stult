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
		return p.parseFunctionLiteral(openBrace)
	}

	if p.current.Type == TokenString && p.peek.Type == TokenColon {
		return p.parseMapLiteral(openBrace)
	}

	return p.parseArrayLiteral(openBrace)
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
