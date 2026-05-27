package main

import "fmt"

type Parser struct {
	lexer *Lexer

	previous Token
	current  Token
	peek     Token

	errors []string
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{lexer: lexer}

	p.advance()
	p.advance()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) skipSeparators() {
	for p.current.Type == TokenNewline || p.current.Type == TokenComma {
		p.advance()
	}
}

func (p *Parser) skipNewlines() {
	for p.current.Type == TokenNewline {
		p.advance()
	}
}

func (p *Parser) advance() {
	p.previous = p.current
	p.current = p.peek
	p.peek = p.lexer.NextToken()
}

func (p *Parser) expectCurrent(tokenType TokenType, message string) bool {
	if p.current.Type == tokenType {
		return true
	}

	p.errorAtCurrent(message)
	return false
}

func (p *Parser) synchronize() {
	for p.current.Type != TokenEOF {
		if p.current.Type == TokenComma || p.current.Type == TokenNewline {
			p.skipSeparators()
			return
		}

		p.advance()
	}
}

func (p *Parser) errorAtCurrent(message string) {
	p.errorAtToken(p.current, message)
}

func (p *Parser) errorAtToken(tok Token, message string) {
	p.errors = append(
		p.errors,
		fmt.Sprintf(
			"line %d, column %d: %s, got %s %q",
			tok.StartOfLine,
			tok.StartOfColumn,
			message,
			tok.Type,
			tok.Literal,
		),
	)
}

func tokensTouch(left Token, right Token) bool {
	return left.EndOfLine == right.StartOfLine && left.EndOfColumn == right.StartOfColumn
}

func tokensOnSameLine(left Token, right Token) bool {
	return left.EndOfLine == right.StartOfLine
}

func strconvQuote(text string) string {
	return fmt.Sprintf("%q", text)
}
