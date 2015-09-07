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
	Apply(e *Evaluator, nodes []Node) (Node, error)
}

type NodeBuiltin struct {
	NodeBase
	f           func(e *Evaluator, nodes []Node) (Node, error)
	description string
}

func (nb NodeBuiltin) IsAtom() bool {
	return false
}

func (nb NodeBuiltin) String() string {
	return nb.description
}

func (nb NodeBuiltin) Apply(e *Evaluator, args []Node) (Node, error) {
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

func equalInt(e *Evaluator, nodes []Node) (Node, error) {
	if len(nodes) < 2 {
		return nil, fmt.Errorf("At least two arguments required")
	}
	first, ok := nodes[0].(NodeInt)
	if !ok {
		return nil, fmt.Errorf("Non-int passed to equalInt")
	}
	for _, n := range nodes[1:] {
		ni, ok := n.(NodeInt)
		if !ok {
			return nil, fmt.Errorf("Non-int passed to equalInt")
		}
		if first.Value() != ni.Value() {
			return NODE_FALSE, nil
		}
	}
	return NODE_TRUE, nil
}

func addInt(e *Evaluator, nodes []Node) (Node, error) {
	var sum int64
	for _, n := range nodes {
		ni, ok := n.(NodeInt)
		if !ok {
			return nil, fmt.Errorf("Non-int passed to addInt")
		}
		sum += ni.Value()
	}
	return NodeInt{value: sum}, nil
}

func mulInt(e *Evaluator, nodes []Node) (Node, error) {
	var prod int64
	prod = 1
	for _, n := range nodes {
		ni, ok := n.(NodeInt)
		if !ok {
			return nil, fmt.Errorf("Non-int passed to addInt")
		}
		prod *= ni.Value()
	}
	return NodeInt{value: prod}, nil
}

func subInt(e *Evaluator, nodes []Node) (Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("Arity-error: expected > 0 args")
	}

	ni, ok := nodes[0].(NodeInt)
	if !ok {
		return nil, fmt.Errorf("Non-int passed to subInt")
	}
	result := ni.Value()
	if len(nodes) == 1 {
		return NodeInt{value: -result}, nil
	}

	for _, n := range nodes[1:] {
		ni, ok := n.(NodeInt)
		if !ok {
			return nil, fmt.Errorf("Non-int passed to subInt")
		}
		result -= ni.Value()
	}
	return NodeInt{value: result}, nil
}

func display(e *Evaluator, nodes []Node) (Node, error) {
	if len(nodes) != 1 {
		return nil, fmt.Errorf("Arity-error: expected == 1 args")
	}

	s := nodes[0].String()
	fmt.Fprintf(e.out, "%s", s)
	return NODE_NIL, nil
}
