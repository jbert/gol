package gol

import "fmt"

func (nl NodeList) Map(f func(n Node) (Node, error)) (NodeList, error) {
	p := nl.children
	res := nl
	res.children = Nil()
	for !p.IsNil() {
		v, err := f(p.Car)
		if err != nil {
			return res, err
		}
		res = res.Cons(v)
		var ok bool
		p, ok = p.Cdr.(NodePair)
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
		return nl.children.Car
	} else {
		return nl.Rest().Nth(n - 1)
	}
}

func (nl NodeList) First() Node {
	return nl.children.Car
}

func (nl NodeList) Rest() NodeList {
	newList := nl
	var ok bool
	newList.children, ok = newList.children.Cdr.(NodePair)
	if !ok {
		panic(fmt.Sprintf("NodeList not a list: %T", newList.children.Cdr))
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

		s = append(s, n.Car.String()...)
		//		s = append(s, []byte(fmt.Sprintf(" [%T]", n.Car))...)

		var ok bool
		n, ok = n.Cdr.(NodePair)
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
	ret.children = NewNodePair(n, prev)
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
	pair := NewNodePair(car, cdr)
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
