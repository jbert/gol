package gol

import "fmt"

func (nl *NodeList) ForeachLast(f func(n Node, last bool) error) error {
	p := nl.children
	for !p.IsNil() {
		pNext, ok := p.Cdr.(*NodePair)
		if !ok {
			panic("Foreach on improoper list")
		}

		err := f(p.Car, pNext.IsNil())
		if err != nil {
			return err
		}
		p = pNext
	}
	return nil
}

func (nl *NodeList) Foreach(f func(n Node) error) error {
	f2 := func(n Node, last bool) error {
		return f(n)
	}
	return nl.ForeachLast(f2)
}

func (nl *NodeList) Map(f func(n Node) (Node, error)) (*NodeList, error) {
	p := nl.children
	res := NewNodeList()
	for !p.IsNil() {
		v, err := f(p.Car)
		if err != nil {
			return nil, err
		}
		if v == nil {
			panic("nil node back from map function")
		}
		res = res.Cons(v)
		var ok bool
		p, ok = p.Cdr.(*NodePair)
		if !ok {
			panic("Map on improoper list")
		}
	}
	rev := res.ReverseCopy()
	return rev, nil
}

func (nl *NodeList) Len() int {
	l := 0
	nl.Foreach(func(child Node) error {
		l++
		return nil
	})
	return l
}

func (nl *NodeList) Nth(n int) Node {
	if n == 0 {
		return nl.children.Car
	} else {
		return nl.Rest().Nth(n - 1)
	}
}

func (nl *NodeList) First() Node {
	return nl.children.Car
}

func (nl *NodeList) Rest() *NodeList {
	listCopy := *nl
	newList := &listCopy
	var ok bool
	newList.children, ok = newList.children.Cdr.(*NodePair)
	if !ok {
		panic(fmt.Sprintf("NodeList not a list: %T", newList.children.Cdr))
	}
	return newList
}

func (nl *NodeList) String() string {
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
		n, ok = n.Cdr.(*NodePair)
		if !ok {
			panic(fmt.Sprintf("NodeList with non-list children: %T\n", n))
		}

	}
	s = append(s, ')')
	return string(s)
}

/*
func (nl *NodeList) String() string {
	s := "("
	first := true
	nl.Map(func(n *Node) (*Node, error) {
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

func (nl *NodeList) Cons(n Node) *NodeList {
	listCopy := *nl
	listCopy.children = NewNodePair(n, nl.children)
	return &listCopy
}

func (nl *NodeList) Append(n Node) *NodeList {
	ret := nl.ReverseCopy()
	ret = ret.Cons(n)
	ret = ret.ReverseCopy()
	return ret
}

func (nl *NodeList) Zip(nl2 *NodeList) *NodeList {
	listCopy := *nl
	ret := &listCopy
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

func (nl *NodeList) ReverseCopy() *NodeList {
	ret := NewNodeList()
	ret.NodeBase = nl.NodeBase

	from := nl.children
	for !from.IsNil() {
		pairCopy := *from
		pairCopy.Cdr = ret.children
		ret.children = &pairCopy

		var ok bool
		from, ok = from.Cdr.(*NodePair)
		if !ok {
			panic("ReverseCopy on improoper list")
		}
	}

	return ret
}
