package gol

import "testing"

func makeNum(n int) Node {
	return NodeInt{value: int64(n)}
}

func makeListTo(n int) NodeList {
	nl := NodeList{}
	for i := n; i > 0; i-- {
		nl = nl.Cons(makeNum(i))
	}
	return nl
}

func TestListReverse(t *testing.T) {
	p := Nil()
	if !p.IsNil() {
		t.Fatalf("Nil pair isn't nil")
	}

	l1 := NodeList{
		children: NodePair{makeNum(2), NodePair{makeNum(1), NodePair{}}},
	}
	if l1.Len() != 2 {
		t.Fatalf("Wrong length of l1: %d", l1.Len())
	}
	t.Logf("L1 %s\n", l1)

	l2 := l1.Reverse()
	if l2.Len() != 2 {
		t.Fatalf("Wrong length of l2: %d", l1.Len())
	}

	if l2.First().(NodeInt).Value() != 1 {
		t.Fatalf("Wrong reversed item")
	}
}

func TestListCons(t *testing.T) {
	t.Log("TestListCons")

	l := NodeList{}
	if l.Len() != 0 {
		t.Fatalf("Empty list doesn't have zero length")
	}
	l2 := l.Cons(makeNum(1))
	if l.Len() != 0 {
		t.Fatalf("Cons changes the source")
	}
	if l2.Len() != 1 {
		t.Fatalf("Cons doesn't return a longer list")
	}

	n := l2.First()
	ni, ok := n.(NodeInt)
	if !ok {
		t.Fatalf("cons-then-first is the wrong type")
	}
	if ni.Value() != 1 {
		t.Fatalf("cons-then-first is the wrong value")
	}

}

func TestListMap(t *testing.T) {
	l := makeListTo(5)
	t.Logf("initial list: %s\n", l)
	expected := int64(1)

	m, err := l.Map(func(n Node) (Node, error) {
		ni, ok := n.(NodeInt)
		if !ok {
			t.Errorf("Non-int node in list")
			return nil, nil
		}
		if ni.Value() != expected {
			t.Errorf("Wrong value in position %d (%d)", expected, ni.Value())
			return nil, nil
		}
		expected++
		return makeNum(int(-ni.Value())), nil
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

	expected = -1
	m2, _ := m.Map(func(n Node) (Node, error) {
		ni, ok := n.(NodeInt)
		if !ok {
			t.Errorf("Non-int node in list")
			return nil, nil
		}
		if ni.Value() != expected {
			t.Errorf("Wrong value in position %d (%d)", expected, ni.Value())
			return nil, nil
		}
		expected--
		return makeNum(int(-ni.Value())), nil
	})
	t.Logf("second map return: %s\n", m2)
}
