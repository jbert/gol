package typ

import (
	"fmt"
	"strings"
)

type Primitive int

const (
	Any Primitive = 0
	Int Primitive = iota
	Bool
	Symbol
	String
	Void
)

type Type interface {
	String() string
	Unify(t Type) error
}

func (p Primitive) String() string {
	switch p {
	case Any:
		return "Any"
	case Int:
		return "Int"
	case Bool:
		return "Bool"
	case Symbol:
		return "Symbol"
	case String:
		return "String"
	case Void:
		return "Void"
	default:
		panic("Unrecognised primitive")
	}
}

func (p Primitive) Unify(t Type) error {
	if tPrim, ok := t.(Primitive); ok && p == tPrim {
		// Both same primitive type
		return nil
	}
	return unifyWithVarOrError(p, t)
}

func unifyWithVarOrError(lh Type, rh Type) error {
	rhVar, ok := rh.(*Var)
	if ok {
		return rhVar.Unify(lh)
	}
	return fmt.Errorf("Can't unify: %T with %T", lh, rh)
}

type Func struct {
	Args   []Type
	Result Type
}

func (f Func) String() string {
	args := make([]string, len(f.Args))
	for i := range f.Args {
		args[i] = f.Args[i].String()
	}
	result := f.Result.String()
	return fmt.Sprintf("(%s) -> %s", strings.Join(args, ","), result)
}

func variadicUnify(a, b []Type) bool {
	if len(b) < len(a) {
		a, b = b, a
	}

	for i := 0; i < len(a)-1; i++ {
		err := a[i].Unify(b[i])
		if err != nil {
			return false
		}
	}
	// Last in a must b variadic, all in b same type
	variadic, ok := a[len(a)-1].(Variadic)
	if !ok {
		return false
	}
	for j := len(a); j < len(b); j++ {
		err := b[j].Unify(variadic.X)
		if err != nil {
			return false
		}
	}

	return true
}

func (f Func) Unify(t Type) error {
	newFunc, ok := t.(Func)
	if !ok {
		return unifyWithVarOrError(f, t)
	}

	err := f.Result.Unify(newFunc.Result)
	if err != nil {
		return err
	}

	if variadicUnify(f.Args, newFunc.Args) {
		return nil
	}

	if len(f.Args) != len(newFunc.Args) {
		return fmt.Errorf("Can't unify: arg count mismatch %d != %d",
			len(f.Args), len(newFunc.Args))
	}
	for i := range f.Args {
		err := f.Args[i].Unify(newFunc.Args[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func NewFunc(args []Type, result Type) Func {
	return Func{Args: args, Result: result}
}

type Pair struct {
	car Type
	cdr Type
}

func NewPair(car Type, cdr Type) Pair {
	return Pair{car: car, cdr: cdr}
}

func (p Pair) Unify(t Type) error {
	switch ty := t.(type) {
	case Pair:
		err := ty.car.Unify(p.car)
		if err != nil {
			return err
		}
		err = ty.cdr.Unify(p.cdr)
		if err != nil {
			return err
		}
		return nil
	case *Var:
		return ty.Unify(ty)
	default:
		return fmt.Errorf("Can't unify: Pair with %T", t)
	}
}

func (p Pair) String() string {
	return fmt.Sprintf("Pair{%s,%s}", p.car.String(), p.cdr.String())
}

type Variadic struct {
	X Type
}

func (v Variadic) String() string {
	return fmt.Sprintf("Variadic{%s}", v.X)
}

func (v Variadic) Unify(t Type) error {
	newVariadic, ok := t.(Variadic)
	if ok {
		return v.X.Unify(newVariadic.X)
	}
	return unifyWithVarOrError(v, t)
}

func NewVariadic(t Type) Variadic {
	return Variadic{t}
}
