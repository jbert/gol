package main

import (
	"unicode"
	"unicode/utf8"
)

func main() {
	s := "(1 2)"
	l := NewLexer()
	go l.Lex(s)
	for item := range l.Items {
		println(item)
	}
}

type Lexer struct {
	s         string
	itemStart int
	pos       int
	Items     chan Item
}

type IType int

type Item struct {
	typ IType
	val string
}

const (
	itemLParen IType = iota
	itemRParen
	itemNum
)

func NewLexer() *Lexer {
	return &Lexer{}
}

type stateFn func(l *Lexer) stateFn

func (l *Lexer) Lex(in string) {
	l.s = in

	for state := lexSexp; state != nil; state = state(l) {
	}
}

func (l *Lexer) isEof() bool {
	return l.pos == len(l.s)
}

func (l *Lexer) skipMatching(f func(r rune) bool) bool {
	for !l.isEof() && f(l.peekNextRune()) {
		l.stepRune()
	}
	l.itemStart = l.pos
	return l.isEof()
}

func (l *Lexer) skipWhitespace() bool {
	return l.skipMatching(unicode.IsSpace)
}

func (l *Lexer) peekNextRune() rune {
	r, _ := utf8.DecodeRuneInString(l.s[l.pos:])
	return r
}

func (l *Lexer) stepRune() {
	_, size := utf8.DecodeRuneInString(l.s[l.pos:])
	l.pos += size
}

func lexSexp(l *Lexer) stateFn {
	if l.skipWhitespace() {
		return nil
	}
	if l.peekNextRune() == '(' {
		return lexList
	} else {
		return lexAtom
	}
}

func lexAtom(l *Lexer) stateFn {
	return lexNum
}

func lexNum(l *Lexer) stateFn {
	return lexInt
}

func lexInt(l *Lexer) stateFn {
	emitMatching(itemNum, unicode.IsDigit)
	return lexSexp
}

func (l *Lexer) emitMatching(itemType IType, f func(r rune) bool) bool {
	for !l.isEof() && f(l.peekNextRune()) {
		l.stepRune()
	}
	return l.emit(itemType)
}

func (l *Lexer) emit(itemType IType) bool {
	l.items <- Item{itemType: itemType, val: l.s[l.itemStart:l.pos]}
	l.itemStart = l.pos
}

func lexList(l *Lexer) stateFn {
	if !l.peekNextRune() == '(' {
		panic("Logic error - lexing list with no paren start")
	}
	l.stepRune()
	l.emit(itemLParen)
	for !l.eof() && !l.peekNextRune() == ')' {
	}
}
