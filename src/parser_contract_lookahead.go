package main

// Binding-contract lookahead is bounded so malformed input cannot make
// statement disambiguation scan arbitrarily far. The limit is intentionally
// generous because structured map contracts can be large and deeply nested.
const maxBindingContractLookaheadTokens = 1 << 16

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

	tokens := p.tokensAhead(maxBindingContractLookaheadTokens)
	angleDepth := 0
	braceDepth := 0
	for index, token := range tokens {
		switch token.Type {
		case TokenLess:
			angleDepth++
		case TokenGreater:
			angleDepth--
			if angleDepth == 0 {
				if index+1 >= len(tokens) {
					return false
				}
				return isAssignmentOperator(tokens[index+1].Type)
			}
			if angleDepth < 0 {
				return false
			}
		case TokenLBrace:
			if angleDepth > 0 {
				braceDepth++
			}
		case TokenRBrace:
			if braceDepth > 0 {
				braceDepth--
			}
		case TokenNewline, TokenComma:
			if braceDepth == 0 {
				return false
			}
		case TokenEOF:
			return false
		}
	}

	return false
}
