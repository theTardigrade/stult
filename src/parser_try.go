package main

func (p *Parser) parseTryCatchStatement() Statement {
	tryToken := p.current
	p.advance() // consume "'"

	if !p.expectCurrent(TokenLBrace, "expected '{' after try marker") {
		return nil
	}

	if !tokensTouch(tryToken, p.current) {
		p.errorAtCurrent("expected try marker and try block to be written without whitespace as \"'{\"")
		return nil
	}

	tryBody, closeBrace, ok := p.parseStatementBlock("try block")
	if !ok {
		return nil
	}

	if p.current.Type != TokenComma || p.peek.Type != TokenLBrace {
		p.errorAtCurrent("expected catch separator written without whitespace as '},{'")
		return nil
	}

	comma := p.current
	if !tokensTouch(closeBrace, comma) || !tokensTouch(comma, p.peek) {
		p.errorAtToken(comma, "expected catch separator to be written without whitespace as '},{'")
		return nil
	}

	p.advance() // consume "," and move to "{"

	catchParameter, catchBody, ok := p.parseCatchBlock()
	if !ok {
		return nil
	}

	return &TryCatchStatement{
		Token:          tryToken,
		TryBody:        tryBody,
		CatchParameter: catchParameter,
		CatchBody:      catchBody,
	}
}

func (p *Parser) parseCatchBlock() (*Token, []Statement, bool) {
	openBrace := p.current

	if !p.expectCurrent(TokenLBrace, "expected '{' to start catch block") {
		return nil, nil, false
	}

	p.advance() // consume "{"
	p.skipSeparators()

	var catchParameter *Token

	if p.current.Type == TokenLParen && !p.isLoopStart() {
		checkpoint := p.checkpoint()
		parameters, ok := p.parseBindingParameters("catch parameter")
		if ok {
			if len(parameters) > 1 {
				p.errorAtToken(parameters[1], "catch block may have at most one parameter")
				return nil, nil, false
			}

			if len(parameters) == 1 {
				parameter := parameters[0]
				catchParameter = &parameter
			}
		} else {
			p.restore(checkpoint)
		}
	}

	body := []Statement{}

	for {
		p.skipSeparators()

		switch p.current.Type {
		case TokenEOF:
			p.errorAtToken(openBrace, "unterminated catch block")
			return nil, nil, false

		case TokenRBrace:
			p.advance()
			return catchParameter, body, true
		}

		stmt := p.parseStatement()
		canFollowTightly := statementAllowsTightFollower(stmt) && p.previous.Type == TokenRBrace

		if stmt == nil {
			return nil, nil, false
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

		p.errorAtCurrent("expected comma, newline, or '}' after catch statement")
		return nil, nil, false
	}
}
