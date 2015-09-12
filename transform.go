package gol

import "fmt"

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
		case "define":
			return transformDefine(n)
		case "let":
			return transformLet(n)
		case "progn":
			return transformProgn(n)
		case "lambda":
			return transformLambda(n)
		case "error":
			return transformError(n)
		case "if":
			return transformIf(n)
		}
	}
	children, err := transformNodes(n.children)
	if err != nil {
		return nil, err
	}
	return NodeList{NodeBase: n.NodeBase, children: children}, nil
}

type NodeIf struct {
	NodeList
	Condition Node
	TBranch   Node
	FBranch   Node
}

func transformIf(n NodeList) (Node, error) {
	if len(n.children) != 4 {
		return nil, fmt.Errorf("Bad if expression - missing test or t/f branch")
	}
	children, err := transformNodes(n.children[1:])
	if err != nil {
		return nil, err
	}
	return NodeIf{
		NodeList:  n,
		Condition: children[0],
		TBranch:   children[1],
		FBranch:   children[2],
	}, nil
}

type NodeError struct {
	Node
	msg string
}

func (ne NodeError) String() string {
	return ne.Error()
}
func (ne NodeError) Error() string {
	pos := ne.Pos()
	return fmt.Sprintf("%s: %s line %d:%d [%s]", ne.msg, pos.File, pos.Line, pos.Column, ne.Node)
}

func nodeErrorf(n Node, f string, args ...interface{}) NodeError {
	return NodeError{Node: n, msg: fmt.Sprintf(f, args...)}
}

func transformError(n NodeList) (Node, error) {
	if len(n.children) != 2 {
		return nil, fmt.Errorf("Bad error expression - exactly one string required")
	}
	return NodeError{n.children[1], n.children[1].String()}, nil
}

type NodeProgn struct {
	NodeList
}

func transformProgn(n NodeList) (Node, error) {
	children, err := transformNodes(n.children[1:])
	if err != nil {
		return nil, err
	}
	n.children = children
	return NodeProgn{n}, nil
}

type NodeDefine struct {
	NodeList
	Symbol Node
	Value  Node
}

func transformDefine(n NodeList) (Node, error) {
	if len(n.children) != 3 {
		return nil, nodeErrorf(n, "Bad define expression - wrong arity")
	}

	// Syntactix suger '(define (f x) body) -> '(define f (lambda (x) body))'
	IDAndArgs, ok := n.children[1].(NodeList)
	if ok {
		//		fmt.Printf("define lambda: %s\n", n)
		//fmt.Printf("len id + args %d - %s\n", len(IDAndArgs.children), IDAndArgs)
		if len(IDAndArgs.children) == 0 {
			return nil, fmt.Errorf("Bad func define expression - no name")
		}
		id := IDAndArgs.children[0]
		args := NodeList{
			children: IDAndArgs.children[1:],
		}
		//fmt.Printf("id %s\n", id)
		//fmt.Printf("args %s\n", args)

		// Replace (f x) -> f
		n.children[1] = id

		// Replace body -> (lambda 'args' body)
		// TODO: helpers for construction (and/or package constants)
		idLambda := NodeIdentifier{nodeAtom{tok: Token{
			Type:  tokIdentifier,
			Value: "lambda",
		}}}
		children := make([]Node, 0)
		children = append(children, idLambda)
		children = append(children, args)
		children = append(children, n.children[2])
		body := NodeList{
			children: children,
		}
		n.children[2] = body
		//fmt.Printf("define lambda: %s\n", n)
		return transformDefine(n)
	}

	id, ok := n.children[1].(NodeIdentifier)
	if !ok {
		return nil, fmt.Errorf("Bad define expression - invalid identifier type %T", n.children[1])
	}

	// Implicit progn for remaining children
	value, err := Transform(n.children[2])
	if err != nil {
		return nil, err
	}
	return NodeDefine{
		NodeList: n,
		Symbol:   id,
		Value:    value,
	}, nil
}

type NodeLet struct {
	NodeList
	Bindings map[string]Node
	Body     Node
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

type NodeLambda struct {
	NodeList
	Args []Node
	Body Node
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
