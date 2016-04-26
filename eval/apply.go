package eval

import (
	"fmt"

	"github.com/jbert/gol"
)

func MakeDefaultEnvironment() Environment {
	defEnv := []gol.Frame{
		gol.Frame{
			"=":       &NodeBuiltin{f: equalInt, description: "="},
			"+":       &NodeBuiltin{f: addInt, description: "+"},
			"-":       &NodeBuiltin{f: subInt, description: "-"},
			"*":       &NodeBuiltin{f: mulInt, description: "*"},
			"display": &NodeBuiltin{f: display, description: "display"},
			"list":    &NodeBuiltin{f: list, description: "list"},
			"length":  &NodeBuiltin{f: length, description: "length"},
			"reverse": &NodeBuiltin{f: reverse, description: "reverse"},
			"append":  &NodeBuiltin{f: listAppend, description: "append"},
			"apply":   &NodeBuiltin{f: apply, description: "apply"},
			"zero?":   &NodeBuiltin{f: zerop, description: "zerop"},
		},
	}
	return defEnv
}

type NodeApplicable interface {
	gol.Node
	Apply(e *Evaluator, nodes *gol.NodeList) (gol.Node, error)
}

type NodeBuiltin struct {
	gol.NodeBase
	f           func(e *Evaluator, nodes *gol.NodeList) (gol.Node, error)
	description string
}

func (nb NodeBuiltin) Pos() gol.Position {
	return gol.Position{File: "<builtin>"}
}

func (nb NodeBuiltin) String() string {
	return nb.description
}

func (nb NodeBuiltin) Apply(e *Evaluator, args *gol.NodeList) (gol.Node, error) {
	return nb.f(e, args)
}

func equalInt(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	if nodes.Len() < 2 {
		return nil, gol.NodeErrorf(nodes, "At least two arguments required")
	}
	first, ok := nodes.First().(*gol.NodeInt)
	if !ok {
		return nil, gol.NodeErrorf(nodes, "Non-int passed to equalInt")
	}
	ret := gol.NODE_TRUE

	rest := nodes.Rest()
	_, err := rest.Map(func(n gol.Node) (gol.Node, error) {
		ni, ok := n.(*gol.NodeInt)
		if !ok {
			return nil, gol.NodeErrorf(nodes, "Non-int passed to equalInt")
		}
		if first.Value() != ni.Value() {
			ret = gol.NODE_FALSE
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func addInt(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	var sum int64
	_, err := nodes.Map(func(n gol.Node) (gol.Node, error) {
		ni, ok := n.(*gol.NodeInt)
		if !ok {
			return nil, gol.NodeErrorf(nodes, "Non-int passed to addInt")
		}
		sum += ni.Value()
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return gol.NewNodeInt(sum), nil
}

func mulInt(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	var prod int64
	prod = 1
	_, err := nodes.Map(func(n gol.Node) (gol.Node, error) {
		ni, ok := n.(*gol.NodeInt)
		if !ok {
			return nil, gol.NodeErrorf(nodes, "Non-int passed to addInt")
		}
		prod *= ni.Value()
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return gol.NewNodeInt(prod), nil
}

func subInt(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	if nodes.Len() == 0 {
		return nil, gol.NodeErrorf(nodes, "Arity-error: expected > 0 args")
	}

	ni, ok := nodes.First().(*gol.NodeInt)
	if !ok {
		return nil, gol.NodeErrorf(nodes, "Non-int passed to subInt")
	}
	result := ni.Value()
	if nodes.Len() == 1 {
		return gol.NewNodeInt(-result), nil
	}

	rest := nodes.Rest()
	_, err := rest.Map(func(n gol.Node) (gol.Node, error) {
		ni, ok := n.(*gol.NodeInt)
		if !ok {
			return nil, gol.NodeErrorf(nodes, "Non-int passed to subInt")
		}
		result -= ni.Value()
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return gol.NewNodeInt(result), nil
}

func display(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	if nodes.Len() != 1 {
		return nil, gol.NodeErrorf(nodes, "Arity-error: expected == 1 args")
	}

	s := nodes.First().String()
	fmt.Fprintf(e.out, "%s", s)
	return gol.NODE_NIL, nil
}

func list(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	return nodes, nil
}

func length(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	if nodes.Len() != 1 {
		return nil, gol.NodeErrorf(nodes, "Arity-error: expected == 1 args")
	}
	nl, ok := nodes.First().(*gol.NodeList)
	if !ok {
		return nil, gol.NodeErrorf(nodes, "Non-list passed to lemgth")
	}

	return gol.NewNodeInt(int64(nl.Len())), nil
}

func reverse(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	if nodes.Len() != 1 {
		return nil, gol.NodeErrorf(nodes, "Arity-error: expected == 1 args")
	}
	nl, ok := nodes.First().(*gol.NodeList)
	if !ok {
		return nil, gol.NodeErrorf(nodes, "Non-list passed to reverse")
	}

	return nl.ReverseCopy(), nil
}

// Given a list-of-lists, return the flattened list containing all the members
func listAppend(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	rev := nodes.ReverseCopy()
	ret, ok := rev.First().(*gol.NodeList)
	if !ok {
		return nil, gol.NodeErrorf(nodes, "Non-list passed to append: %s %T", ret, ret)
	}

	_, err := rev.Rest().Map(func(child gol.Node) (gol.Node, error) {
		l, ok := child.(*gol.NodeList)
		if !ok {
			return nil, gol.NodeErrorf(nodes, "Non-list passed to append: %s %T", child, child)
		}
		l.ReverseCopy().Map(func(lChild gol.Node) (gol.Node, error) {
			ret = ret.Cons(lChild)
			return nil, nil
		})
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func apply(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	// Last should be a list, we
	args := nodes.ReverseCopy()
	l, ok := args.First().(*gol.NodeList)
	if !ok {
		return nil, gol.NodeErrorf(nodes, "Non-list passed as last arg to apply")
	}
	args.Rest().Map(func(child gol.Node) (gol.Node, error) {
		l = l.Cons(child)
		return nil, nil
	})

	return e.Apply(l)
}

func zerop(e *Evaluator, nodes *gol.NodeList) (gol.Node, error) {
	if nodes.Len() != 1 {
		return nil, gol.NodeErrorf(nodes, "Arity-error: expected == 1 args")
	}
	ni, ok := nodes.First().(*gol.NodeInt)
	if !ok {
		return nil, gol.NodeErrorf(nodes, "Non-int passed to zero?")
	}
	if ni.Value() == 0 {
		return gol.NODE_TRUE, nil
	} else {
		return gol.NODE_FALSE, nil
	}
}
