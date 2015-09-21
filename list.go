package gol

import "fmt"

// ----------------------------------------

type NodePair struct {
	car Node
	cdr Node
}

func (np NodePair) String() string {
	if np.IsNil() {
		return "()"
	} else {
		return fmt.Sprintf("(%v %v)", np.car, np.cdr)
	}
}

func (np NodePair) Pos() Position {
	return np.car.Pos()
}

func (np NodePair) IsAtom() bool {
	return false
}

func Nil() NodePair {
	return NodePair{}
}

func (np *NodePair) IsNil() bool {
	return np.car == nil && np.cdr == nil
}

// ----------------------------------------

type NodeList struct {
	NodeBase
	children NodePair
}

func (nl NodeList) Pos() Position {
	if nl.Len() == 0 {
		return Position{File: "<empty list>"}
	} else {
		return nl.First().Pos()
	}
}

func (nl NodeList) IsAtom() bool {
	return false
}

func (nl *NodeList) Map(f func(n Node) (Node, error)) (NodeList, error) {
	p := nl.children
	res := *nl
	res.children = Nil()
	for !p.IsNil() {
		v, err := f(p.car)
		if err != nil {
			return res, err
		}
		res = res.Cons(v)
		var ok bool
		p, ok = p.cdr.(NodePair)
		if !ok {
			panic("Map on improoper list")
		}
	}
	rev := res.Reverse()
	return rev, nil
}

func (nl *NodeList) Len() int {
	l := 0
	nl.Map(func(child Node) (Node, error) {
		l++
		return nil, nil
	})
	return l
}

func (nl NodeList) Nth(n int) Node {
	if n == 0 {
		return nl.children.car
	} else {
		return nl.Rest().Nth(n - 1)
	}
}

func (nl *NodeList) First() Node {
	return nl.children.car
}

func (nl NodeList) Rest() NodeList {
	newList := nl
	var ok bool
	newList.children, ok = newList.children.cdr.(NodePair)
	if !ok {
		panic(fmt.Sprintf("NodeList not a list: %T", newList.children.cdr))
	}
	return newList
}

func (nl NodeList) String() string {
	s := []byte("(")
	n := nl.children
	first := true
	for {
		if n.IsNil() {
			break
		}

		if !first {
			s = append(s, ' ')
		}
		first = false

		s = append(s, n.car.String()...)

		var ok bool
		n, ok = n.cdr.(NodePair)
		if !ok {
			panic(fmt.Sprintf("NodeList with non-list children: %T\n", n))
		}

	}
	s = append(s, ')')
	return string(s)
}

/*
func (nl NodeList) String() string {
	s := "("
	first := true
	nl.Map(func(n Node) (Node, error) {
		if !first {
			first = true
			s += " "
		}
		s += n.String()
		return nil, nil
	})
	s += ")"
	return s
}
*/

func (nl NodeList) Cons(n Node) NodeList {
	ret := nl
	prev := ret.children
	ret.children = NodePair{car: n, cdr: prev}
	return ret
}

func (nl NodeList) Append(n Node) NodeList {
	ret := nl.Reverse()
	ret = ret.Cons(n)
	ret = ret.Reverse()
	return ret
}

func (nl NodeList) Zip(nl2 NodeList) NodeList {
	ret := nl
	ret.children = Nil()

	car := nl.First()
	cdr := nl2.First()
	if car == nil || cdr == nil {
		return ret
	}
	pair := NodePair{car: car, cdr: cdr}
	rest := nl.Rest().Zip(nl2.Rest())
	return rest.Cons(pair)
}

func moveOne(from, to *NodeList) bool {
	if from.children.IsNil() {
		return false
	}
	pair := from.First()
	nextfrom := from.Rest()
	from.children = nextfrom.children

	nextto := to.Cons(pair)
	to.children = nextto.children
	return true
}

func (nl NodeList) Reverse() NodeList {
	ret := nl
	ret.children = Nil()

	// nl is a copy, so ok to modify
	for moveOne(&nl, &ret) {
	}

	return ret
}
