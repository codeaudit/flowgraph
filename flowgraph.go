// Package flowgraph layers a ready-send flow mechanism on top of goroutines.
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"log"
	"os"
)

// ack channel wrapper
type ackWrap struct {
	ack chan bool
	d Datum
}

// Log for tracing flowgraph execution.
var StdoutLog = log.New(os.Stdout, "", 0)

// Log for collecting error messages.
var StderrLog = log.New(os.Stderr, "", 0)

// Compile global flowgraph stats.
var GlobalStats = false

// Trace level constants.
const (
	Q = iota  // quiet
	V         // trace Node execution
	VV        // trace channel IO
	VVV       // trace state before select
	VVVV      // full-length array dumps
)

// Enable tracing, writes to StdoutLog if TraceLevel>Q.
var TraceLevel = Q

// Indent trace by Node id tabs.
var TraceIndent = false

// Trace timestamp format
var TraceFireCnt = true
var TraceSeconds = false
var TracePointer = false

// Unique Node id.
var NodeID int64

// Global count of number of Node executions.
var globalFireCnt int64

// MakeGraph returns a slice of Edge and a slice of Node.
func MakeGraph(sze, szn int) ([]Edge,[]Node) {
	return MakeEdges(sze),MakeNodes(szn)
}
