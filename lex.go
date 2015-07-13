package gol

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type TokType int

const (
	tokLParen TokType = iota
	tokRParen
	tokInt
	tokIdentifier
	tokSymbol
	tokString
)

func (tt TokType) String() string {
	switch tt {
	case tokLParen:
		return "tokLParen"
	case tokRParen:
		return "tokRParen"
	case tokInt:
		return "tokInt"
	case tokIdentifier:
		return "tokIdentifier"
	case tokSymbol:
		return "tokSymbol"
	case tokString:
		return "tokString"
	default:
		return "<unknown>"
	}
}

type Position struct {
	File   string
	Line   int
	Column int
}

type Token struct {
	Type     TokType
	Value    string
	Position Position
}

func (t Token) String() string {
	return fmt.Sprintf("%s [%s]", t.Type, t.Value)
}

type Lexer struct {
	Tokens chan Token
	fname  string
	s      string
	pos    int
	start  int
	line   int
	col    int
}

func NewLexer(fname string, data string) *Lexer {
	return &Lexer{
		Tokens: make(chan Token),
		s:      data,
		fname:  fname,
	}
}

func (l *Lexer) Run() error {
	for !l.isEOF() {
		l.skipWhitespace()
		r := l.peekNextRune()
		switch {
		case r == '"':
			l.skipRune()
			l.emitMatching(tokString, func(r rune) bool {
				return r != '"'
			})
			l.skipRune()
		case r == '(':
			l.stepRune()
			l.emit(tokLParen)
		case r == ')':
			l.stepRune()
			l.emit(tokRParen)
		case r == '+' || r == '-' || unicode.IsDigit(r):
			l.stepRune() // Allow the leading sign
			l.emitMatching(tokInt, unicode.IsDigit)
		case unicode.IsSpace(r):
			l.stepRune()
		case r == '\'':
			l.emitMatching(tokSymbol, func(r rune) bool {
				return !unicode.IsSpace(r) && r != '(' && r != ')'
			})
		case !unicode.IsSpace(r):
			l.emitMatching(tokIdentifier, func(r rune) bool {
				return !unicode.IsSpace(r) && r != '(' && r != ')'
			})
		default:
			close(l.Tokens)
			return fmt.Errorf("Internal error: unrecognised rune [%c]", r)
		}
	}
	close(l.Tokens)
	return nil
}

func (l *Lexer) isEOF() bool {
	return l.pos == len(l.s)
}

func (l *Lexer) skipWhitespace() bool {
PEEKING:
	for !l.isEOF() {
		next := l.peekNextRune()
		if next == '\n' {
			l.line++
			l.col = 0
		}
		if unicode.IsSpace(next) {
			l.stepRune()
		} else {
			break PEEKING
		}

	}
	l.start = l.pos
	return l.isEOF()
}

func (l *Lexer) skipRune() {
	l.stepRune()
	l.start = l.pos
}

func (l *Lexer) peekNextRune() rune {
	r, _ := utf8.DecodeRuneInString(l.s[l.pos:])
	return r
}

func (l *Lexer) stepRune() {
	_, size := utf8.DecodeRuneInString(l.s[l.pos:])
	l.pos += size
	l.col += size
}

func (l *Lexer) emitMatching(tokType TokType, f func(r rune) bool) {
	for !l.isEOF() && f(l.peekNextRune()) {
		l.stepRune()
	}
	l.emit(tokType)
}

func (l *Lexer) emit(tokType TokType) {
	pos := Position{File: l.fname, Line: l.line, Column: l.col}
	l.Tokens <- Token{Position: pos, Type: tokType, Value: l.s[l.start:l.pos]}
	l.start = l.pos
}
