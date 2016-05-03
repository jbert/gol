package gol

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrNoMoreTokens = errors.New("No More Tokens")

type PosError struct {
	msg string
	pos Position
}

func (pe PosError) Error() string {
	return fmt.Sprintf("%s: %s line %d:%d", pe.msg, pe.pos.File, pe.pos.Line, pe.pos.Column)
}
func posErrorf(pos Position, f string, args ...interface{}) PosError {
	return PosError{pos: pos, msg: fmt.Sprintf(f, args...)}
}

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
	progn := NewNodeList()
	progn = progn.Cons(&NodeIdentifier{
		nodeAtom{
			tok: Token{
				tokIdentifier,
				"progn",
				Position{},
			},
		},
	})
	//	fmt.Printf("progn is: %s %T %d\n", progn, progn, progn.Len())
	for {
		tree, err := p.parseSexp()
		if err != nil {
			if err == ErrNoMoreTokens {
				break
			}
			return nil, err
		}
		progn = progn.Cons(tree)
	}
	progn = progn.ReverseCopy()
	//	fmt.Printf("progn is: %s %T %d\n", progn, progn, progn.Len())
	return progn, nil
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

type ParserError struct {
	reason string
}

func (pe ParserError) Error() string {
	return pe.reason
}

func (p *Parser) Error(tok Token, reason string) error {
	return posErrorf(tok.Pos, "%s", reason)
}

func (p *Parser) parseSexp() (Node, error) {
	tok, err := p.peekToken()
	if err != nil {
		return nil, err
	}
	switch tok.Type {
	case tokQuote:
		p.stepToken()
		arg, err := p.parseSexp()
		if err != nil {
			return nil, err
		}
		return &NodeQuote{Arg: arg, Quasi: false}, nil
	case tokBackQuote:
		p.stepToken()
		arg, err := p.parseSexp()
		if err != nil {
			return nil, err
		}
		return &NodeQuote{Arg: arg, Quasi: true}, nil
	case tokComma:
		p.stepToken()
		arg, err := p.parseSexp()
		if err != nil {
			return nil, err
		}
		return &NodeUnQuote{Arg: arg}, nil
	case tokLParen:
		return p.parseList()
	default:
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
		return nil, p.Error(tok, "Found L Paren, expected atom")
	case tokRParen:
		return nil, p.Error(tok, "Found R Paren, expected atom")
	case tokIdentifier:
		return &NodeIdentifier{nodeAtom{tok: tok}}, nil
	case tokSymbol:
		return &NodeSymbol{nodeAtom{tok: tok}}, nil
	case tokString:
		return &NodeString{nodeAtom{tok: tok}}, nil
	case tokBool:
		if tok.Value != "#t" && tok.Value != "#f" {
			return nil, posErrorf(tok.Pos, "Bad boolean value [%s]", tok.Value)
		}
		return &NodeBool{nodeAtom{tok: tok}}, nil
	case tokInt:
		// Special case '+' and '-'
		if tok.Value == "+" || tok.Value == "-" {
			return &NodeIdentifier{nodeAtom{tok: tok}}, nil
		}

		v, err := strconv.ParseInt(tok.Value, 10, 64)
		if err != nil {
			return nil, posErrorf(tok.Pos, "Can't parse [%s] as integer: %s", tok.Value, err)
		}
		return NewNodeInt(v), nil
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
		return nil, p.Error(tok, "Parser logic error - missing L paren at start of list")
	}
	p.stepToken()
	nodeList := NewNodeList()
	for {
		t, err := p.peekToken()
		if err != nil {
			return nil, err
		}
		if t.Type == tokRParen {
			p.stepToken()
			return nodeList, nil
		}
		node, err := p.parseSexp()
		if err != nil {
			return nil, err
		}
		nodeList = nodeList.Append(node)
	}
}
