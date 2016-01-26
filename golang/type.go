package golang

import (
	"fmt"
	"strings"

	"github.com/jbert/gol/typ"
)

func golangStringForType(t typ.Type) string {
	switch ty := t.(type) {
	case typ.Primitive:
		return golangStringForPrimitive(ty)
	case typ.Func:
		return golangStringForFunc(ty)
	default:
		panic("Unknown type")
	}
}

func golangStringForPrimitive(p typ.Primitive) string {
	switch p {
	case typ.Unknown:
		panic("Can't get golang string of 'Unknown' type")
	case typ.Num:
		return "int64"
	case typ.Bool:
		return "bool"
	case typ.Symbol:
		return "string"
	case typ.String:
		return "string"
	default:
		panic("Can't get golang string of unrecognised type")
	}
}

func golangStringForFunc(f typ.Func) string {
	args := make([]string, len(f.Args))
	for i := range f.Args {
		args[i] = golangStringForType(f.Args[i])
	}
	result := golangStringForType(f.Result)
	return fmt.Sprintf("func(%s) %s", strings.Join(args, ","), result)
}
