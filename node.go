package gol

import (
	"fmt"

	"github.com/jbert/gol/typ"
)

type NodeBase struct {
	t typ.Type
}

func (nb NodeBase) Type(e typ.Env) typ.Type {
	// Catch at runtime for now - when we have some inference
	// then we can return the stored type
	//	panic(fmt.Sprintf("Unimplemented Type: %T", nb))
	return nb.t
}

type Node interface {
	String() string
	Pos() Position
	Type(e typ.Env) typ.Type
}

type nodeAtom struct {
	NodeBase
	tok Token
}

func (na nodeAtom) Pos() Position {
	return na.tok.Pos
}

type NodeInt struct {
	nodeAtom
	value int64
}

func NewNodeInt(n int64) Node {
	return NodeInt{value: n}
}

func (ni NodeInt) String() string {
	return fmt.Sprintf("%d", ni.value)
}
func (ni NodeInt) Value() int64 {
	return ni.value
}
func (ni NodeInt) Type(e typ.Env) typ.Type {
	return typ.Int
}

type NodeIdentifier struct {
	nodeAtom
}

func (ni NodeIdentifier) Type(e typ.Env) typ.Type {
	//	fmt.Printf("Lookup %s\n", ni.String())
	t, err := e.Lookup(ni.String())
	if err != nil {
		return typ.Unknown
	}
	return t
}

func makeIdentifier(s string) NodeIdentifier {
	return NodeIdentifier{nodeAtom{tok: Token{
		Type:  tokIdentifier,
		Value: s,
	}}}
}

type NodeSymbol struct {
	nodeAtom
}

func (ns NodeSymbol) Type(e typ.Env) typ.Type {
	return typ.Symbol
}

type NodeString struct {
	nodeAtom
}

func (ns NodeString) Type(e typ.Env) typ.Type {
	return typ.String
}

type NodeBool struct {
	nodeAtom
}

func (nb NodeBool) Type(e typ.Env) typ.Type {
	return typ.Bool
}

var NODE_FALSE = NodeBool{
	nodeAtom{
		tok: Token{
			Type:  tokBool,
			Value: "#f",
		},
	},
}

var NODE_TRUE = NodeBool{
	nodeAtom{
		tok: Token{
			Type:  tokBool,
			Value: "#t",
		},
	},
}

var NODE_NIL = NodeList{}

func (nb NodeBool) IsTrue() bool {
	return nb.String() == "#t"
}

func (ns NodeString) String() string {
	// Unescape
	value := make([]rune, 0, len(ns.tok.Value))
	escaped := false
RUNE:
	for _, r := range ns.tok.Value {
		if r == '\\' {
			if !escaped {
				escaped = true
				continue RUNE
			} else {
				escaped = false
				// fall through
			}
		}
		if escaped {
			switch r {
			case 'n':
				value = append(value, '\n')
			default:
				value = append(value, r)
			}
		} else {
			value = append(value, r)
		}
	}
	return string(value)
}

func (na nodeAtom) String() string {
	return na.tok.Value
}

// ----------------------------------------

type NodePair struct {
	Car Node
	Cdr Node
}

func (np NodePair) Type(e typ.Env) typ.Type {
	// TODO: return an and-type:
	// - List[car.Type()]			// homogenous list
	// - List[datum]			// heterogenous list
	// - Pair[car.Type(), cdr.Type()]	// Pair
	return typ.NewPair(np.Car.Type(e), np.Cdr.Type(e))
}

func NewNodePair(car, cdr Node) NodePair {
	return NodePair{Car: car, Cdr: cdr}
}

func (np NodePair) String() string {
	if np.IsNil() {
		return "()"
	} else {
		return fmt.Sprintf("(%v %v)", np.Car, np.Cdr)
	}
}

func (np NodePair) Pos() Position {
	return np.Car.Pos()
}

func Nil() NodePair {
	return NodePair{}
}

func (np *NodePair) IsNil() bool {
	return np.Car == nil && np.Cdr == nil
}

type NodeList struct {
	NodeBase
	children NodePair
}

func (nl NodeList) Pos() Position {
	if nl.Len() == 0 {
		return Position{File: "<empty list>"}
	} else {
		return nl.First().Pos()
	}
}

func (nl NodeList) Type(e typ.Env) typ.Type {
	headType := nl.First().Type(e)
	//	fmt.Printf("Type of NL [0] (%s)\n", headType.String())
	f, ok := headType.(typ.Func)
	if !ok {
		// TODO infer!
		panic("Non-function in head position - can't infer type")
	}
	// TODO: validate args
	return f.Result
}

// ----------------------------------------
// NodeList sub types

type NodeLambda struct {
	NodeList
	Args NodeList
	Body Node
}

type NodeUnQuote struct {
	NodeList
	Arg Node
}

func (nq NodeUnQuote) String() string {
	return "," + nq.Arg.String()
}

type NodeQuote struct {
	NodeList
	Arg   Node
	Quasi bool
}

func (nq NodeQuote) String() string {
	argStr := nq.Arg.String()
	if nq.Quasi {
		return "'" + argStr
	} else {
		return "`" + argStr
	}

}

type NodeIf struct {
	NodeList
	Condition Node
	TBranch   Node
	FBranch   Node
}
type NodeSet struct {
	NodeList
	Id    NodeIdentifier
	Value Node
}

type NodeProgn struct {
	NodeList
}

type NodeDefine struct {
	NodeList
	Symbol Node
	Value  Node
}

type Frame map[string]Node

type NodeLet struct {
	NodeList
	Bindings Frame
	Body     Node
}

// ----------------------------------------

type NodeError struct {
	Node
	msg string
}

func (ne NodeError) String() string {
	return ne.Error()
}
func (ne NodeError) Error() string {
	pos := ne.Pos()
	return fmt.Sprintf("%s: %s line %d:%d [%s]", ne.msg, pos.File, pos.Line, pos.Column, ne.Node)
}

func NodeErrorf(n Node, f string, args ...interface{}) NodeError {
	return NodeError{Node: n, msg: fmt.Sprintf(f, args...)}
}
