
(define (show-string str)
  (display "string: ")
  (display (substring str 0 (min (string-length str) 4)))
  (display "...\n"))

(define str-length (inexact->exact (round (* 2 1024 1024)))) ;; 2MB
(define str1
  (let ((tmp (make-string str-length #\a)))
    (string-set! tmp 0 #\")
    (string-set! tmp (- str-length 1) #\")
    tmp))
(show-string str1)

(define str2 (call-with-input-string str1 read))
(show-string str2)
