package main

func (p *Parser) tokensAhead(count int) []Token {
	if count <= 0 {
		return nil
	}

	tokens := []Token{p.current}
	if count == 1 {
		return tokens
	}

	tokens = append(tokens, p.peek)
	lexerCopy := *p.lexer
	for len(tokens) < count {
		tokens = append(tokens, lexerCopy.NextToken())
	}

	return tokens
}

func (p *Parser) currentLessLooksLikeBindingContractBeforeAssignment() bool {
	if p.current.Type != TokenLess || !tokensTouch(p.previous, p.current) {
		return false
	}

	tokens := p.tokensAhead(80)
	depth := 0
	for index, token := range tokens {
		switch token.Type {
		case TokenLess:
			depth++
		case TokenGreater:
			depth--
			if depth == 0 {
				if index+1 >= len(tokens) {
					return false
				}
				return isAssignmentOperator(tokens[index+1].Type)
			}
			if depth < 0 {
				return false
			}
		case TokenNewline, TokenComma, TokenEOF:
			return false
		}
	}

	return false
}
