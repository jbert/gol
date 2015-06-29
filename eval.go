package gol

import (
	"fmt"
	"io"
)

type Evaluator struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func NewEvaluator(out io.Writer, in io.Reader, err io.Writer) *Evaluator {
	return &Evaluator{
		in:  in,
		out: out,
		err: err,
	}
}

func (e *Evaluator) Eval(node Node, env Environment) (Node, error) {
	var value Node
	switch n := node.(type) {
	case NodeLet:
		return e.evalLet(n, env)
	case NodeList:
		return e.evalList(n, env)
	case NodeInt:
		value = n
	case NodeIdentifier:
		var err error
		value, err = env.Lookup(n.String())
		if err != nil {
			return nil, err
		}
	case NodeSymbol:
		value = n
	default:
		return nil, fmt.Errorf("Unrecognised node type %T", node)
	}
	return value, nil
}

func (e *Evaluator) evalLet(nl NodeLet, env Environment) (Node, error) {
	f := Frame{}
	for k, v := range nl.Bindings {
		var err error
		f[k], err = e.Eval(v, env)
		if err != nil {
			return nil, err
		}
	}
	env = env.WithFrame(f)
	return e.Eval(nl.Body, env)
}

func (e *Evaluator) evalList(nl NodeList, env Environment) (Node, error) {
	if len(nl.children) == 0 {
		return nl, nil // empty list self-evaluates
	}

	nodes := make([]Node, 0)
	for _, child := range nl.children {
		newVal, err := e.Eval(child, env)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, newVal)

	}

	applicable, ok := nodes[0].(NodeApplicable)
	if !ok {
		return nil, fmt.Errorf("Can't evaluate list with non-applicable head: %T", nodes[0])
	}

	node, err := applicable.Apply(nodes[1:])
	if err != nil {
		return nil, err
	}

	return node, nil
}
