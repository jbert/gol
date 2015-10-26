package gol

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Gol struct {
	eval *Evaluator
}

func New() *Gol {
	g := Gol{}
	return &g
}

func (g *Gol) EvalFile(fname string) (Node, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	return g.EvalReader(fname, f)
}

type ParseError struct {
	error
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("Parse error: %s", pe.error)
}

type EvalError struct {
	error
}

func (ee EvalError) Error() string {
	return fmt.Sprintf("Eval error: %s", ee.error)
}

type LexError struct {
	error
}

func (le LexError) Error() string {
	return fmt.Sprintf("Lex error: %s", le.error)
}

func (g *Gol) EvalReader(srcName string, r io.Reader) (Node, error) {
	env := MakeDefaultEnvironment()
	err := g.loadStandardLib(&env)
	if err != nil {
		return nil, err
	}

	return g.evalReaderWithEnv(srcName, r, &env)
}

func (g *Gol) evalReaderWithEnv(srcName string, r io.Reader, env *Environment) (Node, error) {
	l := NewLexer(srcName, r)

	// Run the lexer until EOF or error
	var lexErr error
	lexDone := make(chan struct{})
	go func() {
		lexErr = l.Run()
		close(lexDone)
	}()

	// Run the parser until the lexer finishes
	p := NewParser(l.Tokens)
	nodeTree, parseErr := p.Parse()
	if parseErr != nil {
		return nil, ParseError{parseErr}
	}

	nodeTree, parseErr = Transform(nodeTree)
	if parseErr != nil {
		return nil, ParseError{parseErr}
	}

	e := NewEvaluator(*env, os.Stdout, os.Stdin, os.Stderr)
	value, err := e.Eval(nodeTree)

	// Hoover up any lexing errors
	<-lexDone
	if lexErr != nil {
		return nil, LexError{lexErr}
	}

	if err != nil {
		switch e := err.(type) {
		case NodeError:
			return nil, e
		default:
			return nil, EvalError{err}
		}
	}
	//	fmt.Printf("EVAL: %s\n", value)

	return value, nil
}

func (g *Gol) loadStandardLib(env *Environment) error {
	r := strings.NewReader(STDLIB)
	_, err := g.evalReaderWithEnv("<stdlib>", r, env)
	return err
}
