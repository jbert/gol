package gol

import "fmt"

type Frame map[string]Node

type Environment []Frame

func MakeDefaultEnvironment() Environment {
	defEnv := []Frame{
		Frame{
			"+": NodeBuiltin{f: addInt, description: "+"},
			"-": NodeBuiltin{f: subInt, description: "-"},
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

type NodeApplicable interface {
	Node
	Apply(e *Evaluator, nodes []Node) (Node, error)
}

type NodeBuiltin struct {
	NodeBase
	f           func(nodes []Node) (Node, error)
	description string
}

func (nb NodeBuiltin) IsAtom() bool {
	return false
}

func (nb NodeBuiltin) String() string {
	return nb.description
}

func (nb NodeBuiltin) Apply(e *Evaluator, args []Node) (Node, error) {
	return nb.f(args)
}

type NodeLambda struct {
	NodeList
	Args []Node
	Body Node
}

func addInt(nodes []Node) (Node, error) {
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

func subInt(nodes []Node) (Node, error) {
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
