package main

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
			if !p.rejectOldCommaBlockSeparator() {
				return nil, Token{}, false
			}

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

func (p *Parser) parseLoopBodyBlock(name string) ([]Token, []Statement, Token, bool) {
	openBrace := p.current

	if !p.expectCurrent(TokenLBrace, "expected '{' to start "+name) {
		return nil, nil, Token{}, false
	}

	p.advance() // consume "{"
	p.skipSeparators()

	rangeParameters := []Token{}

	if parameters, ok := p.parseOptionalLoopRangeParameters(); ok {
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
			if !p.rejectOldCommaBlockSeparator() {
				return nil, nil, Token{}, false
			}

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

func (p *Parser) parseOptionalLoopRangeParameters() ([]Token, bool) {
	if p.current.Type != TokenLParen || p.isLoopStart() {
		return nil, false
	}

	checkpoint := p.checkpoint()

	parameters, ok := p.parseBindingParameters("range parameter")
	if !ok {
		p.restore(checkpoint)
		return nil, false
	}

	return parameters, true
}

func (p *Parser) finishFunctionBodyStatement(stmt Statement) bool {
	if p.current.Type == TokenComma || p.current.Type == TokenNewline {
		if !p.rejectOldCommaBlockSeparator() {
			return false
		}

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
