package regexp

import (
	"github.com/vectaport/flowgraph"
)

type repeatStruct struct {
	entries map[string]*repeatEntry
	min, max int
}

type repeatEntry struct {
	prev string
	cnt int
}

type repeatMap map[string]repeatEntry

func repeatFire (n *flowgraph.Node) {

	newmatch := n.Srcs[0]
	subsrc := n.Srcs[1]
	dnstreq := n.Srcs[2]

	oldmatch := n.Dsts[0]
	subdst := n.Dsts[1]
	upstreq := n.Dsts[2]

	st := n.Aux.(repeatStruct)
	rmap := st.entries
	rmin := st.min
	// rmax := st.max

	// flow from downstream
	if dnstreq.Flow  { // set in repeatRdy

		newmatch.Flow = false
		subsrc.Flow = false

		// match >0
		match := dnstreq.SrcGet().(Search)
		if match.State==Done {
			delete(rmap, match.Orig)
			upstreq.DstPut(match)
		} else {
			if rmap[match.Orig]==nil {
				n.Tracef("panic:  nil return from rmap for \"%+v\"  (%v)\n", match, rmap)
				panic("nil return from rmap")
			}
			
			p := rmap[match.Orig].prev
			if len(p)==0 {
				n.Tracef("panic:  unable to send new string downstream\n");
				panic("unable to send new string downstream")
			}
			match.Curr = p[1:]
			subdst.DstPut(match)
		}
		
		return
	}

	// flow from subordinate regexp
	if subsrc.Flow {  // set in repeatRdy()
	   	

		newmatch.Flow = false
		dnstreq.Flow = false
		
		match := subsrc.SrcGet().(Search)

		rs := rmap[match.Orig]
		if match.State == Live {
			rs.prev = match.Curr
			rs.cnt++
			rmap[match.Orig] = rs

			// if not enough yet, match the next 
			if rs.cnt < rmin {
			        match.Curr = match.Curr[1:]
				subdst.DstPut(match)  
				return
			}

		}

		// deal with a submatch not working
		if len(match.Curr)>1 {
			match.Curr = match.Curr[1:]
			subdst.DstPut(match)
			return
		}

		// match failed
		oldmatch.DstPut(match)
		return
		
	}

	// incoming data flow
	if newmatch.Flow { // set in repeatRdy()

		subsrc.Flow = false
		dnstreq.Flow = false
		
		match := newmatch.SrcGet().(Search)
		rs := rmap[match.Orig]
		if rs==nil {
			rs = &repeatEntry{}
			rmap[match.Orig] = rs
		}
		rs.prev = match.Curr
		rmap[match.Orig] = rs

		// if no matches are required, pass it on
		if st.min==0 {
			oldmatch.DstPut(match)
			return
		}

		// otherwise attempt a match
		subdst.DstPut(match)
		return


	}
	

}

func repeatRdy (n *flowgraph.Node) bool {
	if !n.Dsts[0].DstRdy(n) || !n.Dsts[1].DstRdy(n) || !n.Dsts[2].DstRdy(n) { return false }
	rdy := false
	for i := range n.Srcs {
		n.Srcs[i].Flow = n.Srcs[i].SrcRdy(n)
		rdy = rdy || n.Srcs[i].Flow
	}
	return rdy
}

// FuncRepeat repeats a match zero or more times
//
// inputs:
// newmatch -- new match string
// subsrc   -- fedback result of last match, successful (remainder string) or not (nil)
// dnstreq  -- receive downstream request for new remainder string
//
// outputs:
// oldmatch -- continue match (remainder string)
// subdst   -- match done, successful (remainder string) or not (nil)
// upstreq  -- send upstream request for new remainder string
func FuncRepeat(newmatch, subsrc, dnstreq flowgraph.Edge, oldmatch, subdst, upstreq flowgraph.Edge, min, max int) flowgraph.Node {

	node := flowgraph.MakeNode("repeat", []*flowgraph.Edge{&newmatch, &subsrc, &dnstreq}, []*flowgraph.Edge{&oldmatch, &subdst, &upstreq}, repeatRdy, repeatFire)
	node.Aux = repeatStruct{entries:make(map[string]*repeatEntry), min:min, max:max}
	return node

}
