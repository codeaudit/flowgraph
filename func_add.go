package flowgraph

import (
	"reflect"
)

func addFire2(a, b Datum) Datum {
	
	switch a.(type) {
        case int8: { return a.(int8)+b.(int8) }
        case uint8: { return a.(uint8)+b.(uint8) }
        case int16: { return a.(int16)+b.(int16) }
        case uint16: { return a.(uint16)+b.(uint16) }
        case int32: { return a.(int32)+b.(int32) }
        case uint32: { return a.(uint32)+b.(uint32) }
        case int64: { return a.(int64)+b.(int64) }
        case uint64: { return a.(uint64)+b.(uint64) }
	case int: { return a.(int)+b.(int) }
	case uint: { return a.(uint)+b.(uint) }
	case float32: { return a.(float32)+b.(float32) }
	case float64: { return a.(float64)+b.(float64) }
	case complex64: { return a.(complex64)+b.(complex64) }
	case complex128: { return a.(complex128)+b.(complex128) }
	case string: { return a.(string)+b.(string) }
	default: { return nil }
	}
}

// Addition primitive
func addFire(n *Node) {

	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]

	atmp,btmp,same := Promote(n, a.Val, b.Val)

	if(!same) {
		n.Errorf("incompatible types for addition (%v+%v)", reflect.TypeOf(a.Val), reflect.TypeOf(b.Val))
		x.Val = nil
	} else {
		x.Val = addFire2(atmp, btmp)
	}
}

// FuncAdd adds values and returns the sum (x = a + b).
func FuncAdd(a, b, x Edge) {

	node := MakeNode("add", []*Edge{&a, &b}, []*Edge{&x}, nil, addFire)
	node.Run()
}

