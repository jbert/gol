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
		{"(- 1 1)", "0"},
		{"(- 1 2)", "-1"},
		{`(let ((x (- 1 2)))
				x)`, "-1"},
		{`(let ((- +))
			(let ((x (- 1 2)))
				x))`, "3"},
		{`((lambda (x) (+ 1 x)) 1)`, "2"},
		{`((lambda (x y) (+ y x)) 1 3)`, "4"},
		{`(+ (+ 1 2) (+ 2 3))`, "8"},
		{`(let ((f (lambda (x) (+ 1 x))))
			(f (+ 1 2)))`, "4"},
		{"()", "()"},
		{`(progn 1 2 3)`, "3"},
		{`(progn)`, "()"},
		{`(progn 1)`, "1"},
		{`(progn 1 2 (+ 1 2))`, "3"},

		{`"hello world"`, "hello world"},
		{`"hello \" world"`, "hello \" world"},

		{`(let ((x 1)) 3 2 x)`, "1"},
		{`((lambda (x) (+ 1 x) (+ 2 x)) 2)`, "4"},
	}
	//	s := `
	//(func (inc (x))
	//	(+ 1 x))
	//`

CASE:
	for i, tc := range testCases {
		evalStr, err := evaluateProgram(tc.code)
		if err != nil {
			fmt.Printf("%d: err [%s] for code: %s\n", i, err, tc.code)
			continue CASE
		}
		if evalStr != tc.result {
			fmt.Printf("%d@ wrong result [%s] != [%s] for code: %s\n", i, evalStr, tc.result, tc.code)
			continue CASE
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
