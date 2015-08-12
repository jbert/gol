package main

import (
	"fmt"
	"os"

	"github.com/jbert/gol"
)

func main() {
	type testCase struct {
		code      string
		result    string
		errOutput string
	}
	testCases := []testCase{
		{"1", "1", ""},
		{"2", "2", ""},
		{"3", "3", ""},
		{"0", "0", ""},
		{"-1", "-1", ""},
		{"+1", "1", ""},
		{"(+ 1 1)", "2", ""},
		{"(let ((x 1)) x)", "1", ""},
		{"(- 1 1)", "0", ""},
		{"(- 1 2)", "-1", ""},
		{`(let ((x (- 1 2)))
					x)`, "-1", ""},
		{`(let ((- +))
				(let ((x (- 1 2)))
					x))`, "3", ""},
		{`((lambda (x) (+ 1 x)) 1)`, "2", ""},
		{`((lambda (x y) (+ y x)) 1 3)`, "4", ""},
		{`(+ (+ 1 2) (+ 2 3))`, "8", ""},
		{`(let ((f (lambda (x) (+ 1 x))))
				(f (+ 1 2)))`, "4", ""},
		{"()", "", "empty application"},
		{`(progn 1 2 3)`, "3", ""},
		{`(progn)`, "()", ""},
		{`(progn 1)`, "1", ""},
		{`(progn 1 2 (+ 1 2))`, "3", ""},

		{`"hello world"`, "hello world", ""},
		{`"hello \" world"`, "hello \" world", ""},

		{`(let ((x 1)) 3 2 x)`, "1", ""},
		{`((lambda (x) (+ 1 x) (+ 2 x)) 2)`, "4", ""},
		{`(error "time to die")`, "", "time to die"},
		{`(progn "foo" "bar")`, "bar", ""},
		{`(+ (error "foo") 1)`, "", "foo"},
		{`(+ 1 (error "foo"))`, "", "foo"},
		{`(progn (error "foo") "bar")`, "", "foo"},

		{`#t`, "#t", ""},
		{`#f`, "#f", ""},

		{`(if #t 2 3)`, "2", ""},
		{`(if #f 2 3)`, "3", ""},
		{`(if #t 2 (error "no"))`, "2", ""},

		{`(define a 2) a`, "2", ""},
		{`(define a 2) (define a 3) a`, "3", ""},

		{`(define f (lambda (x) (+ 1 x))) (+ 1 3)`, "4", ""},
		{`((lambda (x) (+ 1 x)) 3)`, "4", ""},
		{`(define f (lambda (x) (+ 1 x))) (f 3)`, "4", ""},
	}
	//	s := `
	//(func (inc (x))
	//	(+ 1 x))
	//`

CASE:
	for i, tc := range testCases {
		//fmt.Printf("%d: running: %s\n", i, tc.code)
		evalStr, errStr, err := evaluateProgram(tc.code)
		if err != nil {
			fmt.Printf("%d: err [%s] for code: %s\n", i, err, tc.code)
			continue CASE
		}
		if evalStr != tc.result {
			fmt.Printf("%d@ wrong result [%s] != [%s] for code: %s\n", i, evalStr, tc.result, tc.code)
			continue CASE
		}
		if errStr != tc.errOutput {
			fmt.Printf("%d@ wrong error [%s] != [%s] for code: %s\n", i, errStr, tc.errOutput, tc.code)
			continue CASE
		}
		fmt.Printf("%d: AOK!\n", i)
	}
}

func evaluateProgram(prog string) (string, string, error) {

	fname := "tt.gol"
	l := gol.NewLexer(fname, prog)
	var lexErr error
	lexDone := make(chan struct{})
	go func() {
		lexErr = l.Run()
		close(lexDone)
	}()

	p := gol.NewParser(l.Tokens)
	nodeTree, parseErr := p.Parse()
	if parseErr != nil {
		return "", "", fmt.Errorf("Error parsing: %s\n", parseErr)
	}

	env := gol.MakeDefaultEnvironment()

	//	fmt.Printf("AST: %s\n", nodeTree)
	e := gol.NewEvaluator(os.Stdout, os.Stdin, os.Stderr)
	value, err := e.Eval(nodeTree, env)

	<-lexDone
	if lexErr != nil {
		return "", "", fmt.Errorf("Error lexing: %s\n", lexErr)
	}

	if err != nil {
		switch e := err.(type) {
		case gol.Node:
			return "", e.String(), nil
		default:
			return "", "", fmt.Errorf("Error evaluating: %s\n", err)
		}
	}
	//	fmt.Printf("EVAL: %s\n", value)

	return value.String(), "", nil
}
