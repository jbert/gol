package gol

import "errors"

type Frame map[string]Node

type Environment []Frame

func makeDefaultEnvironment() Environment {
	defEnv := []Frame{
	//		Frame{"+": addNum},
	}
	return defEnv
}

var ErrNotFound = errors.New("Identifier not found")

func (e Environment) Lookup(s string) (Node, error) {
	for _, f := range []Frame(e) {
		node, ok := f[s]
		if ok {
			return node, nil
		}
	}
	return nil, ErrNotFound
}
