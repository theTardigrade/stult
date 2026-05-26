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

	TokenPlusAssign  TokenType = ":+"
	TokenMinusAssign TokenType = ":-"
	TokenComma       TokenType = ","
	TokenColon       TokenType = ":"
	TokenAt          TokenType = "@"
	TokenCaret       TokenType = "^"

	TokenPlus  TokenType = "+"
	TokenMinus TokenType = "-"
	TokenStar  TokenType = "*"
	TokenSlash TokenType = "/"

	TokenAnd TokenType = "&"
	TokenOr  TokenType = "|"

	TokenLParen TokenType = "("
	TokenRParen TokenType = ")"

	TokenLBracket TokenType = "["
	TokenRBracket TokenType = "]"

	TokenLBrace TokenType = "{"
	TokenRBrace TokenType = "}"

	TokenEqual        TokenType = "="
	TokenNotEqual     TokenType = "!"
	TokenLess         TokenType = "<"
	TokenLessEqual    TokenType = "<="
	TokenGreater      TokenType = ">"
	TokenGreaterEqual TokenType = ">="
)

type Token struct {
	Type          TokenType
	Literal       string
	StartOfLine   int
	StartOfColumn int
	EndOfLine     int
	EndOfColumn   int
	IsImmutable   bool // true for all-uppercase identifiers such as PI or MAX_SIZE
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
		return Token{
			Type:          TokenEOF,
			Literal:       "",
			StartOfLine:   line,
			StartOfColumn: col,
			EndOfLine:     line,
			EndOfColumn:   col,
		}

	case '\n':
		l.readChar()
		return l.makeToken(TokenNewline, "\n", line, col)

	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return l.makeToken(TokenIllegal, "==", line, col)
		}

		l.readChar()
		return l.makeToken(TokenEqual, "=", line, col)

	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return l.makeToken(TokenIllegal, "!=", line, col)
		}

		l.readChar()
		return l.makeToken(TokenNotEqual, "!", line, col)

	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return l.makeToken(TokenLessEqual, "<=", line, col)
		}

		l.readChar()
		return l.makeToken(TokenLess, "<", line, col)

	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return l.makeToken(TokenGreaterEqual, ">=", line, col)
		}

		l.readChar()
		return l.makeToken(TokenGreater, ">", line, col)

	case ',':
		l.readChar()
		return l.makeToken(TokenComma, ",", line, col)

	case ':':
		if l.peekChar() == '+' {
			l.readChar()
			l.readChar()
			return l.makeToken(TokenPlusAssign, ":+", line, col)
		}

		if l.peekChar() == '-' {
			l.readChar()
			l.readChar()
			return l.makeToken(TokenMinusAssign, ":-", line, col)
		}

		l.readChar()
		return l.makeToken(TokenColon, ":", line, col)

	case '@':
		l.readChar()
		return l.makeToken(TokenAt, "@", line, col)

	case '^':
		l.readChar()
		return l.makeToken(TokenCaret, "^", line, col)

	case '+':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return l.makeToken(TokenIllegal, "+=", line, col)
		}

		l.readChar()
		return l.makeToken(TokenPlus, "+", line, col)

	case '-':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return l.makeToken(TokenIllegal, "-=", line, col)
		}

		l.readChar()
		return l.makeToken(TokenMinus, "-", line, col)

	case '*':
		l.readChar()
		return l.makeToken(TokenStar, "*", line, col)

	case '/':
		l.readChar()
		return l.makeToken(TokenSlash, "/", line, col)

	case '&':
		l.readChar()
		return l.makeToken(TokenAnd, "&", line, col)

	case '|':
		l.readChar()
		return l.makeToken(TokenOr, "|", line, col)

	case '(':
		l.readChar()
		return l.makeToken(TokenLParen, "(", line, col)

	case ')':
		l.readChar()
		return l.makeToken(TokenRParen, ")", line, col)

	case '[':
		l.readChar()
		return l.makeToken(TokenLBracket, "[", line, col)

	case ']':
		l.readChar()
		return l.makeToken(TokenRBracket, "]", line, col)

	case '{':
		l.readChar()
		return l.makeToken(TokenLBrace, "{", line, col)

	case '}':
		l.readChar()
		return l.makeToken(TokenRBrace, "}", line, col)

	case '"':
		literal, ok := l.readString()
		if !ok {
			return l.makeToken(TokenIllegal, literal, line, col)
		}

		return l.makeToken(TokenString, literal, line, col)
	}

	if isIdentStart(l.ch) {
		literal := l.readIdentifier()
		return l.makeIdentifierToken(literal, line, col)
	}

	if unicode.IsDigit(l.ch) || (l.ch == '.' && unicode.IsDigit(l.peekChar())) {
		literal, ok := l.readNumber()
		if !ok {
			return l.makeToken(TokenIllegal, literal, line, col)
		}

		return l.makeToken(TokenNumber, literal, line, col)
	}

	literal := string(l.ch)
	l.readChar()
	return l.makeToken(TokenIllegal, literal, line, col)
}

func (l *Lexer) makeToken(tokenType TokenType, literal string, line int, column int) Token {
	return Token{
		Type:          tokenType,
		Literal:       literal,
		StartOfLine:   line,
		StartOfColumn: column,
		EndOfLine:     l.line,
		EndOfColumn:   l.col,
	}
}

func (l *Lexer) makeIdentifierToken(literal string, line int, column int) Token {
	return Token{
		Type:          TokenIdentifier,
		Literal:       literal,
		StartOfLine:   line,
		StartOfColumn: column,
		EndOfLine:     l.line,
		EndOfColumn:   l.col,
		IsImmutable:   isImmutableIdentifier(literal),
	}
}

func (l *Lexer) skipIgnored() (Token, bool) {
	for {
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
			l.readChar()
		}

		if l.ch != '#' {
			return Token{}, false
		}

		if l.countConsecutiveHashes() >= 3 {
			line, col := l.line, l.col
			l.readHashRun()
			return l.makeToken(TokenIllegal, "three or more consecutive '#' symbols are not allowed", line, col), true
		}

		if l.peekChar() == '#' {
			line, col := l.line, l.col

			// Bounded comment: ## comment until next ##
			l.readChar()
			l.readChar()

			closed := false

			for l.ch != 0 {
				if l.ch == '#' {
					hashCount := l.countConsecutiveHashes()

					if hashCount >= 3 {
						line, col := l.line, l.col
						l.readHashRun()
						return l.makeToken(TokenIllegal, "three or more consecutive '#' symbols are not allowed", line, col), true
					}

					if hashCount == 2 {
						l.readChar()
						l.readChar()
						closed = true
						break
					}
				}

				l.readChar()
			}

			if !closed {
				return l.makeToken(TokenIllegal, "unterminated bounded comment", line, col), true
			}

			continue
		}

		// Line comment: # comment until newline
		for l.ch != 0 && l.ch != '\n' {
			if l.ch == '#' && l.countConsecutiveHashes() >= 3 {
				line, col := l.line, l.col
				l.readHashRun()
				return l.makeToken(TokenIllegal, "three or more consecutive '#' symbols are not allowed", line, col), true
			}

			l.readChar()
		}

		// Do not consume the newline. It is still a statement separator.
	}
}

func (l *Lexer) countConsecutiveHashes() int {
	count := 0

	for offset := 0; l.peekAhead(offset) == '#'; offset++ {
		count++
	}

	return count
}

func (l *Lexer) readHashRun() {
	for l.ch == '#' {
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
	return l.peekAhead(1)
}

func (l *Lexer) peekAhead(offset int) rune {
	index := l.pos - 1 + offset
	if index < 0 || index >= len(l.input) {
		return 0
	}

	return l.input[index]
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
