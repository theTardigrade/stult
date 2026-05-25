package main

import (
	"strings"
	"unicode"
)

type TokenType string

const (
	TokenIllegal TokenType = "ILL"
	TokenEOF     TokenType = "EOF"
	TokenNewline TokenType = "NL"

	TokenIdentifier TokenType = "ID"
	TokenNumber     TokenType = "NUM"
	TokenString     TokenType = "STR"

	TokenAssign TokenType = "="
	TokenComma  TokenType = ","

	TokenPlus  TokenType = "+"
	TokenMinus TokenType = "-"
	TokenStar  TokenType = "*"
	TokenSlash TokenType = "/"

	TokenLParen TokenType = "("
	TokenRParen TokenType = ")"
	TokenLBrace TokenType = "{"
	TokenRBrace TokenType = "}"

	TokenEqual        TokenType = "=="
	TokenNotEqual     TokenType = "!="
	TokenLess         TokenType = "<"
	TokenLessEqual    TokenType = "<="
	TokenGreater      TokenType = ">"
	TokenGreaterEqual TokenType = ">="
)

type Token struct {
	Type        TokenType
	Literal     string
	Line        int
	Column      int
	IsImmutable bool // true for all-uppercase identifiers such as PI or MAX_SIZE
}

type Lexer struct {
	input []rune
	pos   int
	ch    rune

	// line and col are the position of the current character, using 1-based columns.
	line int
	col  int

	// nextLine and nextCol are the position that will be assigned to the next character read.
	nextLine int
	nextCol  int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:    []rune(input),
		nextLine: 1,
		nextCol:  1,
	}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() Token {
	if tok, ok := l.skipIgnored(); ok {
		return tok
	}

	line, col := l.line, l.col

	switch l.ch {
	case 0:
		return Token{Type: TokenEOF, Literal: "", Line: line, Column: col}

	case '\n':
		l.readChar()
		return Token{Type: TokenNewline, Literal: "\n", Line: line, Column: col}

	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: TokenEqual, Literal: "==", Line: line, Column: col}
		}
		l.readChar()
		return Token{Type: TokenAssign, Literal: "=", Line: line, Column: col}

	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: TokenNotEqual, Literal: "!=", Line: line, Column: col}
		}
		literal := string(l.ch)
		l.readChar()
		return Token{Type: TokenIllegal, Literal: literal, Line: line, Column: col}

	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: TokenLessEqual, Literal: "<=", Line: line, Column: col}
		}
		l.readChar()
		return Token{Type: TokenLess, Literal: "<", Line: line, Column: col}

	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: TokenGreaterEqual, Literal: ">=", Line: line, Column: col}
		}
		l.readChar()
		return Token{Type: TokenGreater, Literal: ">", Line: line, Column: col}

	case ',':
		l.readChar()
		return Token{Type: TokenComma, Literal: ",", Line: line, Column: col}

	case '+':
		l.readChar()
		return Token{Type: TokenPlus, Literal: "+", Line: line, Column: col}

	case '-':
		l.readChar()
		return Token{Type: TokenMinus, Literal: "-", Line: line, Column: col}

	case '*':
		l.readChar()
		return Token{Type: TokenStar, Literal: "*", Line: line, Column: col}

	case '/':
		l.readChar()
		return Token{Type: TokenSlash, Literal: "/", Line: line, Column: col}

	case '(':
		l.readChar()
		return Token{Type: TokenLParen, Literal: "(", Line: line, Column: col}

	case ')':
		l.readChar()
		return Token{Type: TokenRParen, Literal: ")", Line: line, Column: col}

	case '{':
		l.readChar()
		return Token{Type: TokenLBrace, Literal: "{", Line: line, Column: col}

	case '}':
		l.readChar()
		return Token{Type: TokenRBrace, Literal: "}", Line: line, Column: col}

	case '"':
		literal, ok := l.readString()
		if !ok {
			return Token{Type: TokenIllegal, Literal: literal, Line: line, Column: col}
		}

		return Token{Type: TokenString, Literal: literal, Line: line, Column: col}
	}

	if isIdentStart(l.ch) {
		literal := l.readIdentifier()
		return Token{
			Type:        TokenIdentifier,
			Literal:     literal,
			Line:        line,
			Column:      col,
			IsImmutable: isImmutableIdentifier(literal),
		}
	}

	if unicode.IsDigit(l.ch) || (l.ch == '.' && unicode.IsDigit(l.peekChar())) {
		literal, ok := l.readNumber()
		if !ok {
			return Token{Type: TokenIllegal, Literal: literal, Line: line, Column: col}
		}
		return Token{Type: TokenNumber, Literal: literal, Line: line, Column: col}
	}

	literal := string(l.ch)
	l.readChar()
	return Token{Type: TokenIllegal, Literal: literal, Line: line, Column: col}
}

func (l *Lexer) skipIgnored() (Token, bool) {
	for {
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
			l.readChar()
		}

		if l.ch != '|' {
			return Token{}, false
		}

		line, col := l.line, l.col

		if l.peekChar() == '|' {
			// Line comment: || comment until newline
			l.readChar()
			l.readChar()
			for l.ch != 0 && l.ch != '\n' {
				l.readChar()
			}
			// Do not consume the newline. It is still a statement separator.
			continue
		}

		// Inline comment: | comment until next |
		l.readChar()
		for l.ch != 0 && l.ch != '|' {
			l.readChar()
		}

		if l.ch == 0 {
			return Token{Type: TokenIllegal, Literal: "unterminated inline comment", Line: line, Column: col}, true
		}

		// Consume closing | and continue scanning.
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.pos - 1
	for isIdentPart(l.ch) {
		l.readChar()
	}
	return string(l.input[start:l.literalEnd()])
}

func (l *Lexer) readNumber() (string, bool) {
	start := l.pos - 1

	for unicode.IsDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' {
		l.readChar()
		for unicode.IsDigit(l.ch) {
			l.readChar()
		}
	}

	if l.ch == 'e' || l.ch == 'E' {
		l.readChar()

		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}

		if !unicode.IsDigit(l.ch) {
			return string(l.input[start:l.literalEnd()]), false
		}

		for unicode.IsDigit(l.ch) {
			l.readChar()
		}
	}

	return string(l.input[start:l.literalEnd()]), true
}

func (l *Lexer) literalEnd() int {
	if l.ch == 0 {
		return l.pos
	}
	return l.pos - 1
}

func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0
		l.line = l.nextLine
		l.col = l.nextCol
		return
	}

	l.ch = l.input[l.pos]
	l.line = l.nextLine
	l.col = l.nextCol
	l.pos++

	if l.ch == '\n' {
		l.nextLine++
		l.nextCol = 1
	} else {
		l.nextCol++
	}
}

func (l *Lexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func isIdentStart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isIdentPart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

func isImmutableIdentifier(name string) bool {
	hasUpper := false

	for _, ch := range name {
		if unicode.IsLower(ch) {
			return false
		}
		if unicode.IsUpper(ch) {
			hasUpper = true
		}
	}

	return hasUpper
}

func (l *Lexer) readString() (string, bool) {
	var out strings.Builder

	// Consume opening quote.
	l.readChar()

	for l.ch != 0 && l.ch != '"' {
		if l.ch == '\n' {
			return "unterminated string", false
		}

		if l.ch == '\\' {
			l.readChar()

			switch l.ch {
			case 'n':
				out.WriteRune('\n')
			case 't':
				out.WriteRune('\t')
			case '"':
				out.WriteRune('"')
			case '\\':
				out.WriteRune('\\')
			case 0:
				return "unterminated string", false
			default:
				return "invalid escape sequence \\" + string(l.ch), false
			}

			l.readChar()
			continue
		}

		out.WriteRune(l.ch)
		l.readChar()
	}

	if l.ch != '"' {
		return "unterminated string", false
	}

	// Consume closing quote.
	l.readChar()

	return out.String(), true
}
