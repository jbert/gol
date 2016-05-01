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
		numUpdates, err := gb.infer(gb.parseTree, typeEnv)
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

func (gb *GolangBackend) infer(n gol.Node, typeEnv typ.Env) (int, error) {
	numChanges := 0

	log.Printf("Inferring node %p: %s [%T]: %s\n", n, n, n, n.Type())
	// Remember this to see if we changed it
	origType := n.Type()

	switch node := n.(type) {
	case *gol.NodeProgn:
		//log.Printf("infer: progn\n")
		first := true
		err := node.NodeList.ForeachLast(func(child gol.Node, last bool) error {
			if first {
				// Skip the 'progn' symbol
				first = false
				return nil
			}

			childChanges, err := gb.infer(child, typeEnv)
			if err != nil {
				return err
			}
			numChanges += childChanges

			if last {
				// Type of last child should match that of progn as a whole
				//log.Printf("Inferring type %s for child %s (type %s)\n", n.Type(), child, child.Type())
				child.NodeUnify(n.Type(), typeEnv)
			} else {
				// All other children should be void
				//log.Printf("Inferring void for %s\n", child)
				child.NodeUnify(typ.Void, typeEnv)
			}
			changed, err := gb.infer(child, typeEnv)
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
		log.Printf("infer: nodelist\n")
		if node.Len() == 0 {
			return 0, fmt.Errorf("Eval of empty list")
		}

		// What types are in the list?
		var head gol.Node
		argTypes := make([]typ.Type, 0)
		first := true
		node.Foreach(func(child gol.Node) error {
			childChanges, err := gb.infer(child, typeEnv)
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
		log.Printf("infer: NodeLambda (%s)\n", n.String())
		argTypes := make([]typ.Type, node.Args.Len())
		i := 0
		node.Args.Foreach(func(child gol.Node) error {
			argTypes[i] = child.Type()
			i++
			return nil
		})

		childChanges, err := gb.infer(node.Body, typeEnv)
		if err != nil {
			return 0, err
		}
		numChanges += childChanges

		resultType := node.Body.Type()

		node.NodeUnify(typ.NewFunc(argTypes, resultType), typeEnv)

	case *gol.NodeIdentifier:
		log.Printf("infer: NodeIdentifier (%s)\n", n.String())
		newType, err := typeEnv.Lookup(n.String())
		if err == nil {
			log.Printf("infer: NodeIdentifier (%s) [%s]\n", n.String(), newType.String())
			node.NodeUnify(newType, typeEnv)
		}

	case *gol.NodeLet:
		log.Printf("infer: NodeLet (%s)\n", n.String())

		frame := make(map[string]typ.Type)
		for k, v := range node.Bindings {
			frame[k] = v.Type()
			childChanges, err := gb.infer(v, typeEnv)
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

		childChanges, err := gb.infer(node.Body, typeEnv)
		if err != nil {
			return 0, err
		}
		numChanges += childChanges

		node.NodeUnify(node.Body.Type(), typeEnv)

	case *gol.NodeInt:
		node.NodeUnify(typ.Int, typeEnv)

	case *gol.NodeError:
	case *gol.NodeSymbol:
	case *gol.NodeString:
	case *gol.NodeBool:

	case *gol.NodePair:
		// TODO: use an And type here....
		// Can leave newType as Any since all Pairs have an inferred type

	case *gol.NodeQuote:
		panic("implement")
	case *gol.NodeUnQuote:
		panic("implement")
	case *gol.NodeIf:
		panic("implement")
	case *gol.NodeSet:
		panic("implement")
	case *gol.NodeDefine:
		panic("implement")

	default:
		return 0, gol.NodeErrorf(n, "infer: unrecognised/unhandled node type %T", n)

	}

	if n.Type() != origType {
		numChanges++
	}

	return numChanges, nil
}
