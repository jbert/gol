package gol

import "errors"

type Frame map[string]Node

type Environment []Frame

func MakeDefaultEnvironment() Environment {
	defEnv := []Frame{
	//		Frame{"+": addNum},
	}
	return defEnv
}

func (e Environment) WithFrame(f Frame) Environment {
	// 'append on the front'
	// Slow to build, but fast to look up
	newEnv := []Frame{f}
	newEnv = append(newEnv, e...)
	return newEnv
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
