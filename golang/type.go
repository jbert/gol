package golang

import (
	"fmt"
	"strings"

	"github.com/jbert/gol/typ"
)

func golangStringForType(t typ.Type) (string, error) {
	switch ty := t.(type) {
	case typ.Primitive:
		return golangStringForPrimitive(ty)
	case typ.Func:
		return golangStringForFunc(ty)
	case typ.Variadic:
		return golangStringForVariadic(ty)
	case *typ.Var:
		tyVal, err := ty.Lookup()
		if err != nil {
			return "", err
		}
		return golangStringForType(tyVal)
	default:
		return "", fmt.Errorf("Can't get golang string for unknown type: %v", t)
	}
}

func golangStringForPrimitive(p typ.Primitive) (string, error) {
	switch p {
	case typ.Any:
		return "interface{}", nil
	case typ.Int:
		return "int64", nil
	case typ.Bool:
		return "bool", nil
	case typ.Symbol:
		return "string", nil
	case typ.String:
		return "string", nil
	default:
		return "", fmt.Errorf("Can't get golang string of unknown primitive type: %s", p)
	}
}

func golangStringForFunc(f typ.Func) (string, error) {
	var err error
	args := make([]string, len(f.Args))
	for i := range f.Args {
		args[i], err = golangStringForType(f.Args[i])
		if err != nil {
			return "", err
		}
	}
	result, err := golangStringForType(f.Result)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("func(%s) %s", strings.Join(args, ","), result), nil
}

func golangStringForVariadic(v typ.Variadic) (string, error) {
	s, err := golangStringForType(v.X)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("...%s", s), nil
}
