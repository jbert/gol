package test

type TestCase struct {
	Code      string
	Result    string
	ErrOutput string
}

func FuncTestCases() []TestCase {
	return []TestCase{
		{`
(define (f x)
	(+ 1 x)
	(+ 2 x)
	(+ 3 x))

(f 7)
`, "10", ""},
	}

}

func QuoteTestCases() []TestCase {
	return []TestCase{
		{"'1", "1", ""},
		{"'()", "()", ""},
		{"'(+ 1 2)", "(+ 1 2)", ""},

		{"`1", "1", ""},
		{"`()", "()", ""},
		{"`(+ 1 2)", "(+ 1 2)", ""},

		{"`,(+ 1 2)", "3", ""},

		{"`(+ ,(+ 2 3) ,(+ 3 4))", "(+ 5 7)", ""},

		{"'(unquote (+ 1 2))", "(unquote (+ 1 2))", ""},

		{"(quote 1)", "1", ""},
		{"(quote ())", "()", ""},
		{"(quote (+ 1 2))", "(+ 1 2)", ""},

		{"(quasiquote 1)", "1", ""},
		{"(quasiquote ())", "()", ""},
		{"(quasiquote (+ 1 2))", "(+ 1 2)", ""},

		{"(quasiquote (unquote (+ 1 2)))", "3", ""},

		{"(quote (unquote (+ 1 2)))", ",(+ 1 2)", ""},

		{"(list 1 2 3)", "(1 2 3)", ""},
		{"(list (+ 1 1) 2 3)", "(2 2 3)", ""},
	}
}

func ErrorTestCases() []TestCase {
	return []TestCase{
		{"()", "", "empty application"},
		{`(error "time to die")`, "", "time to die"},
		{`(+ (error "foo") 1)`, "", "foo"},
		{`(+ 1 (error "foo"))`, "", "foo"},
		{`(progn (error "foo") "bar")`, "", "foo"},
	}
}

func TypeTestCases() []TestCase {
	return []TestCase{
		{`(+ "foo" 1)`, "", "error - what kind?"},

		{`"foo"`, `"foo"`, ""},
		{`(+ 1 2)`, `3`, ""},
		{`(number->string (+ 1 2))`, `"3"`, ""},

		{`(string-concat "foo" "bar")`, `"foobar"`, ""},
		{`(string-concat (number->string 1) (number->string 2))`, `"12"`, ""},

		{`(let ((f (lambda (x) (+ 1 x))))
			 (f (+ 1 2)))`, "4", ""},
		{`(let ((f (lambda (x) (+ 1 x))))
			 (f "foo"))`, "", "error - what kind?"},
	}
}

func BasicTestCases() []TestCase {
	return []TestCase{
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
		{`(progn 1 2 3)`, "3", ""},
		{`(progn)`, "()", ""},
		{`(progn 1)`, "1", ""},
		{`(progn 1 2 (+ 1 2))`, "3", ""},

		{`"hello world"`, "hello world", ""},
		{`"hello \" world"`, "hello \" world", ""},

		{`(let ((x 1)) 3 2 x)`, "1", ""},
		{`((lambda (x) (+ 1 x) (+ 2 x)) 2)`, "4", ""},
		{`(progn "foo" "bar")`, "bar", ""},
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
		{`((lambda () 2))`, "2", ""},
		{`(define f (lambda () 2)) (f)`, "2", ""},
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
}
