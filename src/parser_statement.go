package main

func (p *Parser) parseStatement() Statement {
	if p.current.Type == TokenCaret {
		return p.parseCaretStatement()
	}

	if p.isLoopStatementStart() {
		return p.parseLoopStatement()
	}

	if p.current.Type == TokenTry {
		return p.parseTryCatchStatement()
	}

	if p.current.Type == TokenLParen {
		expr, closeParen, ok := p.parseParenthesizedExpression("conditional expression cannot be empty")
		if !ok {
			return nil
		}

		if p.current.Type == TokenLBrace {
			stmt, ok := p.finishConditionalStatement(expr, closeParen)
			if !ok {
				return nil
			}

			return stmt
		}

		target := expr

		if p.current.Type == TokenQuestion {
			questionExpression, ok := p.parseQuestionExpression(expr, closeParen)
			if !ok {
				return nil
			}

			target = questionExpression
		}

		target = p.parseExpressionTail(target, precLowest)
		if target == nil {
			return nil
		}

		return p.finishExpressionOrAssignmentStatement(target)
	}

	target := p.parseExpression(precLowest)
	if target == nil {
		return nil
	}

	return p.finishExpressionOrAssignmentStatement(target)
}

func (p *Parser) finishExpressionOrAssignmentStatement(target Expression) Statement {
	if !isAssignmentOperator(p.current.Type) {
		return &ExpressionStatement{Expression: target}
	}

	assignToken := p.current
	p.advance() // consume assignment operator

	value := p.parseExpression(precLowest)
	if value == nil {
		p.errorAtToken(assignToken, "expected expression after assignment")
		return nil
	}

	if isCompoundAssignmentOperator(assignToken.Type) {
		if !isAssignableExpression(target) {
			p.errorAtToken(assignToken, "invalid assignment target")
			return nil
		}

		return &CompoundAssignmentStatement{
			Target:   target,
			Operator: assignToken,
			Value:    value,
		}
	}

	switch t := target.(type) {
	case *IdentifierExpression:
		return &AssignmentStatement{
			Name:        t.Token,
			Value:       value,
			IsImmutable: t.IsImmutable,
			IsOuter:     t.IsOuter,
		}

	case *IndexExpression:
		return &IndexAssignmentStatement{
			Target: t,
			Value:  value,
		}

	default:
		p.errorAtToken(assignToken, "invalid assignment target")
		return nil
	}
}

func isAssignmentOperator(tokenType TokenType) bool {
	return tokenType == TokenColon || isCompoundAssignmentOperator(tokenType)
}

func isCompoundAssignmentOperator(tokenType TokenType) bool {
	return tokenType == TokenPlusAssign ||
		tokenType == TokenMinusAssign ||
		tokenType == TokenStarAssign ||
		tokenType == TokenSlashAssign
}

func isAssignableExpression(expr Expression) bool {
	switch expr.(type) {
	case *IdentifierExpression, *IndexExpression:
		return true
	default:
		return false
	}
}

func (p *Parser) parseCaretStatement() Statement {
	caret := p.current
	p.advance() // consume "^"

	if p.current.Type != TokenLParen || !tokensTouch(caret, p.current) {
		return &BreakStatement{Token: caret}
	}

	value, _, ok := p.parseParenthesizedExpression("return expression cannot be empty")
	if !ok {
		return nil
	}

	return &ReturnStatement{
		Token: caret,
		Value: value,
	}
}

func statementAllowsTightFollower(stmt Statement) bool {
	switch stmt.(type) {
	case *ConditionalStatement, *LoopStatement, *TryCatchStatement:
		return true
	default:
		return false
	}
}
