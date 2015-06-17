package gol

import "fmt"

type NodeLet struct {
	NodeList
	Bindings map[string]Node
	Body     Node
}

type NodeError struct {
	msg          string
	originalNode Node
}

func (ne *NodeError) Error() string {
	pos := ne.originalNode.Pos()
	return fmt.Sprintf("%s: %s line %d:%d", ne.msg, pos.File, pos.Line, pos.Column)
}

func Decorate(node Node) (Node, error) {
	switch n := node.(type) {
	case NodeList:
		return decorateList(n)
	default:
		return node, nil
	}
}

func decorateList(n NodeList) (Node, error) {
	if len(n.children) == 0 {
		return n, nil
	}

	first := n.children[0]
	id, ok := first.(NodeIdentifier)
	if !ok {
		return n, nil
	}
	switch id.String() {
	case "let":
		return decorateLet(n)
	default:
		return n, nil
	}
}

func decorateLet(n NodeList) (Node, error) {
	if len(n.children) < 3 {
		return nil, fmt.Errorf("Bad let expression - missing bindings or body")
	}
	nLet := NodeLet{NodeList: n, Bindings: make(map[string]Node)}
	bindings, ok := n.children[1].(NodeList)
	if !ok {
		return nil, fmt.Errorf("Bad let expression - bindings must be a list")
	}
	for _, pairNode := range bindings.children {
		pair, ok := pairNode.(NodeList)
		if !ok {
			return nil, fmt.Errorf("Bad let expression - bindings must be pairs")
		}
		if len(pair.children) != 2 {
			return nil, fmt.Errorf("Bad let expression - bindings must be pairs")
		}
		id, ok := pair.children[0].(NodeIdentifier)
		if !ok {
			return nil, fmt.Errorf("Bad let expression - invalid identifier")
		}
		nLet.Bindings[id.String()] = pair.children[1]
	}

	// TODO: support implicit progn
	nLet.Body = n.children[2]
	return nLet, nil
}
