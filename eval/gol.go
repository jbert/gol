package eval

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jbert/gol"
)

type Gol struct {
	eval *Evaluator
}

func New() *Gol {
	g := Gol{}
	return &g
}

func (g *Gol) EvalFile(fname string) (gol.Node, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	return g.EvalReader(fname, f)
}

type EvalError struct {
	error
}

func (ee EvalError) Error() string {
	return fmt.Sprintf("Eval error: %s", ee.error)
}

func (g *Gol) EvalProgram(srcName string, prog string) (gol.Node, error) {
	r := strings.NewReader(prog)
	return g.EvalReader(srcName, r)
}

func (g *Gol) EvalReader(srcName string, r io.Reader) (gol.Node, error) {
	env := MakeDefaultEnvironment()
	err := g.loadStandardLib(&env)
	if err != nil {
		return nil, err
	}

	return g.evalReaderWithEnv(srcName, r, &env)
}

func (g *Gol) evalReaderWithEnv(srcName string, r io.Reader, env *Environment) (gol.Node, error) {
	l := gol.NewLexer(srcName, r)

	// Run the lexer until EOF or error
	var lexErr error
	lexDone := make(chan struct{})
	go func() {
		lexErr = l.Run()
		close(lexDone)
	}()

	// Run the parser until the lexer finishes
	p := gol.NewParser(l.Tokens)
	nodeTree, parseErr := p.Parse()
	if parseErr != nil {
		return nil, parseErr
	}

	nodeTree, parseErr = gol.Transform(nodeTree)
	if parseErr != nil {
		return nil, parseErr
	}

	e := NewEvaluator(*env, os.Stdout, os.Stdin, os.Stderr)
	value, err := e.Eval(nodeTree)

	// Hoover up any lexing errors
	<-lexDone
	if lexErr != nil {
		return nil, lexErr
	}

	if err != nil {
		switch e := err.(type) {
		case *gol.NodeError:
			return nil, e
		default:
			return nil, EvalError{err}
		}
	}
	//	fmt.Printf("EVAL: %s\n", value)

	return value, nil
}

func (g *Gol) loadStandardLib(env *Environment) error {
	r := strings.NewReader(gol.STDLIB)
	_, err := g.evalReaderWithEnv("<stdlib>", r, env)
	return err
}
