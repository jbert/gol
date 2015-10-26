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
		case "set!":
			return transformSet(n)
		case "quote":
			return transformQuote(n)
		case "quasiquote":
			return transformQuasiQuote(n)
		case "unquote":
			return transformUnQuote(n)
		}
	}
	ret, err := transformNodes(n)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type NodeUnQuote struct {
	NodeList
	Arg Node
}

func (nq NodeUnQuote) String() string {
	return "," + nq.Arg.String()
}

type NodeQuote struct {
	NodeList
	Arg   Node
	quasi bool
}

func (nq NodeQuote) String() string {
	argStr := nq.Arg.String()
	if nq.quasi {
		return "'" + argStr
	} else {
		return "`" + argStr
	}

}

func transformQuasiQuote(n NodeList) (Node, error) {
	if n.Len() != 2 {
		return nil, fmt.Errorf("Bad quasiquote expression - more than one arg")
	}
	child, err := Transform(n.Rest().First())
	if err != nil {
		return nil, err
	}
	return NodeQuote{NodeList: n, Arg: child, quasi: true}, nil
}

func transformQuote(n NodeList) (Node, error) {
	if n.Len() != 2 {
		return nil, fmt.Errorf("Bad quote expression - more than one arg")
	}
	child, err := Transform(n.Rest().First())
	if err != nil {
		return nil, err
	}
	return NodeQuote{NodeList: n, Arg: child, quasi: false}, nil
}

func transformUnQuote(n NodeList) (Node, error) {
	if n.Len() != 2 {
		return nil, fmt.Errorf("Bad quote expression - more than one arg")
	}
	child, err := Transform(n.Rest().First())
	if err != nil {
		return nil, err
	}
	return NodeUnQuote{NodeList: n, Arg: child}, nil
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

type NodeSet struct {
	NodeList
	Id    NodeIdentifier
	Value Node
}

func transformSet(n NodeList) (Node, error) {
	if n.Len() != 3 {
		return nil, fmt.Errorf("Bad set! expression - missing id or value")
	}
	children, err := transformNodes(n.Rest())
	if err != nil {
		return nil, err
	}
	id, ok := children.First().(NodeIdentifier)
	if !ok {
		return nil, fmt.Errorf("Bad set! expression - non-identifier")
	}
	return NodeSet{
		NodeList: n,
		Id:       id,
		Value:    children.Nth(1),
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
		return transformSugaryDefine(IDAndArgs, n)
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
	return NodeDefine{
		NodeList: n,
		Symbol:   id,
		Value:    makeProgn(children),
	}, nil
}

func transformSugaryDefine(IDAndArgs NodeList, n NodeList) (Node, error) {
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
	body := NodeList{}
	lambdaBody := n.Rest().Rest()
	lambdaBody = lambdaBody.Cons(makeIdentifier("progn"))
	body = body.Cons(lambdaBody)
	body = body.Cons(args)
	body = body.Cons(makeIdentifier("lambda"))

	newDefine = newDefine.Append(body)
	//fmt.Printf("define lambda: %s\n", n)
	return transformDefine(newDefine)
}

type NodeLet struct {
	NodeList
	Bindings Frame
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
	nLet.Body = makeProgn(children)
	return nLet, nil
}

type NodeLambda struct {
	NodeList
	Args NodeList
	Body Node
}

func transformLambda(n NodeList) (Node, error) {
	if n.Len() < 3 {
		return nil, fmt.Errorf("Bad lambda expression - missing args or body [len %d]: %s", n.Len(), n)
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

	body := n.Rest().Rest()
	body, err = transformNodes(body)
	if err != nil {
		return nil, err
	}
	nLambda := NodeLambda{NodeList: n, Args: args, Body: makeProgn(body)}
	return nLambda, nil
}

func makeProgn(nl NodeList) NodeProgn {
	nl = nl.Cons(makeIdentifier("progn"))
	return NodeProgn{nl}
}
