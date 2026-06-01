package main

type lexerCheckpoint struct {
	pos      int
	ch       rune
	line     int
	col      int
	nextLine int
	nextCol  int
}

func (l *Lexer) checkpoint() lexerCheckpoint {
	return lexerCheckpoint{
		pos:      l.pos,
		ch:       l.ch,
		line:     l.line,
		col:      l.col,
		nextLine: l.nextLine,
		nextCol:  l.nextCol,
	}
}

func (l *Lexer) restore(checkpoint lexerCheckpoint) {
	l.pos = checkpoint.pos
	l.ch = checkpoint.ch
	l.line = checkpoint.line
	l.col = checkpoint.col
	l.nextLine = checkpoint.nextLine
	l.nextCol = checkpoint.nextCol
}
