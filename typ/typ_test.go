package typ

import (
	"log"
	"testing"
)

func TestFuncUnifyBasic(t *testing.T) {
	vA := NewVar()

	f := Func{
		Args:   []Type{vA},
		Result: vA,
	}
	log.Printf("F has type: %s\n", f)

	//if f.Concrete() {
	//t.Fatalf("F shouldn't be concrete")
	//}

	err := vA.Unify(Int)
	if err != nil {
		t.Fatalf("Can't unify var to int")
	}

	//if !f.Concrete() {
	//t.Fatalf("F should now be concrete")
	//}

	intToInt := Func{
		Args:   []Type{Int},
		Result: Int,
	}

	if f.String() != intToInt.String() {
		t.Fatalf("F string is: %s", f)
	}
}

func TestVarUnifyChain(t *testing.T) {
	makeVars := func(n int) []Type {
		//		debugResetSymbols()
		vars := make([]Type, n)
		for i := range vars {
			vars[i] = NewVar()
		}
		return vars
	}

	unifyOrFail := func(v []Type, i int, j int) {
		log.Printf("Unify %s with %s\n", v[i], v[j])
		err := v[i].Unify(v[j])
		if err != nil {
			t.Fatalf("Can't unify: %s", err)
		}
	}

	checkAll := func(v []Type, s string) {
		for i := range v {
			if v[i].String() != s {
				t.Fatalf("Index [%d] is not a [%s] it's an [%s]", i, s, v[i].String())
			}
		}
		log.Printf("---- All vars are %s\n", s)
	}

	var err error
	var v []Type

	// -------------
	t.Logf("Two chains, join the starts\n")
	v = makeVars(4)
	unifyOrFail(v, 0, 1)
	unifyOrFail(v, 2, 3)
	unifyOrFail(v, 0, 2)

	err = v[0].Unify(Int)
	if err != nil {
		t.Fatal("Can't unify with Int: %s", err)
	}
	checkAll(v, "Int")

	// -------------
	t.Logf("Two chains, join the ends\n")
	v = makeVars(4)
	unifyOrFail(v, 0, 1)
	unifyOrFail(v, 2, 3)
	unifyOrFail(v, 1, 3)

	err = v[0].Unify(Int)
	if err != nil {
		t.Fatal("Can't unify with Int: %s", err)
	}
	checkAll(v, "Int")

	// -------------
	t.Logf("Two chains, join one start to one end\n")
	v = makeVars(4)
	unifyOrFail(v, 0, 1)
	unifyOrFail(v, 2, 3)
	unifyOrFail(v, 0, 3)

	err = v[0].Unify(Int)
	if err != nil {
		t.Fatal("Can't unify with Int: %s", err)
	}
	checkAll(v, "Int")

	// -------------
	t.Logf("One chain, unify  start")
	v = makeVars(4)
	unifyOrFail(v, 0, 1)
	unifyOrFail(v, 1, 2)
	unifyOrFail(v, 2, 3)

	err = v[0].Unify(Int)
	if err != nil {
		t.Fatal("Can't unify with Int: %s", err)
	}
	checkAll(v, "Int")

	// -------------
	t.Logf("One chain, unify  end")
	v = makeVars(4)
	unifyOrFail(v, 0, 1)
	unifyOrFail(v, 1, 2)
	unifyOrFail(v, 2, 3)

	err = v[3].Unify(Int)
	if err != nil {
		t.Fatal("Can't unify with Int: %s", err)
	}
	checkAll(v, "Int")

	// -------------
	t.Logf("One chain, unify  middle")
	v = makeVars(4)
	unifyOrFail(v, 0, 1)
	unifyOrFail(v, 1, 2)
	unifyOrFail(v, 2, 3)

	err = v[2].Unify(Int)
	if err != nil {
		t.Fatal("Can't unify with Int: %s", err)
	}
	checkAll(v, "Int")

	// -------------
	t.Logf("Loop check: One chain, self-unify start")
	v = makeVars(4)
	unifyOrFail(v, 0, 1)
	unifyOrFail(v, 1, 2)
	unifyOrFail(v, 2, 3)
	unifyOrFail(v, 0, 0)

	err = v[2].Unify(Int)
	if err != nil {
		t.Fatal("Can't unify with Int: %s", err)
	}
	checkAll(v, "Int")

	// -------------
	t.Logf("Loop check: One chain, self-unify end")
	v = makeVars(4)
	unifyOrFail(v, 0, 1)
	unifyOrFail(v, 1, 2)
	unifyOrFail(v, 2, 3)
	unifyOrFail(v, 0, 3)

	err = v[2].Unify(Int)
	if err != nil {
		t.Fatal("Can't unify with Int: %s", err)
	}
	checkAll(v, "Int")

	// -------------
	t.Logf("Loop check: One chain, self-unify middle")
	v = makeVars(4)
	unifyOrFail(v, 0, 1)
	unifyOrFail(v, 1, 2)
	unifyOrFail(v, 2, 3)
	unifyOrFail(v, 2, 3)

	err = v[2].Unify(Int)
	if err != nil {
		t.Fatal("Can't unify with Int: %s", err)
	}
	checkAll(v, "Int")
}
