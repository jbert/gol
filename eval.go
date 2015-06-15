package gol

import (
	"fmt"
	"io"
)

type Evaluator struct {
	in  io.Reader
	out io.Writer
	err io.Writer
	env Environment
}

func NewEvaluator(out io.Writer, in io.Reader, err io.Writer) *Evaluator {
	return &Evaluator{
		in:  in,
		out: out,
		err: err,
		env: makeDefaultEnvironment(),
	}
}

func (e *Evaluator) Eval(node Node) (Node, error) {
	var value Node
	switch n := node.(type) {
	case NodeList:
		return e.evalList(n)
	case NodeNum:
		value = n
	case NodeIdentifier:
		var err error
		value, err = e.env.Lookup(n.String())
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

func (e *Evaluator) evalList(nl NodeList) (Node, error) {
	return nil, fmt.Errorf("TODO: implement")
}
