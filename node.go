package gol

import "fmt"

type NodeBase struct {
}

type Node interface {
	String() string
	Pos() Position
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

func (nn NodeInt) String() string {
	return fmt.Sprintf("%d", nn.value)
}
func (nn NodeInt) Value() int64 {
	return nn.value
}

type NodeIdentifier struct {
	nodeAtom
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
type NodeString struct {
	nodeAtom
}
type NodeBool struct {
	nodeAtom
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
