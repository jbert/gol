package gol

var STDLIB = `
(define (write x) (display x))
(define (newline) (display "\n"))

(define (cons a b)
	(lambda (x)
		(if (= x 1)
		    a
		    b)))

(define (car p)
	(p 1))

(define (cdr p)
	(p 2))

`
