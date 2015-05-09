package flowgraph

import (
	"sort"
	"sync"
	"sync/atomic"
)

type DoubleDatum struct {
	a,b Datum
}

var poolQsortSz int64
var poolQsortMu = &sync.Mutex{}

type RecursiveSort interface {
	sort.Interface

	// SubSlice returns a sub-slice.
	SubSlice(n, m int) Datum

	// Slice returns current slice.
	Slice() []int
	// SliceSorted tests if current slice is sorted.
	SliceSorted() bool

	// Original returns original slice
	Original() []int
	// OriginalSorted tests if original slice is sorted.
	OriginalSorted() bool

	// Depth returns the depth of a recursive sort
	Depth() int64
	// ID returns a unique ID for the object
	ID() int64
}

func (n *Node) deltaPool(delta int) {
	poolSz := atomic.AddInt64(&poolQsortSz, int64(delta))
	n.Tracef("\tpool %s\n", func() string {var s string; for i:=int64(0); i<poolSz; i++ { s += "*" }; return s}())
}

func (n *Node) freeNode (num int) bool {
	poolQsortMu.Lock()
	defer poolQsortMu.Unlock()
	
	d := n.Srcs[0].Val.(RecursiveSort)
	n.Tracef("Original(%p) sorted %t, Sliced sorted %t, poolsz=%d, depth=%d, id=%d, len=%d\n", d.Original(), d.OriginalSorted(), d.SliceSorted(), poolQsortSz, d.Depth(), d.ID(), d.Len())

	var f bool
	if poolQsortSz>int64(num) {
		n.deltaPool(-num)
		f = true
	} else {
		f = false
	}
	return f
}

func qsortFire (n *Node) {

	a := n.Srcs[0]
	steerAck := a.Ack2 != nil
	a.SendAck(n) // write early to let flow go on

	x := n.Dsts[0]
	a.NoOut = true
	if _,ok := a.Val.(RecursiveSort); !ok {
		n.LogError("not of type RecursiveSort (%T)\n", a.Val)
		return
	}

	d := a.Val.(RecursiveSort)
	l := d.Len()

	if l <= 4096 || !n.freeNode(2) {
		sort.Sort(d)
		x.Val=x.AckWrap(d)
		x.SendData(n)
		x.NoOut = true
		if steerAck {
			n.deltaPool(1)
		}
		return
	}

	mlo,mhi := doPivot(d, 0, l)
	var lo,hi Datum
	c := 0
	xData := x.Data
	xName := x.Name
	x.Data = a.Data // recurse
	x.Name = x.Name+"("+a.Name+")"
	if mlo>0 {
		n.Tracef("Original(%p) recurse left [0:%d]\n", d.Original(), mlo)
		lo = x.AckWrap(d.SubSlice(0, mlo))
		x.Val = lo
		x.SendData(n)
		c++
	}
	if l-mhi>0 {
		n.Tracef("Original(%p) recurse right [%d:%d]\n", d.Original(), mhi, l)
		hi = x.AckWrap(d.SubSlice(mhi, l))
		x.Val = hi
		x.SendData(n)
		c++
	}
	x.Data = xData
	x.Name = xName

	x.Val = DoubleDatum{lo, hi}
	x.NoOut = true
	
	if steerAck {
		n.deltaPool(1)
	}
	x.RdyCnt = c
	x.NoOut = true

}

// FuncQsort recursively implements a quicksort with goroutines 
// (x=qsort(a)).
func FuncQsort(a, x Edge, poolSz int) []Node {
	
	// Make a pool of qsort nodes that can be dynamically used, 
	// and reserve one for the front end input into this dynamically 
	// extruded flowgraph.
	n := MakeNodes(poolSz)
	poolQsortSz = int64(poolSz)-1

	for i:=0; i<poolSz; i++ {
		aa, xx := a,x  // make a copy of the Edge's for each one
		n[i] = MakeNodePool("qsort", []*Edge{&aa}, []*Edge{&xx}, 
			nil, qsortFire)
	}
	return n

}
