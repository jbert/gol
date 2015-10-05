package gol

import "fmt"

type Frame map[string]Node

type Environment []Frame

func MakeDefaultEnvironment() Environment {
	defEnv := []Frame{
		Frame{
			"=":       NodeBuiltin{f: equalInt, description: "="},
			"+":       NodeBuiltin{f: addInt, description: "+"},
			"-":       NodeBuiltin{f: subInt, description: "-"},
			"*":       NodeBuiltin{f: mulInt, description: "*"},
			"display": NodeBuiltin{f: display, description: "display"},
			"list":    NodeBuiltin{f: list, description: "list"},
		},
	}
	return defEnv
}

func (e Environment) WithFrame(f Frame) Environment {
	// 'append on the front'
	// Slow to build, but fast to look up
	newEnv := []Frame{f}
	newEnv = append(newEnv, e...)
	return newEnv
}

func (e Environment) Lookup(s string) (Node, error) {
	for _, f := range []Frame(e) {
		node, ok := f[s]
		if ok {
			return node, nil
		}
	}
	return nil, fmt.Errorf("Identifier [%s] not found", s)
}

func (e Environment) AddDefine(id string, value Node) error {
	// Add to top-level frame (at end)
	topLevel := e[len(e)-1]
	topLevel[id] = value
	return nil
}

type NodeApplicable interface {
	Node
	Apply(e *Evaluator, nodes NodeList) (Node, error)
}

type NodeBuiltin struct {
	NodeBase
	f           func(e *Evaluator, nodes NodeList) (Node, error)
	description string
}

func (nb NodeBuiltin) Pos() Position {
	return Position{File: "<builtin>"}
}

func (nb NodeBuiltin) IsAtom() bool {
	return false
}

func (nb NodeBuiltin) String() string {
	return nb.description
}

func (nb NodeBuiltin) Apply(e *Evaluator, args NodeList) (Node, error) {
	return nb.f(e, args)
}

var NODE_FALSE = NodeBool{
	nodeAtom{
		tok: Token{
			Type:  tokBool,
			Value: "#f",
		},
	},
}

var NODE_TRUE = NodeBool{
	nodeAtom{
		tok: Token{
			Type:  tokBool,
			Value: "#t",
		},
	},
}

var NODE_NIL = NodeList{}

func equalInt(e *Evaluator, nodes NodeList) (Node, error) {
	if nodes.Len() < 2 {
		return nil, fmt.Errorf("At least two arguments required")
	}
	first, ok := nodes.First().(NodeInt)
	if !ok {
		return nil, fmt.Errorf("Non-int passed to equalInt")
	}
	ret := NODE_TRUE

	rest := nodes.Rest()
	_, err := rest.Map(func(n Node) (Node, error) {
		ni, ok := n.(NodeInt)
		if !ok {
			return nil, fmt.Errorf("Non-int passed to equalInt")
		}
		if first.Value() != ni.Value() {
			ret = NODE_FALSE
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func addInt(e *Evaluator, nodes NodeList) (Node, error) {
	var sum int64
	_, err := nodes.Map(func(n Node) (Node, error) {
		ni, ok := n.(NodeInt)
		if !ok {
			return nil, fmt.Errorf("Non-int passed to addInt")
		}
		sum += ni.Value()
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return NodeInt{value: sum}, nil
}

func mulInt(e *Evaluator, nodes NodeList) (Node, error) {
	var prod int64
	prod = 1
	_, err := nodes.Map(func(n Node) (Node, error) {
		ni, ok := n.(NodeInt)
		if !ok {
			return nil, fmt.Errorf("Non-int passed to addInt")
		}
		prod *= ni.Value()
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return NodeInt{value: prod}, nil
}

func subInt(e *Evaluator, nodes NodeList) (Node, error) {
	if nodes.Len() == 0 {
		return nil, fmt.Errorf("Arity-error: expected > 0 args")
	}

	ni, ok := nodes.First().(NodeInt)
	if !ok {
		return nil, fmt.Errorf("Non-int passed to subInt")
	}
	result := ni.Value()
	if nodes.Len() == 1 {
		return NodeInt{value: -result}, nil
	}

	rest := nodes.Rest()
	_, err := rest.Map(func(n Node) (Node, error) {
		ni, ok := n.(NodeInt)
		if !ok {
			return nil, fmt.Errorf("Non-int passed to subInt")
		}
		result -= ni.Value()
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return NodeInt{value: result}, nil
}

func display(e *Evaluator, nodes NodeList) (Node, error) {
	if nodes.Len() != 1 {
		return nil, fmt.Errorf("Arity-error: expected == 1 args")
	}

	s := nodes.First().String()
	fmt.Fprintf(e.out, "%s", s)
	return NODE_NIL, nil
}

func list(e *Evaluator, nodes NodeList) (Node, error) {
	return nodes, nil
}
