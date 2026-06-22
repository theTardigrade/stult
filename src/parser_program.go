package main

func (p *Parser) ParseProgram() *Program {
	program := &Program{}

	for p.current.Type != TokenEOF {
		p.skipSeparators()

		if p.current.Type == TokenEOF {
			break
		}

		stmt := p.parseStatement()
		if stmt == nil {
			p.synchronize()
			continue
		}

		canFollowTightly := statementAllowsTightFollower(stmt) && p.previous.Type == TokenRBrace

		program.Statements = append(program.Statements, stmt)

		if p.current.Type == TokenComma || p.current.Type == TokenNewline {
			if !p.rejectOldCommaBlockSeparator() {
				return program
			}

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
