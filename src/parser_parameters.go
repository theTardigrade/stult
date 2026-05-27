package main

func (p *Parser) parseFunctionParameters() ([]Token, bool) {
	return p.parseBindingParameters("function parameter")
}

func (p *Parser) parseBindingParameters(name string) ([]Token, bool) {
	openParen := p.current
	p.advance() // consume "("

	parameters := []Token{}

	p.skipNewlines()

	if p.current.Type == TokenRParen {
		p.advance()
		return parameters, true
	}

	for {
		if p.current.Type == TokenEOF {
			p.errorAtToken(openParen, "unterminated parameter list")
			return nil, false
		}

		if p.current.Type != TokenIdentifier {
			p.errorAtCurrent("expected parameter name")
			return nil, false
		}

		parameters = append(parameters, p.current)
		p.advance()

		if p.current.Type == TokenRParen {
			p.advance()

			if !p.validateBindingNames(parameters, name) {
				return nil, false
			}

			return parameters, true
		}

		if p.current.Type != TokenComma && p.current.Type != TokenNewline {
			p.errorAtCurrent("expected comma, newline, or ')' after parameter")
			return nil, false
		}

		p.skipSeparators()

		if p.current.Type == TokenRParen {
			p.advance()

			if !p.validateBindingNames(parameters, name) {
				return nil, false
			}

			return parameters, true
		}
	}
}

func (p *Parser) validateBindingNames(parameters []Token, name string) bool {
	seen := map[string]bool{}

	for _, parameter := range parameters {
		bindingName := parameter.Literal

		if bindingName == "_" {
			continue
		}

		if seen[bindingName] {
			p.errorAtToken(parameter, "duplicate "+name+" "+strconvQuote(bindingName))
			return false
		}

		seen[bindingName] = true
	}

	return true
}
