package gol

import (
	"errors"
	"fmt"
)

type Frame map[string]Node

type Environment []Frame

func MakeDefaultEnvironment() Environment {
	defEnv := []Frame{
		Frame{"+": NodeBuiltin{f: addInt, description: "+"}},
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

var ErrNotFound = errors.New("Identifier not found")

func (e Environment) Lookup(s string) (Node, error) {
	for _, f := range []Frame(e) {
		node, ok := f[s]
		if ok {
			return node, nil
		}
	}
	return nil, ErrNotFound
}

type NodeApplicable interface {
	Node
	Apply(nodes []Node) (Node, error)
}

type NodeBuiltin struct {
	NodeBase
	f           func(nodes []Node) (Node, error)
	description string
}

func (nb NodeBuiltin) String() string {
	return nb.description
}

func (nb NodeBuiltin) Apply(args []Node) (Node, error) {
	return nb.f(args)
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
