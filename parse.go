package gol

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrNoMoreTokens = errors.New("No More Tokens")

type Parser struct {
	tokens chan Token
	peek   *Token
}

func NewParser(tokens chan Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() (Node, error) {
	// We evaluate a program as an implicit progn over the whole text
	// TODO - bad idea, instead parse sexp by sexp with 'ParseOne' and a loop in 'ParseAll'
	progn := NodeList{}
	progn.children = append(progn.children, NodeIdentifier{
		nodeAtom{
			tok: Token{
				tokIdentifier,
				"progn",
				Position{},
			},
		},
	})
	for {
		tree, err := p.parseSexp()
		if err != nil {
			if err == ErrNoMoreTokens {
				break
			}
			return nil, err
		}
		progn.children = append(progn.children, tree)
	}
	return Transform(progn)
}

func (p *Parser) peekToken() (Token, error) {
	if p.peek == nil {
		tok, ok := <-p.tokens
		if !ok {
			return TokBug, ErrNoMoreTokens
		}
		p.peek = &tok
	}
	return *p.peek, nil
}

func (p *Parser) stepToken() {
	p.peek = nil
}

type NodeBase struct {
}

type Node interface {
	String() string
	Pos() Position
	IsAtom() bool
}

type NodeList struct {
	NodeBase
	children []Node
}

func (nl NodeList) Pos() Position {
	if len(nl.children) == 0 {
		return Position{File: "<empty list>"}
	} else {
		return nl.children[0].Pos()
	}
}

func (nl NodeList) IsAtom() bool {
	return false
}

func (nl *NodeList) Add(node Node) {
	nl.children = append(nl.children, node)
}

func nodesToString(nodes []Node) string {
	s := []byte("(")
	for i, n := range nodes {
		if i != 0 {
			s = append(s, ' ')
		}
		s = append(s, n.String()...)
	}
	s = append(s, ')')
	return string(s)
}

func (nl NodeList) String() string {
	return nodesToString(nl.children)
}

type nodeAtom struct {
	NodeBase
	tok Token
}

func (na nodeAtom) Pos() Position {
	return na.tok.Pos
}

func (na nodeAtom) IsAtom() bool {
	return true
}

type NodeInt struct {
	nodeAtom
	value int64
}

func (nn NodeInt) String() string {
	return fmt.Sprintf("%d", nn.value)
}
func (nn NodeInt) Value() int64 {
	return nn.value
}

type NodeIdentifier struct {
	nodeAtom
}
type NodeSymbol struct {
	nodeAtom
}
type NodeString struct {
	nodeAtom
}
type NodeBool struct {
	nodeAtom
}

func (nb NodeBool) IsTrue() bool {
	return nb.String() == "#t"
}

func (ns NodeString) String() string {
	// Unescape
	value := make([]rune, 0, len(ns.tok.Value))
	escaped := false
RUNE:
	for _, r := range ns.tok.Value {
		if r == '\\' {
			if !escaped {
				escaped = true
				continue RUNE
			} else {
				escaped = false
				// fall through
			}
		}
		if escaped {
			switch r {
			case 'n':
				value = append(value, '\n')
			default:
				value = append(value, r)
			}
		} else {
			value = append(value, r)
		}
	}
	return string(value)
}

func (na nodeAtom) String() string {
	return na.tok.Value
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
	tok, err := p.peekToken()
	if err != nil {
		return nil, err
	}
	if tok.Type == tokLParen {
		return p.parseList()
	} else {
		return p.parseAtom()
	}
}

func (p *Parser) parseAtom() (Node, error) {
	tok, err := p.peekToken()
	if err != nil {
		return nil, err
	}
	p.stepToken()
	switch tok.Type {
	case tokLParen:
		return nil, p.Error("Found L Paren, expected atom")
	case tokRParen:
		return nil, p.Error("Found R Paren, expected atom")
	case tokIdentifier:
		return NodeIdentifier{nodeAtom{tok: tok}}, nil
	case tokSymbol:
		return NodeSymbol{nodeAtom{tok: tok}}, nil
	case tokString:
		return NodeString{nodeAtom{tok: tok}}, nil
	case tokBool:
		if tok.Value != "#t" && tok.Value != "#f" {
			return nil, fmt.Errorf("Bad boolean value [%s]", tok.Value)
		}
		return NodeBool{nodeAtom{tok: tok}}, nil
	case tokInt:
		// Special case '+' and '-'
		if tok.Value == "+" || tok.Value == "-" {
			return NodeIdentifier{nodeAtom{tok: tok}}, nil
		}

		v, err := strconv.ParseInt(tok.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Can't parse [%s] as integer: %s", tok.Value, err)
		}
		return NodeInt{value: v}, nil
	default:
		panic("Unknown atom type")
	}
}

func (p *Parser) parseList() (Node, error) {
	tok, err := p.peekToken()
	if err != nil {
		return nil, err
	}
	if tok.Type != tokLParen {
		return nil, p.Error("Parser logic error - missing L paren at start of list")
	}
	p.stepToken()
	NodeList := NodeList{}
	for {
		t, err := p.peekToken()
		if err != nil {
			return nil, err
		}
		if t.Type == tokRParen {
			p.stepToken()
			return NodeList, nil
		}
		node, err := p.parseSexp()
		if err != nil {
			return nil, err
		}
		NodeList.Add(node)
	}
}
