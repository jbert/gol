package main

import (
	"fmt"
	"os"
	"unicode"
	"unicode/utf8"
)

func main() {
	fname := "tt.gol"
	s := `
(func (inc (x))
	(+ 1 x))
`

	l := NewLexer(fname, s)
	var lexErr error
	go func() {
		lexErr = l.Run()
	}()

	p := NewParser(l.Tokens)
	a, parseErr := p.Parse()
	if parseErr != nil {
		fmt.Printf("Error parsing: %s\n", parseErr)
		os.Exit(-1)
	}

	fmt.Println(a)

	if lexErr != nil {
		fmt.Printf("Error lexing: %s\n", lexErr)
		os.Exit(-1)
	}
}

// ==================================================
// Parser
// ==================================================

type Parser struct {
	tokens chan Token
	peek   *Token
}

func NewParser(tokens chan Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) peekToken() Token {
	if p.peek == nil {
		tok := <-p.tokens
		p.peek = &tok
	}
	return *p.peek
}

func (p *Parser) stepToken() {
	p.peek = nil
}

type ASTType int

const (
	astList ASTType = iota
	astSymbol
	astNum
)

type Node interface {
	Type() ASTType
	String() string
}

type nodeList struct {
	children []Node
}

func (nl *nodeList) Add(node Node) {
	nl.children = append(nl.children, node)
}
func (nl nodeList) Type() ASTType {
	return astList
}
func (nl nodeList) String() string {
	s := []byte("(")
	for i, child := range nl.children {
		if i != 0 {
			s = append(s, ' ')
		}
		s = append(s, child.String()...)
	}
	s = append(s, ')')
	return string(s)
}

// TODO: break out atom types
type nodeAtom struct {
	tok Token
}

func (na nodeAtom) Type() ASTType {
	switch na.tok.Type {
	case tokNum:
		return astNum
	case tokSymbol:
		return astSymbol
	default:
		panic("Unknown token type in nodeAtom")
	}
}
func (na nodeAtom) String() string {
	return na.tok.Value
}

func (p *Parser) Parse() (Node, error) {
	return p.parseSexp()
}

type ParserError struct {
	reason string
}

func (pe ParserError) Error() string {
	return pe.reason
}

func (p *Parser) Error(reason string) error {
	// TODO - add file, line and column
	return ParserError{reason: reason}
}

func (p *Parser) parseSexp() (Node, error) {
	if p.peekToken().Type == tokLParen {
		return p.parseList()
	} else {
		return p.parseAtom()
	}
}

func (p *Parser) parseAtom() (Node, error) {
	tok := p.peekToken()
	p.stepToken()
	switch tok.Type {
	case tokLParen:
		return nil, p.Error("Found L Paren, expected atom")
	case tokRParen:
		return nil, p.Error("Found R Paren, expected atom")
	case tokSymbol:
		fallthrough
	case tokNum:
		return nodeAtom{tok: tok}, nil
	default:
		panic("Unknown atom type")
	}
}

func (p *Parser) parseList() (Node, error) {
	if p.peekToken().Type != tokLParen {
		return nil, p.Error("Parser logic error - missing L paren at start of list")
	}
	p.stepToken()
	nodeList := nodeList{}
	for {
		t := p.peekToken()
		if t.Type == tokRParen {
			p.stepToken()
			return nodeList, nil
		}
		node, err := p.parseSexp()
		if err != nil {
			return nil, err
		}
		nodeList.Add(node)
	}
}

// ==================================================
// Lexer
// ==================================================

type TokType int

const (
	tokLParen TokType = iota
	tokRParen
	tokNum
	tokSymbol
)

func (tt TokType) String() string {
	switch tt {
	case tokLParen:
		return "tokLParen"
	case tokRParen:
		return "tokRParen"
	case tokNum:
		return "tokNum"
	case tokSymbol:
		return "tokSymbol"
	default:
		return "<unknown>"
	}
}

type Token struct {
	Type  TokType
	Value string
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
		case r == '(':
			l.stepRune()
			l.emit(tokLParen)
		case r == ')':
			l.stepRune()
			l.emit(tokRParen)
		case unicode.IsDigit(r):
			l.emitMatching(tokNum, unicode.IsDigit)
		case unicode.IsSpace(r):
			l.stepRune()
		case !unicode.IsSpace(r):
			l.emitMatching(tokSymbol, func(r rune) bool {
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
	for !l.isEOF() && unicode.IsSpace(l.peekNextRune()) {
		l.stepRune()
	}
	l.start = l.pos
	return l.isEOF()
}

func (l *Lexer) peekNextRune() rune {
	r, _ := utf8.DecodeRuneInString(l.s[l.pos:])
	return r
}

func (l *Lexer) stepRune() {
	_, size := utf8.DecodeRuneInString(l.s[l.pos:])
	l.pos += size
}

func (l *Lexer) emitMatching(tokType TokType, f func(r rune) bool) {
	for !l.isEOF() && f(l.peekNextRune()) {
		l.stepRune()
	}
	l.emit(tokType)
}

func (l *Lexer) emit(tokType TokType) {
	l.Tokens <- Token{Type: tokType, Value: l.s[l.start:l.pos]}
	l.start = l.pos
}
