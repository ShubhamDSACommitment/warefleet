package gridworld

import "container/heap"

// Reservation is the space-time table of cells claimed by already-planned
// robots: Occupied[(cell,t)] blocks a vertex at time t, edge entries prevent
// two robots swapping cells head-on between t and t+1.
type Reservation struct {
	Occupied map[timedCell]bool
	edges    map[edgeKey]bool
	// parked robots block their final cell for all t >= that entry
	parkedFrom map[Cell]int
}

type timedCell struct {
	C Cell
	T int
}

// TimedCell builds an occupancy key for direct Reservation.Occupied writes
// (used by callers that model non-path waits, e.g. dwell at a pickup).
func TimedCell(c Cell, t int) timedCell { return timedCell{c, t} }

type edgeKey struct {
	A, B Cell // ordered: A < B lexicographically
	T    int  // transition happening between T and T+1
}

func NewReservation() *Reservation {
	return &Reservation{
		Occupied:   map[timedCell]bool{},
		edges:      map[edgeKey]bool{},
		parkedFrom: map[Cell]int{},
	}
}

func orderedEdge(a, b Cell, t int) edgeKey {
	if a.Y > b.Y || (a.Y == b.Y && a.X > b.X) {
		a, b = b, a
	}
	return edgeKey{a, b, t}
}

// ReservePath claims every (cell,t) along the path, its transitions, and
// parks the robot on the final cell from the path's end onward.
func (r *Reservation) ReservePath(path []Cell, startT int) {
	for i, c := range path {
		r.Occupied[timedCell{c, startT + i}] = true
		if i > 0 {
			r.edges[orderedEdge(path[i-1], c, startT+i-1)] = true
		}
	}
	if len(path) > 0 {
		last := path[len(path)-1]
		end := startT + len(path) - 1
		if cur, ok := r.parkedFrom[last]; !ok || end < cur {
			r.parkedFrom[last] = end
		}
	}
}

// Park marks a cell occupied for all times >= t (an idle robot standing there).
func (r *Reservation) Park(c Cell, t int) { r.parkedFrom[c] = t }

// Unpark removes a standing reservation (the robot is about to move).
func (r *Reservation) Unpark(c Cell) { delete(r.parkedFrom, c) }

func (r *Reservation) blockedAt(c Cell, t int) bool {
	if r.Occupied[timedCell{c, t}] {
		return true
	}
	if from, ok := r.parkedFrom[c]; ok && t >= from {
		return true
	}
	return false
}

// --- space-time A* ---

type stNode struct {
	c   Cell
	t   int
	f   int
	idx int
}

type stHeap []*stNode

func (h stHeap) Len() int           { return len(h) }
func (h stHeap) Less(i, j int) bool { return h[i].f < h[j].f }
func (h stHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].idx, h[j].idx = i, j }
func (h *stHeap) Push(x any)        { n := x.(*stNode); n.idx = len(*h); *h = append(*h, n) }
func (h *stHeap) Pop() any          { old := *h; n := old[len(old)-1]; *h = old[:len(old)-1]; return n }

func manhattan(a, b Cell) int { return abs(a.X-b.X) + abs(a.Y-b.Y) }

// FindPath runs space-time A* from start (at time startT) to goal, avoiding
// static obstacles and the reservation table. Moves: 4-neighbours + wait,
// 1 tick each. Returns the path INCLUDING the start cell, or nil if none is
// found within horizon ticks.
func FindPath(g *Grid, res *Reservation, start, goal Cell, startT, horizon int) []Cell {
	if g.Blocked(goal) {
		return nil
	}
	type key = timedCell
	came := map[key]key{}
	gScore := map[key]int{start2key(start, startT): 0}

	open := &stHeap{}
	heap.Init(open)
	heap.Push(open, &stNode{c: start, t: startT, f: manhattan(start, goal)})

	moves := []Cell{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for open.Len() > 0 {
		cur := heap.Pop(open).(*stNode)
		if cur.c == goal {
			// require the goal to stay free briefly so we don't stop inside
			// someone's imminent path; then reconstruct
			return reconstruct(came, key{cur.c, cur.t}, start, startT)
		}
		if cur.t-startT > horizon {
			continue
		}
		for _, m := range moves {
			next := Cell{cur.c.X + m.X, cur.c.Y + m.Y}
			nt := cur.t + 1
			if g.Blocked(next) || res.blockedAt(next, nt) {
				continue
			}
			if m != (Cell{0, 0}) && res.edges[orderedEdge(cur.c, next, cur.t)] {
				continue
			}
			nk := key{next, nt}
			tentative := gScore[key{cur.c, cur.t}] + 1
			if old, ok := gScore[nk]; ok && tentative >= old {
				continue
			}
			gScore[nk] = tentative
			came[nk] = key{cur.c, cur.t}
			heap.Push(open, &stNode{c: next, t: nt, f: tentative + manhattan(next, goal)})
		}
	}
	return nil
}

func start2key(c Cell, t int) timedCell { return timedCell{c, t} }

func reconstruct(came map[timedCell]timedCell, end timedCell, start Cell, startT int) []Cell {
	var rev []Cell
	cur := end
	for {
		rev = append(rev, cur.C)
		if cur.C == start && cur.T == startT {
			break
		}
		prev, ok := came[cur]
		if !ok {
			break
		}
		cur = prev
	}
	path := make([]Cell, len(rev))
	for i, c := range rev {
		path[len(rev)-1-i] = c
	}
	return path
}
