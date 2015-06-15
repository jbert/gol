package main

import (
	"fmt"
	"os"

	"github.com/jbert/gol"
)

func main() {
	fname := "tt.gol"
	s := `
(func (inc (x))
	(+ 1 x))
`

	l := gol.NewLexer(fname, s)
	var lexErr error
	go func() {
		lexErr = l.Run()
	}()

	p := gol.NewParser(l.Tokens)
	nodeTree, parseErr := p.Parse()
	if parseErr != nil {
		fmt.Printf("Error parsing: %s\n", parseErr)
		os.Exit(-1)
	}

	fmt.Printf("AST: %s\n", nodeTree)
	e := gol.NewEvaluator(os.Stdout, os.Stdin, os.Stderr)
	value, err := e.Eval(nodeTree)
	if err != nil {
		fmt.Printf("Error evaluating: %s\n", err)
		os.Exit(-1)
	}
	fmt.Printf("EVAL: %s\n", value)

	if lexErr != nil {
		fmt.Printf("Error lexing: %s\n", lexErr)
		os.Exit(-1)
	}
}
