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
	a, parseErr := p.Parse()
	if parseErr != nil {
		fmt.Printf("Error parsing: %s\n", parseErr)
		os.Exit(-1)
	}

	fmt.Println(a)
	e := gol.NewEvaluator(a, os.Stdout, os.Stdin, os.Stderr)
	err := e.Run()

	if lexErr != nil {
		fmt.Printf("Error lexing: %s\n", lexErr)
		os.Exit(-1)
	}
}
