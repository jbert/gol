package typ

import (
	"fmt"
	"strings"
)

type Primitive int

const (
	Unknown Primitive = 0
	Num     Primitive = iota
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
	case Num:
		return "Num"
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
