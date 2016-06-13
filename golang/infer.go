package golang

import (
	"fmt"
	"log"

	"github.com/jbert/gol"
	"github.com/jbert/gol/typ"
)

func (gb *GolangBackend) InferTypes() error {
	typeEnv := newDefaultTypeEnv()
	/*
		np, ok := gb.parseTree.(gol.NodePtr)
		if !ok {
			return fmt.Errorf("Internal error: node of type '%T' is not a NodePtr", gb.parseTree)
		}
	*/

	// I'm unsure we'll always converge
	maxLoops := 100

	numLoops := 0

INFERRING:
	for ; numLoops < maxLoops; numLoops++ {
		numUpdates, err := gb.infer(gb.parseTree, typeEnv, 0)
		if err != nil {
			return err
		}
		if numUpdates == 0 {
			break INFERRING
		}
	}
	if numLoops == maxLoops {
		return fmt.Errorf("Infer loop ran for %d iterations", maxLoops)
	}

	fmt.Printf("Program node %p: %s (type is: %s) numLoops %d\n", gb.parseTree, gb.parseTree, gb.parseTree.Type(), numLoops)

	return nil
}

func (gb *GolangBackend) infer(n gol.Node, typeEnv typ.Env, depth int) (int, error) {
	numChanges, err := gb.inferUnwrappedError(n, typeEnv, depth)
	if err != nil {
		_, isNodeErr := err.(*gol.NodeError)
		if !isNodeErr {
			err = gol.NodeErrorf(n, err.Error())
		}
	}
	return numChanges, err
}

func indentString(len int) string {
	buf := make([]byte, len)
	for i := range buf {
		buf[i] = ' '
	}
	return string(buf)
}

func (gb *GolangBackend) inferUnwrappedError(n gol.Node, typeEnv typ.Env, depth int) (int, error) {
	numChanges := 0

	iprintf := func(format string, args ...interface{}) {
		indentStr := indentString(depth)
		s := fmt.Sprintf(format, args...)
		log.Printf("infer: %d %s%s", depth, indentStr, s)
	}

	iprintf("node %p: %s [%T]: %s (%s)\n", n, n, n, n.Type(), typ.ResolveStr(n.Type()))
	// Remember this to see if we changed it
	origType := n.Type()

	switch node := n.(type) {
	case *gol.NodeProgn:
		iprintf("progn\n")
		first := true
		err := node.NodeList.ForeachLast(func(child gol.Node, last bool) error {
			if first {
				// Skip the 'progn' symbol
				first = false
				return nil
			}

			childChanges, err := gb.infer(child, typeEnv, depth+1)
			if err != nil {
				return err
			}
			numChanges += childChanges

			if last {
				// Type of last child should match that of progn as a whole
				iprintf("Inferring type %s for child %s (type %s)\n", n.Type(), child, child.Type())
				err = child.NodeUnify(n.Type(), typeEnv)
				if err != nil {
					return err
				}
			} else {
				// All other children should be void
				iprintf("Inferring void for %s [%T] %s\n", child, child.Type(), child.Type())
				err = child.NodeUnify(typ.Void, typeEnv)
				if err != nil {
					return err
				}
				iprintf("after infer void for %s [%T] %s\n", child, child.Type(), child.Type())
			}
			changed, err := gb.infer(child, typeEnv, depth+1)
			if err != nil {
				return err
			}
			numChanges += changed
			return nil
		})
		if err != nil {
			return 0, err
		}

	case *gol.NodeList:
		iprintf("nodelist\n")
		if node.Len() == 0 {
			return 0, fmt.Errorf("Eval of empty list")
		}

		// What types are in the list?
		var head gol.Node
		argTypes := make([]typ.Type, 0)
		first := true
		node.Foreach(func(child gol.Node) error {
			childChanges, err := gb.infer(child, typeEnv, depth+1)
			if err != nil {
				return err
			}
			numChanges += childChanges

			if first {
				first = false
				head = child
			} else {
				argTypes = append(argTypes, child.Type())
			}
			return nil
		})

		// What type of function would fit these (and return type)?
		wantedType := typ.Func{
			Args:   argTypes,
			Result: node.Type(),
		}

		// Unify that what we have in head position
		err := head.NodeUnify(wantedType, typeEnv)
		if err != nil {
			return 0, err
		}

	case *gol.NodeLambda:
		iprintf("NodeLambda (%s): %s\n", n.String(), typ.ResolveStr(n.Type()))
		argTypes := make([]typ.Type, node.Args.Len())
		i := 0
		frame := make(map[string]typ.Type)
		err := node.Args.Foreach(func(child gol.Node) error {
			id, ok := child.(*gol.NodeIdentifier)
			if !ok {
				return gol.NodeErrorf(n, "non-identifier in lambda args: %s", child.String())
			}

			frame[id.String()] = child.Type()
			argTypes[i] = child.Type()
			i++
			return nil
		})
		if err != nil {
			return 0, err
		}

		oldEnv := typeEnv
		defer func() {
			typeEnv = oldEnv
		}()
		typeEnv = typeEnv.WithFrame(frame)

		childChanges, err := gb.infer(node.Body, typeEnv, depth+1)
		if err != nil {
			return 0, err
		}
		numChanges += childChanges

		resultType := node.Body.Type()

		calcType := typ.NewFunc(argTypes, resultType)
		err = node.NodeUnify(calcType, typeEnv)
		if err != nil {
			return 0, err
		}

	case *gol.NodeIdentifier:
		iprintf("NodeIdentifier (%s)\n", n.String())
		newType, err := typeEnv.Lookup(n.String())
		if err == nil {
			iprintf("NodeIdentifier (%s) [%s]\n", n.String(), newType.String())
			err = node.NodeUnify(newType, typeEnv)
			if err != nil {
				return 0, err
			}
		}

	case *gol.NodeLet:
		iprintf("NodeLet (%s)\n", n.String())

		frame := make(map[string]typ.Type)
		for k, v := range node.Bindings {
			frame[k] = v.Type()
			childChanges, err := gb.infer(v, typeEnv, depth+1)
			if err != nil {
				return 0, err
			}
			numChanges += childChanges
		}

		oldEnv := typeEnv
		defer func() {
			typeEnv = oldEnv
		}()
		typeEnv = typeEnv.WithFrame(frame)

		childChanges, err := gb.infer(node.Body, typeEnv, depth+1)
		if err != nil {
			return 0, err
		}
		numChanges += childChanges

		err = node.NodeUnify(node.Body.Type(), typeEnv)
		if err != nil {
			return 0, err
		}

	case *gol.NodeInt:
		err := node.NodeUnify(typ.Int, typeEnv)
		if err != nil {
			return 0, err
		}
	case *gol.NodeIf:
		iprintf("NodeIf (%s)\n", n.String())
		childChanges, err := gb.infer(node.Condition, typeEnv, depth+1)
		if err != nil {
			return 0, err
		}
		numChanges += childChanges
		childChanges, err = gb.infer(node.TBranch, typeEnv, depth+1)
		if err != nil {
			return 0, err
		}
		numChanges += childChanges
		childChanges, err = gb.infer(node.FBranch, typeEnv, depth+1)
		if err != nil {
			return 0, err
		}
		numChanges += childChanges

		err = node.Condition.NodeUnify(typ.Bool, typeEnv)
		if err != nil {
			return 0, err
		}
		err = node.TBranch.NodeUnify(node.Type(), typeEnv)
		if err != nil {
			return 0, err
		}
		err = node.FBranch.NodeUnify(node.Type(), typeEnv)
		if err != nil {
			return 0, err
		}

	case *gol.NodeError:

	case *gol.NodeSymbol:
	case *gol.NodeString:
	case *gol.NodeBool:

	case *gol.NodeDefine:
		// JB - hack into top level
		typeEnv.AddTopLevel(node.Symbol.String(), node.Value.Type())
		iprintf("NodeDefine (%s)\n", n.String())

		childChanges, err := gb.infer(node.Value, typeEnv, depth+1)
		if err != nil {
			return 0, err
		}
		numChanges += childChanges

	case *gol.NodePair:
		// TODO: use an And type here....
		// Can leave newType as Any since all Pairs have an inferred type

	case *gol.NodeQuote:
		panic("implement")
	case *gol.NodeUnQuote:
		panic("implement")
	case *gol.NodeSet:
		panic("implement")

	default:
		return 0, gol.NodeErrorf(n, "unrecognised/unhandled node type %T", n)

	}

	if n.Type() != origType {
		numChanges++
	}

	iprintf("done\n")

	return numChanges, nil
}
