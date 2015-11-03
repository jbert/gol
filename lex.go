package gol

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

type TokType int

const (
	tokBug TokType = iota // Make the zero value something which causes an error
	tokLParen
	tokRParen
	tokInt
	tokIdentifier
	tokSymbol
	tokString
	tokBool
	tokQuote
	tokBackQuote
	tokComma
)

func (tt TokType) String() string {
	switch tt {
	case tokQuote:
		return "tokQuote"
	case tokBackQuote:
		return "tokBackQuote"
	case tokComma:
		return "tokComma"
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
	case tokBool:
		return "tokBool"
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
	Type  TokType
	Value string
	Pos   Position
}

var TokBug = Token{tokBug, "error", Position{"error", 0, 0}}

func (t Token) String() string {
	return fmt.Sprintf("%s [%s]", t.Type, t.Value)
}

type Lexer struct {
	Tokens chan Token
	fname  string
	// Source of new runes
	r       io.Reader
	seenEOF bool

	// Token we are assembling
	buf []byte
	// index of next rune in unlexed data
	pos int

	line int
	col  int
}

func NewLexer(fname string, r io.Reader) *Lexer {
	return &Lexer{
		Tokens: make(chan Token),
		r:      r,
		line:   1, // People count lines from 1! Who know.
		fname:  fname,
	}
}

func (l *Lexer) Run() error {
	for !l.isEOF() {
		l.skipWhitespace()
		r := l.peekNextRune()
		switch {
		case r == '\'':
			l.stepRune()
			l.emit(tokQuote)
		case r == '`':
			l.stepRune()
			l.emit(tokBackQuote)
		case r == ',':
			l.stepRune()
			l.emit(tokComma)
		case r == utf8.RuneError:
			// EOF case
			break
		case r == '#':
			l.stepRune()
			l.stepRune()
			l.emit(tokBool)
		case r == '"':
			l.skipRune()
			escaped := true
			l.emitMatching(tokString, func(r rune) bool {
				if r == '\\' {
					escaped = !escaped
					return true
				} else if escaped {
					escaped = false
					return true
				} else {
					escaped = false
					return r != '"'
				}
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
			return posErrorf(l.currentPosition(), "Internal error: unrecognised rune [%c]", r)
		}
	}
	close(l.Tokens)
	return nil
}

func (l *Lexer) isEOF() bool {
	return l.seenEOF && l.pos == len(l.buf)
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
			l.discardToPos()
		} else {
			break PEEKING
		}

	}
	return l.isEOF()
}

func (l *Lexer) skipRune() {
	l.fetchCheck()
	l.stepRune()
	l.discardToPos()
}

func (l *Lexer) peekNextRune() rune {
	l.fetchCheck()
	r, _ := utf8.DecodeRune(l.buf[l.pos:])
	return r
}

func (l *Lexer) stepRune() {
	l.fetchCheck()
	_, size := utf8.DecodeRune(l.buf[l.pos:])
	l.pos += size
	l.col += size
}

func (l *Lexer) discardToPos() {
	l.buf = l.buf[l.pos:]
	l.pos = 0
}

func (l *Lexer) fetchCheck() {
	// Grab a bit more if we can
	bufSize := 1024
	buf := make([]byte, bufSize)
	n, err := l.r.Read(buf)
	if err == io.EOF {
		l.seenEOF = true
	}
	if n > 0 {
		l.buf = append(l.buf, buf[:n]...)
	}
}

func (l *Lexer) emitMatching(tokType TokType, f func(r rune) bool) {
	for !l.isEOF() && f(l.peekNextRune()) {
		l.stepRune()
	}
	l.emit(tokType)
}

func (l Lexer) currentPosition() Position {
	return Position{File: l.fname, Line: l.line, Column: l.col}
}

func (l *Lexer) emit(tokType TokType) {
	tok := Token{Pos: l.currentPosition(), Type: tokType, Value: string(l.buf[:l.pos])}
	l.Tokens <- tok
	l.discardToPos()
}
