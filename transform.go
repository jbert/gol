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

func Transform(node Node) (Node, error) {
	switch n := node.(type) {
	case NodeList:
		return transformList(n)
	default:
		return node, nil
	}
}

func transformNodes(ns []Node) ([]Node, error) {
	newNs := make([]Node, 0)
	for _, n := range ns {
		newNode, err := Transform(n)
		if err != nil {
			return nil, err
		}
		newNs = append(newNs, newNode)
	}
	return newNs, nil
}

func transformList(n NodeList) (Node, error) {
	if len(n.children) == 0 {
		return n, nil
	}

	first := n.children[0]
	id, ok := first.(NodeIdentifier)
	if ok {
		switch id.String() {
		case "let":
			return transformLet(n)
		case "progn":
			return transformProgn(n)
		case "lambda":
			return transformLambda(n)
		}
	}
	children, err := transformNodes(n.children)
	if err != nil {
		return nil, err
	}
	return NodeList{NodeBase: n.NodeBase, children: children}, nil
}

type NodeProgn struct {
	NodeList
}

func transformProgn(n NodeList) (Node, error) {
	n.children = n.children[1:]
	return NodeProgn{n}, nil
}

func transformLet(n NodeList) (Node, error) {
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
		var err error
		nLet.Bindings[id.String()], err = Transform(pair.children[1])
		if err != nil {
			return nil, err
		}
	}

	children, err := transformNodes(n.children[2:])
	if err != nil {
		return nil, err
	}
	nLet.Body = NodeProgn{NodeList{children: children}}
	return nLet, nil
}

func transformLambda(n NodeList) (Node, error) {
	if len(n.children) < 3 {
		return nil, fmt.Errorf("Bad lambda expression - missing args or body")
	}
	args, ok := n.children[1].(NodeList)
	if !ok {
		return nil, fmt.Errorf("Bad lambda expression - args must be a list")
	}
	for _, argNode := range args.children {
		_, ok := argNode.(NodeIdentifier)
		if !ok {
			return nil, fmt.Errorf("Bad lambda expression - arg must be identifier")
		}
	}
	nLambda := NodeLambda{NodeList: n, Args: args.children}

	children, err := transformNodes(n.children[2:])
	if err != nil {
		return nil, err
	}
	nLambda.Body = NodeProgn{NodeList{children: children}}
	return nLambda, nil
}
