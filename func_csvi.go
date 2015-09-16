package flowgraph

import (		
	"encoding/csv"
	"io"
	"os"
)

type csvState struct {
	csvreader *csv.Reader
	header []string
	record []string
}

func find(s string, v []string) int {
	for i := range v {
		if v[i]==s {
			return i
		}
	}
	return -1
}

func csviRdy (n *Node) bool {
	if n.Aux == nil { return false }
	
	if n.DefaultRdyFunc() {
		r := n.Aux.(csvState).csvreader
		h := n.Aux.(csvState).header
		record,err := r.Read()
		if err == io.EOF {
			n.Aux = nil
			return false
		} else {
			check(err)
			n.Aux = csvState{r, h, record}
			return true
		}
	}
	return false
}

func csviFire (n *Node) {	 
	x := n.Dsts

	// process data record
	record := n.Aux.(csvState).record
	header := n.Aux.(csvState).header
	l := len(x)
	if l>len(record) { l = len(record) }
	for i:=0; i<l; i++ {
		j := find(x[i].Name, header)
		if j>=0 {
			if record[j]!="*" {
				v := ParseDatum(record[j])
				x[i].Val = v	
			} else {
				x[i].NoOut = true
			}
		} else {
			n.LogError("Named input missing from .csv file:  %s\n", x[i].Name)
			os.Exit(1)
		}
	}
}

// FuncCSVI reads a vector of input data values from a Reader.
// 
func FuncCSVI(x []Edge, r io.Reader) Node {

	var xp []*Edge
	for i := range x {
		xp = append(xp, &x[i])
	}

	node := MakeNode("csvi", nil, xp, csviRdy, csviFire)
	r2 := csv.NewReader(r)

	// save headers
	headers, err := r2.Read()
	check(err)
	node.Aux = csvState{csvreader:r2, header:headers}

	return node
	
}
	