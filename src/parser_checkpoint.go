package main

type parserCheckpoint struct {
	previous  Token
	current   Token
	peek      Token
	errorsLen int
	lexer     lexerCheckpoint
}

func (p *Parser) checkpoint() parserCheckpoint {
	return parserCheckpoint{
		previous:  p.previous,
		current:   p.current,
		peek:      p.peek,
		errorsLen: len(p.errors),
		lexer:     p.lexer.checkpoint(),
	}
}

func (p *Parser) restore(checkpoint parserCheckpoint) {
	p.previous = checkpoint.previous
	p.current = checkpoint.current
	p.peek = checkpoint.peek

	if checkpoint.errorsLen < len(p.errors) {
		p.errors = p.errors[:checkpoint.errorsLen]
	}

	p.lexer.restore(checkpoint.lexer)
}
