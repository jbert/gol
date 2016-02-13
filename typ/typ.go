package typ

import (
	"fmt"
	"strings"
)

type Primitive int

const (
	Unknown Primitive = 0
	Int     Primitive = iota
	Bool
	Symbol
	String
)

type Type interface {
	String() string
}

func (p Primitive) String() string {
	switch p {
	case Unknown:
		return "Unknown"
	case Int:
		return "Int"
	case Bool:
		return "Bool"
	case Symbol:
		return "Symbol"
	case String:
		return "String"
	default:
		panic("Unrecognised primitive")
	}
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

func (p Pair) String() string {
	return fmt.Sprintf("Pair{%s,%s}", p.car.String(), p.cdr.String())
}

type Variadic struct {
	X Type
}

func (v Variadic) String() string {
	return fmt.Sprintf("Variadic{%s}", v.X)
}

func NewVariadic(t Type) Variadic {
	return Variadic{t}
}
