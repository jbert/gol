package eval

import (
	"fmt"

	"github.com/jbert/gol"
)

type Environment []gol.Frame

func (e Environment) WithFrame(f gol.Frame) Environment {
	// 'append on the front'
	// Slow to build, but fast to look up
	newEnv := []gol.Frame{f}
	newEnv = append(newEnv, e...)
	return newEnv
}

func (e Environment) Lookup(s string) (gol.Node, error) {
	for _, f := range []gol.Frame(e) {
		node, ok := f[s]
		if ok {
			return node, nil
		}
	}
	return nil, fmt.Errorf("Identifier [%s] not found", s)
}

func (e Environment) AddDefine(id string, value gol.Node) error {
	// Add to top-level frame (at end)
	topLevel := e[len(e)-1]
	topLevel[id] = value
	return nil
}

func (e Environment) Set(id string, value gol.Node) (gol.Node, error) {
	for _, f := range []gol.Frame(e) {
		_, ok := f[id]
		if ok {
			f[id] = value
			return value, nil
		}
	}
	return nil, fmt.Errorf("set - Identifier [%s] not found", id)
}

type NodeProcedure struct {
	*gol.NodeLambda
	Env Environment
}
