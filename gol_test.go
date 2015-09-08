package gol

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestBasic(t *testing.T) {
	type testCase struct {
		code      string
		result    string
		errOutput string
	}
	testCases := []testCase{
		{"1", "1", ""},
		{`1
			`, "1", ""},
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

		{`(define (f) 2) (f)`, "2", ""},
		{`(define (f x) (+ 1 x)) (f 3)`, "4", ""},
		{`(define (f x) 1) (f 3)`, "1", ""},

		{`(define (fact x) 6) (fact 3)
				  `, "6", ""},

		{`
				(define (fact-helper x res)
				  (if (= x 0)
				      res
				      (fact-helper (- x 1) (* res x))))

				(define (fact x)
				  (fact-helper x 1))

				(fact 3)
				  `, "6", ""},
		{`(display "hello, world\n")`, "()", ""},
	}
	//	s := `
	//(func (inc (x))
	//	(+ 1 x))
	//`

CASE:
	for i, tc := range testCases {
		//	fmt.Printf("%d: running: %s\n", i, tc.code)
		evalStr, errStr, err := evaluateProgram(tc.code)
		if err != nil {
			t.Errorf("%d: err [%s] for code: %s\n", i, err, tc.code)
			continue CASE
		}
		if !strings.HasPrefix(errStr, tc.errOutput) {
			t.Errorf("%d@ wrong error [%s] != [%s] for code: %s\n", i, errStr, tc.errOutput, tc.code)
			continue CASE
		}
		if evalStr != tc.result {
			t.Errorf("%d@ wrong result [%s] != [%s] for code: %s\n", i, evalStr, tc.result, tc.code)
			continue CASE
		}
		t.Logf("%d: AOK!\n", i)
	}
}

func evaluateProgram(prog string) (string, string, error) {

	fname := "<internal>"
	r := strings.NewReader(prog)
	l := NewLexer(fname, r)
	var lexErr error
	lexDone := make(chan struct{})
	go func() {
		lexErr = l.Run()
		close(lexDone)
	}()

	p := NewParser(l.Tokens)
	nodeTree, parseErr := p.Parse()
	if parseErr != nil {
		return "", "", fmt.Errorf("Error parsing: %s\n", parseErr)
	}

	env := MakeDefaultEnvironment()

	//	fmt.Printf("AST: %s\n", nodeTree)
	e := NewEvaluator(os.Stdout, os.Stdin, os.Stderr)
	value, err := e.Eval(nodeTree, env)

	<-lexDone
	if lexErr != nil {
		return "", "", fmt.Errorf("Error lexing: %s\n", lexErr)
	}

	if err != nil {
		switch e := err.(type) {
		case Node:
			return "", e.String(), nil
		default:
			return "", "", fmt.Errorf("Error evaluating: %s\n", err)
		}
	}
	//	fmt.Printf("EVAL: %s\n", value)

	return value.String(), "", nil
}
