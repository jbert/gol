package eval

import (
	"io"

	"github.com/jbert/gol"
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

func (e *Evaluator) Eval(node gol.Node) (gol.Node, error) {
	switch n := node.(type) {
	case *gol.NodeError:
		return nil, n
	case *gol.NodeIdentifier:
		value, err := e.Env.Lookup(n.String())
		if err != nil {
			return nil, gol.NodeErrorf(node, "Failed to find [%s]: %s", n.String(), err.Error())
		}
		return value, nil
	case *gol.NodeInt:
		return n, nil
	case *gol.NodeSymbol:
		return n, nil
	case *gol.NodeString:
		return n, nil
	case *gol.NodeBool:
		return n, nil
	case *gol.NodeQuote:
		if n.Quasi {
			e.nesting++
			value, err := e.Eval(n.Arg)
			e.nesting--
			return value, err
		}
		return n.Arg, nil
	case *gol.NodeUnQuote:
		e.nesting--
		value, err := e.Eval(n.Arg)
		e.nesting++
		return value, err
	case *gol.NodeLambda:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalLambda(n)
	case *gol.NodeList:
		if e.Quoting() {
			return e.evalList(n)
		}
		return e.evalList(n)
	case *gol.NodeIf:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalIf(n)
	case *gol.NodeSet:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalSet(n)
	case *gol.NodeLet:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalLet(n)
	case *gol.NodeProgn:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalProgn(n)
	case *gol.NodeDefine:
		if e.Quoting() {
			return e.evalList(n.NodeList)
		}
		return e.evalDefine(n)
	default:
		return nil, gol.NodeErrorf(n, "Unrecognised node type %T", node)

	}
}

func (e *Evaluator) evalDefine(nd *gol.NodeDefine) (gol.Node, error) {
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

func (e *Evaluator) evalSet(ns *gol.NodeSet) (gol.Node, error) {
	value, err := e.Eval(ns.Value)
	if err != nil {
		return nil, err
	}
	e.Env.Set(ns.Id.String(), value)

	return value, nil
}

func (e *Evaluator) evalIf(ni *gol.NodeIf) (gol.Node, error) {
	condition, err := e.Eval(ni.Condition)
	if err != nil {
		return nil, err
	}
	conditionBool, ok := condition.(*gol.NodeBool)
	if !ok {
		return nil, gol.NodeErrorf(ni, "Non-boolean in 'if' condition")
	}
	if conditionBool.IsTrue() {
		return e.Eval(ni.TBranch)
	} else {
		return e.Eval(ni.FBranch)
	}
}

func (e *Evaluator) evalLet(nl *gol.NodeLet) (gol.Node, error) {

	f := gol.Frame{}
	oldEnv := e.Env
	defer func() {
		e.Env = oldEnv
	}()
	e.Env = e.Env.WithFrame(f)

	for k, _ := range nl.Bindings {
		f[k] = gol.NODE_NIL
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

func (e *Evaluator) evalLambda(nl *gol.NodeLambda) (gol.Node, error) {
	return &NodeProcedure{
		NodeLambda: nl,
		Env:        e.Env,
	}, nil
}

func (np NodeProcedure) Apply(e *Evaluator, argVals *gol.NodeList) (gol.Node, error) {
	if argVals.Len() != np.Args.Len() {
		return nil, gol.NodeErrorf(argVals, "Arg mismatch")
	}

	f := gol.Frame{}
	z := np.Args.Zip(argVals)
	err := z.Foreach(func(n gol.Node) error {
		pair, ok := n.(*gol.NodePair)
		if !ok {
			return gol.NodeErrorf(argVals, "Internal error - zip returns non-pair")
		}
		id := pair.Car
		val := pair.Cdr
		idStr := id.String()

		f[idStr] = val
		return nil
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

func (e *Evaluator) evalProgn(np *gol.NodeProgn) (gol.Node, error) {
	// Value if no children
	var lastVal gol.Node
	lastVal = gol.NewNodeList()

	body := np.Rest()
	err := body.Foreach(func(child gol.Node) error {
		v, err := e.Eval(child)
		if err != nil {
			return err
		}
		lastVal = v
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Return last
	return lastVal, nil
}

func (e *Evaluator) evalList(nl *gol.NodeList) (gol.Node, error) {
	nodes, err := nl.Map(func(child gol.Node) (gol.Node, error) {
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

func (e *Evaluator) Apply(nl *gol.NodeList) (gol.Node, error) {
	if nl.Len() == 0 {
		return nil, gol.NodeErrorf(nl, "empty application")
	}
	applicable, ok := nl.First().(NodeApplicable)
	if !ok {
		return nil, gol.NodeErrorf(nl, "Can't evaluate list with non-applicable head: %T [%s]", nl.First(), nl)
	}

	node, err := applicable.Apply(e, nl.Rest())
	if err != nil {
		return nil, err
	}

	return node, nil
}
