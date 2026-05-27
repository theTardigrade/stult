package main

func (p *Parser) ParseProgram() *Program {
	program := &Program{}

	for p.current.Type != TokenEOF {
		p.skipSeparators()

		if p.current.Type == TokenEOF {
			break
		}

		stmt := p.parseStatement()
		canFollowTightly := statementAllowsTightFollower(stmt) && p.previous.Type == TokenRBrace

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		if p.current.Type == TokenComma || p.current.Type == TokenNewline {
			p.skipSeparators()
		} else if p.current.Type == TokenEOF {
			break
		} else if canFollowTightly {
			continue
		} else {
			p.errorAtCurrent("expected comma, newline, or end of file after statement")
			p.synchronize()
		}
	}

	return program
}
