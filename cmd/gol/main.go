package main

import (
	"fmt"
	"os"

	"github.com/jbert/gol"
)

func main() {
	type testCase struct {
		code   string
		result string
	}
	testCases := []testCase{
		{"1", "1"},
		{"2", "2"},
		{"3", "3"},
		{"0", "0"},
		{"-1", "-1"},
		{"+1", "1"},
		{"(+ 1 1)", "2"},
		{"(let ((x 1)) x)", "1"},
	}
	//	s := `
	//(func (inc (x))
	//	(+ 1 x))
	//`

	for i, tc := range testCases {
		evalStr, err := evaluateProgram(tc.code)
		if err != nil {
			fmt.Printf("%d: err [%s] for code: %s\n", i, err, tc.code)
		}
		if evalStr != tc.result {
			fmt.Printf("%d@ wrong result [%s] != [%s] for code: %s\n", i, evalStr, tc.result, tc.code)
		}
		fmt.Printf("%d: AOK!\n", i)
	}
}

func evaluateProgram(prog string) (string, error) {

	fname := "tt.gol"
	l := gol.NewLexer(fname, prog)
	var lexErr error
	go func() {
		lexErr = l.Run()
	}()

	p := gol.NewParser(l.Tokens)
	nodeTree, parseErr := p.Parse()
	if parseErr != nil {
		return "", fmt.Errorf("Error parsing: %s\n", parseErr)
	}

	nodeTree, err := gol.Decorate(nodeTree)
	if err != nil {
		return "", fmt.Errorf("Error decorating: %s\n", err)
	}

	env := gol.MakeDefaultEnvironment()

	//	fmt.Printf("AST: %s\n", nodeTree)
	e := gol.NewEvaluator(os.Stdout, os.Stdin, os.Stderr)
	value, err := e.Eval(nodeTree, env)
	if err != nil {
		return "", fmt.Errorf("Error evaluating: %s\n", err)
	}
	//	fmt.Printf("EVAL: %s\n", value)

	if lexErr != nil {
		return "", fmt.Errorf("Error lexing: %s\n", lexErr)
	}

	return value.String(), nil
}
