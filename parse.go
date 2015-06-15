package gol

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
