package main

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

	for p.current.Type == TokenOr {
		separator := p.current

		switch p.peek.Type {
		case TokenLParen:
			if elseBody != nil {
				p.errorAtToken(separator, "else-if cannot appear after else")
				return nil, false
			}

			if !tokensOnSameLine(closeBrace, separator) || !tokensOnSameLine(separator, p.peek) {
				p.errorAtToken(separator, "expected else-if separator to stay on one line as '}|('")
				return nil, false
			}

			p.advance() // consume "|" and move to "("

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
			if !tokensOnSameLine(closeBrace, separator) || !tokensOnSameLine(separator, p.peek) {
				p.errorAtToken(separator, "expected else separator to stay on one line as '}|{'")
				return nil, false
			}

			p.advance() // consume "|" and move to "{"

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

func (p *Parser) parseIndexExpression(object Expression) (Expression, bool) {
	openBracket := p.current
	p.advance() // consume "["

	p.skipNewlines()

	index := p.parseExpression(precLowest)
	if index == nil {
		p.errorAtCurrent("expected index expression")
		return nil, false
	}

	if p.current.Type == TokenRangeInclusive || p.current.Type == TokenRangeExclusive {
		rangeToken := p.current
		p.advance()

		end := p.parseRangeEndExpression()
		if end == nil {
			p.errorAtToken(rangeToken, "expected range index end expression")
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

		p.skipNewlines()

		if !p.expectCurrent(TokenRBracket, "expected ']' after range index expression") {
			return nil, false
		}

		p.advance()

		return &RangeIndexExpression{
			Object:      object,
			Start:       index,
			End:         end,
			Step:        step,
			IsInclusive: rangeToken.Type == TokenRangeInclusive,
		}, true
	}

	p.skipNewlines()

	if !p.expectCurrent(TokenRBracket, "expected ']' after index expression") {
		p.errorAtToken(openBracket, "unterminated index expression")
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
