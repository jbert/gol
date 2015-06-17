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
	env = env.WithFrame(nl.Bindings)
	return e.Eval(nl.Body, env)
}
