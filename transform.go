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

func transformNodes(nl NodeList) (NodeList, error) {
	return nl.Map(func(child Node) (Node, error) {
		newNode, err := Transform(child)
		if err != nil {
			return nil, err
		}
		return newNode, nil
	})
}

func transformList(n NodeList) (Node, error) {
	if n.Len() == 0 {
		return n, nil
	}

	first := n.First()
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
	ret, err := transformNodes(n.Rest())
	if err != nil {
		return nil, err
	}
	return ret.Cons(first), nil
}

type NodeIf struct {
	NodeList
	Condition Node
	TBranch   Node
	FBranch   Node
}

func transformIf(n NodeList) (Node, error) {
	if n.Len() != 4 {
		return nil, fmt.Errorf("Bad if expression - missing test or t/f branch")
	}
	children, err := transformNodes(n.Rest())
	if err != nil {
		return nil, err
	}
	return NodeIf{
		NodeList:  n,
		Condition: children.First(),
		TBranch:   children.Nth(1),
		FBranch:   children.Nth(2),
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
	if n.Len() != 2 {
		return nil, fmt.Errorf("Bad error expression - exactly one string required")
	}
	return NodeError{n.Nth(1), n.Nth(1).String()}, nil
}

type NodeProgn struct {
	NodeList
}

func transformProgn(n NodeList) (Node, error) {
	children, err := transformNodes(n.Rest())
	if err != nil {
		return nil, err
	}
	return NodeProgn{children.Cons(n.First())}, nil
}

type NodeDefine struct {
	NodeList
	Symbol Node
	Value  Node
}

func transformDefine(n NodeList) (Node, error) {
	if n.Len() < 3 {
		return nil, nodeErrorf(n, "Bad define expression - wrong arity")
	}

	// Syntactix suger '(define (f x) body) -> '(define f (lambda (x) body))'
	IDAndArgs, ok := n.Nth(1).(NodeList)
	if ok {
		//		fmt.Printf("define lambda: %s\n", n)
		//fmt.Printf("len id + args %d - %s\n", len(IDAndArgs.children), IDAndArgs)
		if IDAndArgs.Len() == 0 {
			return nil, fmt.Errorf("Bad func define expression - no name")
		}
		id := IDAndArgs.First()
		args := IDAndArgs.Rest()
		//fmt.Printf("id %s\n", id)
		//fmt.Printf("args %s\n", args)

		newDefine := n
		newDefine.children = NodePair{}
		newDefine = newDefine.Append(n.First())

		// Replace (f x) -> f
		newDefine = newDefine.Append(id)

		// Replace body -> (lambda 'args' body)
		// TODO: helpers for construction (and/or package constants)
		idLambda := NodeIdentifier{nodeAtom{tok: Token{
			Type:  tokIdentifier,
			Value: "lambda",
		}}}
		body := NodeList{}
		body = body.Cons(n.Nth(2))
		body = body.Cons(args)
		body = body.Cons(idLambda)

		newDefine = newDefine.Append(body)
		//fmt.Printf("define lambda: %s\n", n)
		return transformDefine(newDefine)
	}

	id, ok := n.Nth(1).(NodeIdentifier)
	if !ok {
		return nil, fmt.Errorf("Bad define expression - invalid identifier type %T", n.Nth(1))
	}

	// Implicit progn for remaining children
	children, err := transformNodes(n.Rest().Rest())
	if err != nil {
		return nil, err
	}
	children = children.Cons(NodeIdentifier{nodeAtom{tok: Token{
		Type:  tokIdentifier,
		Value: "progn",
	}}})
	return NodeDefine{
		NodeList: n,
		Symbol:   id,
		Value:    NodeProgn{children},
	}, nil
}

type NodeLet struct {
	NodeList
	Bindings map[string]Node
	Body     Node
}

func transformLet(n NodeList) (Node, error) {
	if n.Len() < 3 {
		return nil, fmt.Errorf("Bad let expression - missing bindings or body")
	}
	nLet := NodeLet{NodeList: n, Bindings: make(map[string]Node)}
	bindings, ok := n.Nth(1).(NodeList)
	if !ok {
		return nil, fmt.Errorf("Bad let expression - bindings must be a list")
	}
	_, err := bindings.Map(func(pairNode Node) (Node, error) {
		pair, ok := pairNode.(NodeList)
		if !ok {
			return nil, fmt.Errorf("Bad let expression - bindings must be pairs")
		}
		if pair.Len() != 2 {
			return nil, fmt.Errorf("Bad let expression - bindings must be pairs")
		}
		id, ok := pair.First().(NodeIdentifier)
		if !ok {
			return nil, fmt.Errorf("Bad let expression - invalid identifier")
		}
		var err error
		nLet.Bindings[id.String()], err = Transform(pair.Nth(1))
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	children, err := transformNodes(n.Rest().Rest())
	if err != nil {
		return nil, err
	}
	nLet.Body = NodeProgn{children}
	return nLet, nil
}

type NodeLambda struct {
	NodeList
	Args NodeList
	Body Node
}

func transformLambda(n NodeList) (Node, error) {
	if n.Len() != 3 {
		return nil, fmt.Errorf("Bad lambda expression - missing args or body")
	}
	args, ok := n.Nth(1).(NodeList)
	if !ok {
		return nil, fmt.Errorf("Bad lambda expression - args must be a list")
	}
	_, err := args.Map(func(argNode Node) (Node, error) {
		_, ok := argNode.(NodeIdentifier)
		if !ok {
			return nil, fmt.Errorf("Bad lambda expression - arg must be identifier")
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	// TODO: implicit progn
	bodyList := n.Rest().Rest()
	body, err := Transform(bodyList.First())
	if err != nil {
		return nil, err
	}
	nLambda := NodeLambda{NodeList: n, Args: args, Body: body}
	return nLambda, nil
}
