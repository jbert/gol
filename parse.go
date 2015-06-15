package gol

import (
	"fmt"
	"strconv"
)

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
	return p.parseSexp()
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

type Node interface {
	String() string
}

type NodeList struct {
	children []Node
}

func (nl *NodeList) Add(node Node) {
	nl.children = append(nl.children, node)
}
func (nl NodeList) String() string {
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

type nodeAtom struct {
	tok Token
}
type NodeNum struct {
	value float64
}

func (nn NodeNum) String() string {
	return fmt.Sprintf("%f", nn.value)
}
func (nn NodeNum) Value() float64 {
	return nn.value
}

type NodeIdentifier struct {
	nodeAtom
}
type NodeSymbol struct {
	nodeAtom
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
	case tokIdentifier:
		return NodeIdentifier{nodeAtom{tok: tok}}, nil
	case tokSymbol:
		return NodeSymbol{nodeAtom{tok: tok}}, nil
	case tokNum:
		v, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("Can't parse [%s] as float: %s", tok.Value, err)
		}
		return NodeNum{value: v}, nil
	default:
		panic("Unknown atom type")
	}
}

func (p *Parser) parseList() (Node, error) {
	if p.peekToken().Type != tokLParen {
		return nil, p.Error("Parser logic error - missing L paren at start of list")
	}
	p.stepToken()
	NodeList := NodeList{}
	for {
		t := p.peekToken()
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
