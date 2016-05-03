package gol

import (
	"fmt"

	"github.com/jbert/gol/typ"
)

type NodeBase struct {
	t typ.Type
}

func (nb *NodeBase) lazyInit() {
	// Zero type is nil. We want a type var
	if nb.t == nil {
		nb.t = typ.NewVar()
	}
}

func (nb *NodeBase) Type() typ.Type {
	nb.lazyInit()
	return nb.t
}

func (nb *NodeBase) NodeUnify(t typ.Type, env typ.Env) error {
	nb.lazyInit()

	//log.Printf("NodeBase Nodeunify (nb type %T)\n", nb)
	return nb.t.Unify(t)
}

type Node interface {
	String() string
	Pos() Position
	Type() typ.Type
	NodeUnify(t typ.Type, env typ.Env) error
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
	return &NodeInt{value: n}
}

func (ni *NodeInt) String() string {
	return fmt.Sprintf("%d", ni.value)
}
func (ni *NodeInt) Value() int64 {
	return ni.value
}

func (ni NodeInt) Type() typ.Type {
	return typ.Int
}

func (ni NodeInt) NodeUnify(t typ.Type, env typ.Env) error {
	return t.Unify(typ.Int)
}

type NodeIdentifier struct {
	nodeAtom
}

func makeIdentifier(s string) *NodeIdentifier {
	return &NodeIdentifier{nodeAtom{tok: Token{
		Type:  tokIdentifier,
		Value: s,
	}}}
}

func (ni NodeIdentifier) NodeUnify(t typ.Type, env typ.Env) error {
	envType, err := env.Lookup(ni.String())
	if err != nil {
		return err
	}
	// Unify our lazy var with our env type
	err = ni.t.Unify(envType)
	if err != nil {
		return err
	}
	// And with the passed in type
	return envType.Unify(t)
}

type NodeSymbol struct {
	nodeAtom
}

func (ns NodeSymbol) Type() typ.Type {
	return typ.Symbol
}

func (ns NodeSymbol) NodeUnify(t typ.Type, env typ.Env) error {
	return t.Unify(typ.Symbol)
}

type NodeString struct {
	nodeAtom
}

func (ns NodeString) Type() typ.Type {
	return typ.String
}

func (ns NodeString) NodeUnify(t typ.Type, env typ.Env) error {
	return t.Unify(typ.String)
}

type NodeBool struct {
	nodeAtom
}

func (nb NodeBool) Type() typ.Type {
	return typ.Bool
}

func (nb NodeBool) NodeUnify(t typ.Type, env typ.Env) error {
	return t.Unify(typ.Bool)
}

var NODE_FALSE = &NodeBool{
	nodeAtom{
		tok: Token{
			Type:  tokBool,
			Value: "#f",
		},
	},
}

var NODE_TRUE = &NodeBool{
	nodeAtom{
		tok: Token{
			Type:  tokBool,
			Value: "#t",
		},
	},
}

var NODE_NIL = &NodeList{}

func (nb *NodeBool) IsTrue() bool {
	return nb.String() == "#t"
}

func (ns *NodeString) String() string {
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
	NodeBase
	Car Node
	Cdr Node
}

func NewNodePair(car, cdr Node) *NodePair {
	pairType := typ.NewPair(car.Type(), cdr.Type())
	np := &NodePair{
		NodeBase: NodeBase{t: pairType},
		Car:      car,
		Cdr:      cdr,
	}
	return np
}

func (np *NodePair) String() string {
	if np.IsNil() {
		return "()"
	} else {
		return fmt.Sprintf("(%v . %v)", np.Car, np.Cdr)
	}
}

func (np *NodePair) Pos() Position {
	return np.Car.Pos()
}

func Nil() *NodePair {
	return &NodePair{}
}

func (np *NodePair) IsNil() bool {
	return np.Car == nil && np.Cdr == nil
}

type NodeList struct {
	NodeBase
	children *NodePair
}

func NewNodeList() *NodeList {
	return &NodeList{
		children: Nil(),
	}
}

func NewNodeListType(t typ.Type) *NodeList {
	nl := NewNodeList()
	nl.t = t
	return nl
}

//func (nl *NodeList) SetType(t typ.Type) error {
//	return nl.NodeBase.SetType(t)
//}

func (nl *NodeList) Pos() Position {
	if nl.Len() == 0 {
		return Position{File: "<empty list>"}
	} else {
		return nl.First().Pos()
	}
}

// ----------------------------------------
// *NodeList sub types

type NodeLambda struct {
	*NodeList
	Args *NodeList
	Body Node
}

type NodeUnQuote struct {
	*NodeList
	Arg Node
}

func (nq *NodeUnQuote) String() string {
	return "," + nq.Arg.String()
}

type NodeQuote struct {
	*NodeList
	Arg   Node
	Quasi bool
}

func (nq *NodeQuote) String() string {
	argStr := nq.Arg.String()
	if nq.Quasi {
		return "'" + argStr
	} else {
		return "`" + argStr
	}

}

type NodeIf struct {
	*NodeList
	Condition Node
	TBranch   Node
	FBranch   Node
}
type NodeSet struct {
	*NodeList
	Id    *NodeIdentifier
	Value Node
}

type NodeProgn struct {
	*NodeList
}

type NodeDefine struct {
	*NodeList
	Symbol Node
	Value  Node
}

type Frame map[string]Node

type NodeLet struct {
	*NodeList
	Bindings Frame
	Body     Node
}

// ----------------------------------------

type NodeError struct {
	Node
	msg string
}

func (ne *NodeError) Type() typ.Type {
	return typ.Void
}
func (ne *NodeError) NodeUnify(t typ.Type, env typ.Env) error {
	return fmt.Errorf("Can't unify an error node type on an error")
}

func (ne *NodeError) String() string {
	return ne.Error()
}
func (ne *NodeError) Error() string {
	pos := ne.Pos()
	return fmt.Sprintf("%s: %s line %d:%d [%s]", ne.msg, pos.File, pos.Line, pos.Column, ne.Node)
}

func NodeErrorf(n Node, f string, args ...interface{}) *NodeError {
	return &NodeError{Node: n, msg: fmt.Sprintf(f, args...)}
}
