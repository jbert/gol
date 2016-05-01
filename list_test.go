package gol

import "testing"

func makeListTo(n int64) *NodeList {
	nl := NewNodeList()
	for i := n; i > 0; i-- {
		nl = nl.Cons(NewNodeInt(i))
	}
	return nl
}

func TestListZip(t *testing.T) {
	l := makeListTo(5)
	l2, _ := l.Map(func(n Node) (Node, error) {
		ni := n.(*NodeInt)
		return NewNodeInt(ni.Value() * 2), nil
	})
	if l2.String() != "(2 4 6 8 10)" {
		t.Fatalf("Can't make doubled list: %s", l2)
	}

	z := l.Zip(l2)
	if z.String() != "((1 . 2) (2 . 4) (3 . 6) (4 . 8) (5 . 10))" {
		t.Fatalf("Can't zip: %s", z)
	}

	if l.String() != "(1 2 3 4 5)" {
		t.Fatalf("zip changed initial list: %s", l)
	}
	if l2.String() != "(2 4 6 8 10)" {
		t.Fatalf("Zip changed doubled list: %s", l2)
	}
}

func TestListFirstRest(t *testing.T) {
	l := makeListTo(5)
	if l.String() != "(1 2 3 4 5)" {
		t.Fatalf("Can't make initial list")
	}
	if l.First().String() != "1" {
		t.Fatalf("Can't get first")
	}
	if l.Rest().String() != "(2 3 4 5)" {
		t.Fatalf("Can't get rest")
	}
	if l.String() != "(1 2 3 4 5)" {
		t.Fatalf("First or rest changed list")
	}
}

func TestListReverseOld(t *testing.T) {
	p := Nil()
	if !p.IsNil() {
		t.Fatalf("Nil pair isn't nil")
	}

	l1 := &NodeList{
		children: &NodePair{
			Car: NewNodeInt(2),
			Cdr: &NodePair{
				Car: NewNodeInt(1),
				Cdr: &NodePair{}}},
	}

	if l1.Len() != 2 {
		t.Fatalf("Wrong length of l1: %d", l1.Len())
	}

	l2 := l1.ReverseCopy()
	if l2.Len() != 2 {
		t.Fatalf("Wrong length of l2: %d", l1.Len())
	}

	if l2.First().(*NodeInt).Value() != 1 {
		t.Fatalf("Wrong reversed item")
	}
}

func TestListReverse(t *testing.T) {
	l := makeListTo(5)
	if l.String() != "(1 2 3 4 5)" {
		t.Fatalf("Can't make initial list")
	}
	rc := l.ReverseCopy()
	if rc.String() != "(5 4 3 2 1)" {
		t.Fatalf("Can't reverse list: %s", rc)
	}
	if l.String() != "(1 2 3 4 5)" {
		t.Fatalf("Reverse changed list: %s", l)
	}
}

func TestListCons(t *testing.T) {
	t.Log("TestListCons")

	l := NewNodeList()
	if l.Len() != 0 {
		t.Fatalf("Empty list doesn't have zero length")
	}
	l2 := l.Cons(NewNodeInt(1))
	if l.Len() != 0 {
		t.Fatalf("Cons changes the source")
	}
	if l2.Len() != 1 {
		t.Fatalf("Cons doesn't return a longer list")
	}

	n := l2.First()
	ni, ok := n.(*NodeInt)
	if !ok {
		t.Fatalf("cons-then-first is the wrong type (not NodeInt): %T", n)
	}
	if ni.Value() != 1 {
		t.Fatalf("cons-then-first is the wrong value")
	}

}

func TestListMap(t *testing.T) {
	l := makeListTo(5)
	t.Logf("initial list: %s\n", l)
	if l.String() != "(1 2 3 4 5)" {
		t.Fatalf("Can't make initial list")
	}

	expected := int64(1)

	m, err := l.Map(func(n Node) (Node, error) {
		ni, ok := n.(*NodeInt)
		if !ok {
			t.Errorf("Non-int node in list")
			return nil, nil
		}
		if ni.Value() != expected {
			t.Errorf("Wrong value in position %d (%d)", expected, ni.Value())
			return nil, nil
		}
		expected++
		return NewNodeInt(-ni.Value()), nil
	})
	if err != nil {
		t.Errorf("Error return from map: %s", err)
	}
	if expected != 6 {
		t.Errorf("Wrong number of nodes: %d != %d", expected-1, 5)
	}

	if m.Len() != int(expected-1) {
		t.Errorf("Wrong number of nodes in returned list: %d != %d", expected-1, 5)
	}
	t.Logf("map return: %s\n", m)

	if l.String() != "(1 2 3 4 5)" {
		t.Fatalf("Map changed initial list: %s", l.String())
	}

	expected = -1
	m2, _ := m.Map(func(n Node) (Node, error) {
		ni, ok := n.(*NodeInt)
		if !ok {
			t.Errorf("Non-int node in list")
			return nil, nil
		}
		if ni.Value() != expected {
			t.Errorf("Wrong value in position %d (%d)", expected, ni.Value())
			return nil, nil
		}
		expected--
		return NewNodeInt(-ni.Value()), nil
	})
	t.Logf("second map return: %s\n", m2)
}

/* ----
 * Not sure if we want this - we need to reach inside the Node to change it,
 * which we shouldn't be doing anyway
func TestListDestructiveMap(t *testing.T) {
	l := makeListTo(5)
	t.Logf("initial list: %s\n", l)
	if l.String() != "(1 2 3 4 5)" {
		t.Fatalf("Can't make initial list")
	}

	l2, err := l.Map(func(n Node) (Node, error) {
		ni, ok := n.(*NodeInt)
		if !ok {
			t.Errorf("Non-int node in list")
			return nil, nil
		}
		// Reach inside the node to change it's value
		ni.value *= 2
		return ni, nil
	})
	if err != nil {
		t.Fatalf("wtf")
	}
	t.Logf("initial list after map: %s\n", l)

	if l2.String() != "(2 4 6 8 10)" {
		t.Fatalf("Modifying map failed")
	}

	if l.String() != "(1 2 3 4 5)" {
		t.Fatalf("Modifying map changed original list")
	}
}
*/

// List copy funcs were copying the lazy-init type, check for that here
func TestListCopyType(t *testing.T) {
	l := makeListTo(5)
	lStr := l.Type().String() // Force lazy init before copy

	l2, _ := l.Map(func(child Node) (Node, error) {
		ni := child.(*NodeInt)
		return NewNodeInt(ni.Value() * 2), nil
	})

	l2Str := l2.Type().String()
	if lStr == l2Str {
		t.Fatalf("Mapped list has same typevar as original")
	}

	l = makeListTo(5)
	lStr = l.Type().String() // Force lazy init before copy
	rest := l.Rest()
	restStr := rest.Type().String()
	if lStr == restStr {
		t.Fatalf("Rest has same typevar as original")
	}
}
