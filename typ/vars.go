package typ

import (
	"errors"
	"fmt"
	"log"
)

var nextSymbolName chan string
var resetSymbols chan bool

func init() {
	nextSymbolName = make(chan string)
	resetSymbols = make(chan bool)
	go writeSymbols()
}

func numToAsciiLetter(n int) byte {
	if n == 0 {
		return 'z'
	}
	if n > 25 {
		panic("invalid letter selector")
	}
	return byte(n - 1 + 'a')
}

func numToString(n int) string {
	var buf []byte
	n += 1
	for n > 0 {
		m := n % 26
		front := []byte{numToAsciiLetter(m)}

		buf = append(front, buf...)

		n /= 26
	}
	return string(buf)
}

func debugResetSymbols() {
	resetSymbols <- true
}

func writeSymbols() {
	for count := 0; true; count++ {
		select {
		case nextSymbolName <- numToString(count):
		case <-resetSymbols:
			count = -1
		}
	}
}

func getNextSymbol() string {
	return <-nextSymbolName
}

type Var struct {
	name string
}

func NewVar() *Var {
	return &Var{
		name: getNextSymbol(),
	}
}

var ErrNotFound = errors.New("Not found")

func (v *Var) String() string {
	ty, err := v.Lookup()
	//log.Printf("Lookup returned ptr [%p] and errptr [%p]\n", ty, err)
	if err != nil {
		if err == ErrNotFound {
			return fmt.Sprintf("TV(%s)", v.name)
		} else {
			panic(fmt.Sprintf("Error from type lookup: %s", err))
		}
	} else {
		return ty.String()
	}
}

var symbolResolution map[string]Type = make(map[string]Type)

func (v *Var) Lookup() (Type, error) {
	//log.Printf("lOOKUP: var [%s]\n", v.name)
	found, ok := symbolResolution[v.name]
	//log.Printf("lOOKUP: map has ptr [%v] ok %v\n", found, ok)
	if !ok {
		//log.Printf("LOOKUP: var (%s) not found\n", v.name)
		return v, ErrNotFound
	}

	foundVar, foundIsVar := found.(*Var)
	if foundIsVar {
		// TODO: error on cycles
		//log.Printf("Type of [%s] is var [%s]\n", v.name, foundVar.name)
		return foundVar.Lookup()
	}

	//log.Printf("LOOKUP: found [%s:%v]\n", found, found)
	return found, nil
}

func (v *Var) endOfChain() (Type, error) {
	// Find the end of v's chain (or may be v itself)
	endType, err := v.Lookup()
	if err != nil {
		if err == ErrNotFound {
			// On ErrNotFound, the last var is returned
		} else {
			return nil, err
		}
	}
	return endType, nil
}

func (v *Var) Unify(t Type) error {
	log.Printf("Unifying type %s [%T] with var %s\n", t, t, v.name)

	// Find the end of v's chain
	vEnd, err := v.endOfChain()
	if err != nil {
		return err
	}

	// If t is a var, find the end of t's chain (or may be t itself)
	var tEnd Type
	tEndVar, tEndIsVar := t.(*Var)
	if tEndIsVar {
		tEnd, err = tEndVar.endOfChain()
		if err != nil {
			return err
		}
	} else {
		tEnd = t
	}
	tEndVar, tEndIsVar = tEnd.(*Var)

	vEndVar, vEndIsVar := vEnd.(*Var)
	//log.Printf("vEnd [%s] tEnd [%s]\n", vEnd, tEnd)

	if tEndIsVar && vEndIsVar {
		// It's a variable.
		if tEndVar.name == vEndVar.name {
			// They're the same! nothing to do
		} else {
			// We have two chains of vars. Link them
			//log.Printf("StoreA %s => %s\n", vEndVar.name, tEndVar.name)
			symbolResolution[vEndVar.name] = tEndVar
		}
		return nil
	} else if vEndIsVar {
		if already, ok := symbolResolution[vEndVar.name]; ok {
			panic(fmt.Sprintf("Storing over type [%s] for [%s]", already, vEndVar.name))
		}
		//log.Printf("StoreB %s => %s\n", vEndVar.name, tEnd)
		symbolResolution[vEndVar.name] = tEnd
		return nil
	} else if tEndIsVar {
		if already, ok := symbolResolution[tEndVar.name]; ok {
			panic(fmt.Sprintf("Storing over type [%s] for [%s]", already, tEndVar.name))
		}
		//log.Printf("StoreC %s => %s\n", tEndVar.name, vEnd)
		symbolResolution[tEndVar.name] = vEnd
		return nil

	} else {
		// Neither is a var, so safe to recurse (probable error)
		return vEnd.Unify(tEnd)
	}

}
