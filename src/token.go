package main

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
	TokenStarAssign  TokenType = ":*"
	TokenSlashAssign TokenType = ":/"

	TokenComma    TokenType = ","
	TokenColon    TokenType = ":"
	TokenDot      TokenType = "."
	TokenQuestion TokenType = "?"
	TokenAt       TokenType = "@"
	TokenCaret    TokenType = "^"

	TokenRangeInclusive TokenType = ".."
	TokenRangeExclusive TokenType = "..."

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
