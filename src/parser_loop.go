package main

func (p *Parser) isLoopStart() bool {
	return p.current.Type == TokenLParen &&
		p.peek.Type == TokenLParen &&
		tokensTouch(p.current, p.peek)
}

func (p *Parser) isLoopStatementStart() bool {
	if !p.isLoopStart() {
		return false
	}

	checkpoint := p.checkpoint()
	ok := p.parseLoopHeaderShape()
	p.restore(checkpoint)

	return ok
}

func (p *Parser) parseLoopHeaderShape() bool {
	outerOpen := p.current

	p.advance() // consume first "(" and move to second "("
	p.advance() // consume second "(" and move to loop expression

	if p.current.Type == TokenRParen {
		p.errorAtToken(outerOpen, "loop expression cannot be empty")
		return false
	}

	condition := p.parseExpression(precLowest)
	if condition == nil {
		p.errorAtToken(outerOpen, "expected loop expression")
		return false
	}

	innerClose := p.current
	if !p.expectCurrent(TokenRParen, "expected ')' after loop expression") {
		return false
	}

	p.advance() // consume inner ")"

	outerClose := p.current
	if !p.expectCurrent(TokenRParen, "expected second ')' after loop expression") {
		return false
	}

	if !tokensTouch(innerClose, outerClose) {
		p.errorAtToken(outerClose, "expected loop expression to close with touching '))'")
		return false
	}

	p.advance() // consume outer ")"

	if !p.expectCurrent(TokenLBrace, "expected '{' after loop expression") {
		return false
	}

	if !tokensOnSameLine(outerClose, p.current) {
		p.errorAtCurrent("expected loop block to start on the same line as the loop expression")
		return false
	}

	return true
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

	if p.current.Type == TokenOr && p.peek.Type == TokenLBrace {
		separator := p.current

		if !tokensTouch(closeBrace, separator) || !tokensTouch(separator, p.peek) {
			p.errorAtToken(separator, "expected after-loop separator to be written without whitespace as '}|{'")
			return nil
		}

		p.advance() // consume "|" and move to "{"

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
