package main

import "fmt"

func (p *Parser) parseBindingContractAfterToken(
	annotated Token,
	description string,
) (*BindingContractDeclaration, bool) {
	if !isBindingContractStart(p.current.Type) {
		return nil, true
	}

	start := p.current
	if !tokensTouch(annotated, start) {
		p.errorAtCurrent("expected binding contract to touch " + description)
		return nil, false
	}

	switch start.Type {
	case TokenContractSameKind, TokenContractAny:
		contract, ok := p.parseBindingContractType()
		if !ok {
			return nil, false
		}

		return &BindingContractDeclaration{
			Token:    start,
			Contract: contract,
		}, true

	case TokenLess:
		p.advance()
		contract, ok := p.parseBindingContractType()
		if !ok {
			return nil, false
		}

		if !p.expectCurrent(TokenGreater, "expected '>' after binding contract") {
			return nil, false
		}

		p.advance()
		return &BindingContractDeclaration{
			Token:    start,
			Contract: contract,
		}, true

	default:
		p.errorAtCurrent("expected binding contract")
		return nil, false
	}
}

func (p *Parser) parseBindingContractType() (BindingContract, bool) {
	left, ok := p.parseBindingContractTerm()
	if !ok {
		return BindingContract{}, false
	}

	if p.current.Type != TokenOr {
		return left, true
	}

	options := []BindingContract{left}
	for p.current.Type == TokenOr {
		p.advance()

		right, ok := p.parseBindingContractTerm()
		if !ok {
			return BindingContract{}, false
		}

		options = append(options, right)
	}

	return BindingContract{
		Kind:    BindingContractUnionKind,
		Options: options,
	}, true
}

func (p *Parser) parseBindingContractTerm() (BindingContract, bool) {
	switch p.current.Type {
	case TokenContractSameKind:
		p.advance()
		return BindingContract{Kind: BindingContractSameKind}, true

	case TokenContractAny:
		p.advance()
		return BindingContract{}, true

	case TokenIdentifier:
		if p.current.Literal == "_" {
			p.errorAtCurrent("expected contract alias name")
			return BindingContract{}, false
		}

		if p.current.Literal == "STD" {
			return p.parseNamedBindingContractType()
		}

		alias := p.current.Literal
		p.advance()
		return BindingContract{Kind: BindingContractAliasKind, AliasName: alias}, true

	default:
		p.errorAtCurrent("expected binding contract")
		return BindingContract{}, false
	}
}

func (p *Parser) parseNamedBindingContractType() (BindingContract, bool) {
	pathTokens := []Token{}

	if p.current.Type != TokenIdentifier || p.current.Literal != "STD" {
		p.errorAtCurrent("expected STD.TYPE contract path")
		return BindingContract{}, false
	}

	pathTokens = append(pathTokens, p.current)
	previous := p.current
	p.advance()

	for len(pathTokens) < 3 {
		if p.current.Type != TokenDot || !tokensTouch(previous, p.current) {
			p.errorAtCurrent("expected '.' in STD.TYPE contract path")
			return BindingContract{}, false
		}

		dot := p.current
		p.advance()

		if p.current.Type != TokenIdentifier || !tokensTouch(dot, p.current) {
			p.errorAtCurrent("expected identifier in STD.TYPE contract path")
			return BindingContract{}, false
		}

		pathTokens = append(pathTokens, p.current)
		previous = p.current
		p.advance()
	}

	if pathTokens[1].Literal != "TYPE" {
		p.errorAtToken(pathTokens[1], "expected STD.TYPE contract path")
		return BindingContract{}, false
	}

	baseName := pathTokens[2].Literal
	contract, ok := bindingContractForStdTypeName(baseName)
	if !ok {
		p.errorAtToken(pathTokens[2], fmt.Sprintf("unknown STD.TYPE contract %q", baseName))
		return BindingContract{}, false
	}

	if p.current.Type == TokenLess || p.current.Type == TokenContractSameKind || p.current.Type == TokenContractAny {
		if !tokensTouch(previous, p.current) {
			p.errorAtCurrent("expected nested binding contract to touch its collection contract")
			return BindingContract{}, false
		}

		if baseName != "ARRAY" && baseName != "MAP" {
			p.errorAtCurrent("only STD.TYPE.ARRAY and STD.TYPE.MAP contracts can take nested contracts")
			return BindingContract{}, false
		}

		var element BindingContract
		if p.current.Type == TokenLess {
			p.advance()

			var ok bool
			element, ok = p.parseBindingContractType()
			if !ok {
				return BindingContract{}, false
			}

			if !p.expectCurrent(TokenGreater, "expected '>' after nested binding contract") {
				return BindingContract{}, false
			}

			p.advance()
		} else {
			var ok bool
			element, ok = p.parseBindingContractType()
			if !ok {
				return BindingContract{}, false
			}
		}

		contract.Element = element.ClonePointer()
	}

	return contract, true
}

func bindingContractForStdTypeName(name string) (BindingContract, bool) {
	switch name {
	case "VOID":
		return BindingContract{Kind: BindingContractExactKind, KindValue: ValueVoid}, true
	case "NUMBER":
		return BindingContract{Kind: BindingContractExactKind, KindValue: ValueNumber}, true
	case "BOOL":
		return BindingContract{Kind: BindingContractExactKind, KindValue: ValueBool}, true
	case "STRING":
		return BindingContract{Kind: BindingContractExactKind, KindValue: ValueString}, true
	case "FUNCTION":
		return BindingContract{Kind: BindingContractExactKind, KindValue: ValueFunction}, true
	case "BUILTIN_FUNCTION":
		return BindingContract{Kind: BindingContractExactKind, KindValue: ValueBuiltinFunction}, true
	case "CONTRACT":
		return BindingContract{Kind: BindingContractExactKind, KindValue: ValueContract}, true
	case "COLLECTION":
		return BindingContract{
			Kind: BindingContractUnionKind,
			Options: []BindingContract{
				{Kind: BindingContractArrayKind},
				{Kind: BindingContractMapKind},
				{Kind: BindingContractExactKind, KindValue: ValueString},
			},
		}, true
	case "ARRAY":
		return BindingContract{Kind: BindingContractArrayKind}, true
	case "MAP":
		return BindingContract{Kind: BindingContractMapKind}, true
	default:
		return BindingContract{}, false
	}
}
