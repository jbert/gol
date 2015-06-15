package gol

import "io"

type Evaluator struct {
	in  io.Reader
	out io.Writer
	err io.Writer
	env Environment
}

type Frame map[string]Node

type Environment []Frame

func NewEvaluator(node Node, out io.Writer, in io.Reader, err io.Writer) *Evaluator {
	return &Evaluator{
		in:  in,
		out: out,
		err: err,
		env: makeDefaultEnvironment(),
	}
}

func makeDefaultEnvironment() Environment {
	defEnv := []Frame{}
	return defEnv
}
