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
	case NodeProgn:
		return e.evalProgn(n, env)
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
	case NodeLambda:
		return e.evalLambda(n, env)
	case NodeSymbol:
		value = n
	case NodeString:
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

type NodeProcedure struct {
	NodeLambda
	Env Environment
}

func (e *Evaluator) evalLambda(nl NodeLambda, env Environment) (Node, error) {
	return NodeProcedure{
		NodeLambda: nl,
		Env:        env,
	}, nil
}

func (np NodeProcedure) Apply(e *Evaluator, argVals []Node) (Node, error) {
	if len(argVals) != len(np.Args) {
		return nil, fmt.Errorf("Arg mismatch")
	}

	f := Frame{}
	for i, argVal := range argVals {
		idStr := np.Args[i].String()

		f[idStr] = argVal
	}
	env := np.Env.WithFrame(f)
	return e.Eval(np.Body, env)
}

func (e *Evaluator) evalProgn(np NodeProgn, env Environment) (Node, error) {
	last := len(np.children)
	if last == 0 {
		return NodeList{children: make([]Node, 0)}, nil
	}
	return e.Eval(np.children[last-1], env)
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

	node, err := applicable.Apply(e, nodes[1:])
	if err != nil {
		return nil, err
	}

	return node, nil
}
