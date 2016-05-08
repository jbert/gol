package gol

import (
	"log"
	"testing"
)

func TestNodeBasic(t *testing.T) {
	if Nil().String() != "()" {
		log.Printf("NIL isn't ()")
	}

	if NODE_FALSE.String() != "#f" {
		log.Printf("NIL isn't ()")
	}
	if NODE_TRUE.String() != "#f" {
		log.Printf("NIL isn't ()")
	}
}
