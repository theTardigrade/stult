package main

func (p *Parser) parseFunctionParameters() ([]FunctionParameter, *Token, bool) {
	openParen := p.current
	p.advance() // consume "("

	parameters := []FunctionParameter{}
	optionalStarted := false

	p.skipNewlines()

	if p.current.Type == TokenRParen {
		p.advance()
		return parameters, nil, true
	}

	for {
		if p.current.Type == TokenEOF {
			p.errorAtToken(openParen, "unterminated parameter list")
			return nil, nil, false
		}

		if p.current.Type == TokenRangeExclusive {
			ellipsis := p.current
			p.advance()

			if p.current.Type != TokenIdentifier {
				p.errorAtToken(ellipsis, "expected parameter name after '...'")
				return nil, nil, false
			}

			variadicParameter := p.current
			p.advance()

			if p.current.Type == TokenQuestion {
				p.errorAtCurrent("variadic function parameter cannot be optional")
				return nil, nil, false
			}

			p.skipNewlines()

			if p.current.Type != TokenRParen {
				p.errorAtCurrent("variadic function parameter must be last")
				return nil, nil, false
			}

			p.advance()

			allParameters := functionParameterTokens(parameters)
			allParameters = append(allParameters, variadicParameter)

			if !p.validateBindingNames(allParameters, "function parameter") {
				return nil, nil, false
			}

			return parameters, &variadicParameter, true
		}

		if p.current.Type != TokenIdentifier {
			p.errorAtCurrent("expected parameter name")
			return nil, nil, false
		}

		parameterToken := p.current
		p.advance()

		isOptional := false

		if p.current.Type == TokenQuestion {
			if !tokensTouch(parameterToken, p.current) {
				p.errorAtCurrent("expected '?' to touch optional parameter name")
				return nil, nil, false
			}

			isOptional = true
			optionalStarted = true
			p.advance()
		} else if optionalStarted {
			p.errorAtToken(parameterToken, "required function parameter cannot follow optional parameter")
			return nil, nil, false
		}

		parameters = append(parameters, FunctionParameter{
			Token:      parameterToken,
			IsOptional: isOptional,
		})

		if p.current.Type == TokenRParen {
			p.advance()

			if !p.validateBindingNames(functionParameterTokens(parameters), "function parameter") {
				return nil, nil, false
			}

			return parameters, nil, true
		}

		if p.current.Type != TokenComma && p.current.Type != TokenNewline {
			p.errorAtCurrent("expected comma, newline, or ')' after parameter")
			return nil, nil, false
		}

		p.skipSeparators()

		if p.current.Type == TokenRParen {
			p.advance()

			if !p.validateBindingNames(functionParameterTokens(parameters), "function parameter") {
				return nil, nil, false
			}

			return parameters, nil, true
		}
	}
}

func functionParameterTokens(parameters []FunctionParameter) []Token {
	tokens := make([]Token, 0, len(parameters))

	for _, parameter := range parameters {
		tokens = append(tokens, parameter.Token)
	}

	return tokens
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
