package gol

import (
	"fmt"
	"io"
)

type Evaluator struct {
	Env     Environment
	in      io.Reader
	out     io.Writer
	err     io.Writer
	nesting int
}

func NewEvaluator(env Environment, out io.Writer, in io.Reader, err io.Writer) *Evaluator {
	return &Evaluator{
		Env: env,
		in:  in,
		out: out,
		err: err,
	}
}

func (e Evaluator) Quoting() bool {
	return e.nesting > 0
}

func (e *Evaluator) Eval(node Node) (Node, error) {
	switch n := node.(type) {
	case NodeError:
		return nil, n
	case NodeIdentifier:
		value, err := e.Env.Lookup(n.String())
		if err != nil {
			return nil, NodeError{node, err.Error()}
		}
		return value, nil
	case NodeInt:
		return n, nil
	case NodeSymbol:
		return n, nil
	case NodeString:
		return n, nil
	case NodeBool:
		return n, nil
	case NodeQuote:
		if n.quasi {
			e.nesting++
			value, err := e.Eval(n.Arg)
			e.nesting--
			return value, err
		}
		return n.Arg, nil
	case NodeUnQuote:
		e.nesting--
		value, err := e.Eval(n.Arg)
		e.nesting++
		return value, err
	case NodeLambda:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalLambda(n)
	case NodeList:
		if e.Quoting() {
			return e.evalList(n)
		}
		return e.evalList(n)
	case NodeIf:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalIf(n)
	case NodeSet:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalSet(n)
	case NodeLet:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalLet(n)
	case NodeProgn:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalProgn(n)
	case NodeDefine:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalDefine(n)
	default:
		return nil, nodeErrorf(n, "Unrecognised node type %T", node)

	}
}

func (e *Evaluator) evalDefine(nd NodeDefine) (Node, error) {
	evalValue, err := e.Eval(nd.Value)
	if err != nil {
		return nd, err
	}
	err = e.Env.AddDefine(nd.Symbol.String(), evalValue)
	if err != nil {
		return nd, err
	}
	return nd.Value, nil
}

func (e *Evaluator) evalSet(ns NodeSet) (Node, error) {
	value, err := e.Eval(ns.Value)
	if err != nil {
		return nil, err
	}
	e.Env.Set(ns.Id.String(), value)

	return value, nil
}

func (e *Evaluator) evalIf(ni NodeIf) (Node, error) {
	condition, err := e.Eval(ni.Condition)
	if err != nil {
		return nil, err
	}
	conditionBool, ok := condition.(NodeBool)
	if !ok {
		return nil, fmt.Errorf("Non-boolean in 'if' condition")
	}
	if conditionBool.IsTrue() {
		return e.Eval(ni.TBranch)
	} else {
		return e.Eval(ni.FBranch)
	}
}

func (e *Evaluator) evalLet(nl NodeLet) (Node, error) {

	f := Frame{}
	oldEnv := e.Env
	defer func() {
		e.Env = oldEnv
	}()
	e.Env = e.Env.WithFrame(f)

	for k, _ := range nl.Bindings {
		f[k] = NODE_NIL
	}
	for k, v := range nl.Bindings {
		var err error
		f[k], err = e.Eval(v)
		if err != nil {
			return nil, err
		}
	}
	value, err := e.Eval(nl.Body)
	return value, err
}

type NodeProcedure struct {
	NodeLambda
	Env Environment
}

func (e *Evaluator) evalLambda(nl NodeLambda) (Node, error) {
	return NodeProcedure{
		NodeLambda: nl,
		Env:        e.Env,
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

	oldEnv := e.Env
	defer func() {
		e.Env = oldEnv
	}()
	e.Env = np.Env.WithFrame(f)
	value, err := e.Eval(np.Body)
	return value, err
}

func (e *Evaluator) evalProgn(np NodeProgn) (Node, error) {
	// Value if no children
	var lastVal Node
	lastVal = NodeList{}

	body := np.Rest()
	_, err := body.Map(func(child Node) (Node, error) {
		v, err := e.Eval(child)
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

func (e *Evaluator) evalList(nl NodeList) (Node, error) {
	nodes, err := nl.Map(func(child Node) (Node, error) {
		newVal, err := e.Eval(child)
		if err != nil {
			return nil, err
		}
		return newVal, nil
	})
	if err != nil {
		return nil, err
	}

	if e.Quoting() {
		return nodes, nil
	}

	return e.Apply(nodes)
}

func (e *Evaluator) Apply(nl NodeList) (Node, error) {
	if nl.Len() == 0 {
		return nil, NodeError{nl, "empty application"}
	}
	applicable, ok := nl.First().(NodeApplicable)
	if !ok {
		return nil, fmt.Errorf("Can't evaluate list with non-applicable head: %T [%s]", nl.First(), nl)
	}

	node, err := applicable.Apply(e, nl.Rest())
	if err != nil {
		return nil, err
	}

	return node, nil
}
