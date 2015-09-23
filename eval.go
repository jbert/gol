package gol

import (
	"fmt"
	"io"
)

type Evaluator struct {
	in      io.Reader
	out     io.Writer
	err     io.Writer
	nesting int
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

	if node.IsAtom() {
		switch n := node.(type) {
		case NodeError:
			value = nil
			return nil, n
		case NodeIdentifier:
			var err error
			value, err = env.Lookup(n.String())
			if err != nil {
				return nil, NodeError{node, err.Error()}
			}
		case NodeInt:
			value = n
		case NodeSymbol:
			value = n
		case NodeString:
			value = n
		case NodeBool:
			value = n
		default:
			return nil, fmt.Errorf("Unrecognised atom node type %T", node)
		}
	} else {
		switch n := node.(type) {
		case NodeLambda:
			return e.evalLambda(n, env)
		case NodeList:
			return e.evalList(n, env)
		case NodeIf:
			return e.evalIf(n, env)
		case NodeLet:
			return e.evalLet(n, env)
		case NodeProgn:
			return e.evalProgn(n, env)
		case NodeDefine:
			return e.evalDefine(n, env)
		default:
			return nil, nodeErrorf(n, "Unrecognised list node type %T", node)

		}
	}
	return value, nil
}

func (e *Evaluator) evalDefine(nd NodeDefine, env Environment) (Node, error) {
	evalValue, err := e.Eval(nd.Value, env)
	if err != nil {
		return nd, err
	}
	err = env.AddDefine(nd.Symbol.String(), evalValue)
	if err != nil {
		return nd, err
	}
	return nd.Value, nil
}

func (e *Evaluator) evalIf(ni NodeIf, env Environment) (Node, error) {
	condition, err := e.Eval(ni.Condition, env)
	if err != nil {
		return nil, err
	}
	conditionBool, ok := condition.(NodeBool)
	if !ok {
		return nil, fmt.Errorf("Non-boolean in 'if' condition")
	}
	if conditionBool.IsTrue() {
		return e.Eval(ni.TBranch, env)
	} else {
		return e.Eval(ni.FBranch, env)
	}
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

func (np NodeProcedure) Apply(e *Evaluator, argVals NodeList) (Node, error) {
	if argVals.Len() != np.Args.Len() {
		return nil, fmt.Errorf("Arg mismatch")
	}

	f := Frame{}
	z := np.Args.Zip(argVals)
	_, err := z.Map(func(n Node) (Node, error) {
		pair, ok := n.(NodePair)
		if !ok {
			return nil, fmt.Errorf("Internal error - zip returns non-pair")
		}
		id := pair.car
		val := pair.cdr
		idStr := id.String()

		f[idStr] = val
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	env := np.Env.WithFrame(f)
	return e.Eval(np.Body, env)
}

func (e *Evaluator) evalProgn(np NodeProgn, env Environment) (Node, error) {
	// Value if no children
	var lastVal Node
	lastVal = NodeList{}

	body := np.Rest()
	_, err := body.Map(func(child Node) (Node, error) {
		v, err := e.Eval(child, env)
		if err != nil {
			return nil, err
		}
		lastVal = v
		return v, nil
	})
	if err != nil {
		return nil, err
	}

	// Return last
	return lastVal, nil
}

func (e *Evaluator) evalList(nl NodeList, env Environment) (Node, error) {
	if nl.Len() == 0 {
		return nil, NodeError{nl, "empty application"}
	}

	nodes, err := nl.Map(func(child Node) (Node, error) {
		newVal, err := e.Eval(child, env)
		if err != nil {
			return nil, err
		}
		return newVal, nil
	})
	if err != nil {
		return nil, err
	}

	applicable, ok := nodes.First().(NodeApplicable)
	if !ok {
		return nil, fmt.Errorf("Can't evaluate list with non-applicable head: %T [%s]", nodes.First(), nodes)
	}

	node, err := applicable.Apply(e, nodes.Rest())
	if err != nil {
		return nil, err
	}

	return node, nil
}
