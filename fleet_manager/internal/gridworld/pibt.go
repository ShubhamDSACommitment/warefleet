package gridworld

import "sort"

// PIBT — Priority Inheritance with Backtracking.
//
// Faithful implementation of Algorithm 1 in:
//
//	Okumura, Machida, Défago, Tamura: "Priority inheritance with backtracking
//	for iterative multi-agent path finding", Artificial Intelligence 310 (2022).
//	(conference version: IJCAI 2019)
//
// One PIBT step decides every agent's next cell for a single timestep:
// agents are processed in dynamic-priority order (longest time since last
// goal arrival first); an agent occupying a higher-priority agent's desired
// cell temporarily INHERITS that priority and must move away; if it cannot,
// the decision BACKTRACKS and the pusher tries its next-best cell.
//
// PIBT plans one step at a time — no long-horizon search — which is why it
// scales to very large fleets and fits lifelong (never-ending) operation.
// Deviation from the paper: candidate tie-breaking is deterministic (cell
// order) instead of random, so benchmark runs are reproducible.
type PIBT struct {
	g     *Grid
	dists map[Cell][]int // BFS distance field per goal, lazily cached
}

// PIBTAgent is one robot in the PIBT stepper. Pos/Goal are grid cells; an
// idle robot simply sets Goal = Pos (it will be pushed aside when in the way
// and drift back afterwards — no special "parked" handling needed).
type PIBTAgent struct {
	ID      string
	Pos     Cell
	Goal    Cell
	elapsed int   // steps since last goal arrival — the dynamic priority
	next    *Cell // this step's decision (nil = undecided)
}

func NewPIBT(g *Grid) *PIBT {
	return &PIBT{g: g, dists: map[Cell][]int{}}
}

// dist is the true (BFS) grid distance from c to goal; -1 if unreachable.
func (p *PIBT) dist(goal, c Cell) int {
	field, ok := p.dists[goal]
	if !ok {
		field = p.bfs(goal)
		p.dists[goal] = field
	}
	if !p.g.InBounds(c) {
		return -1
	}
	return field[p.g.idx(c)]
}

func (p *PIBT) bfs(goal Cell) []int {
	field := make([]int, p.g.W*p.g.H)
	for i := range field {
		field[i] = -1
	}
	if p.g.Blocked(goal) {
		return field
	}
	queue := []Cell{goal}
	field[p.g.idx(goal)] = 0
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, m := range [4]Cell{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			n := Cell{c.X + m.X, c.Y + m.Y}
			if p.g.InBounds(n) && !p.g.Blocked(n) && field[p.g.idx(n)] == -1 {
				field[p.g.idx(n)] = field[p.g.idx(c)] + 1
				queue = append(queue, n)
			}
		}
	}
	return field
}

// Step decides and applies one timestep for all agents. Guarantees per the
// paper: no two agents end on the same cell (vertex conflict) and no two
// agents swap cells (edge conflict). Every agent with the highest priority
// is guaranteed to move closer to its goal.
func (p *PIBT) Step(agents []*PIBTAgent) {
	// dynamic priority: longest-waiting first, ID as the stable tie-break
	order := make([]*PIBTAgent, len(agents))
	copy(order, agents)
	sort.SliceStable(order, func(i, j int) bool {
		if order[i].elapsed != order[j].elapsed {
			return order[i].elapsed > order[j].elapsed
		}
		return order[i].ID < order[j].ID
	})

	claimed := map[Cell]*PIBTAgent{} // next-position reservations this step
	byPos := map[Cell]*PIBTAgent{}   // current positions
	for _, a := range order {
		a.next = nil
		byPos[a.Pos] = a
	}

	for _, a := range order {
		if a.next == nil {
			p.plan(a, nil, claimed, byPos)
		}
	}

	for _, a := range agents {
		if a.next != nil {
			a.Pos = *a.next
			a.next = nil
		}
		if a.Pos == a.Goal {
			a.elapsed = 0
		} else {
			a.elapsed++
		}
	}
}

// plan is funcPIBT(a_i, a_j) from the paper: decide agent a's next cell,
// recursively pushing undecided lower-priority agents out of the way.
// Returns false (invalid) if a cannot move anywhere and stays put.
func (p *PIBT) plan(a, pusher *PIBTAgent, claimed map[Cell]*PIBTAgent, byPos map[Cell]*PIBTAgent) bool {
	// candidates: 4 neighbours + stay, closest-to-goal first, deterministic
	cand := make([]Cell, 0, 5)
	for _, m := range [5]Cell{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		c := Cell{a.Pos.X + m.X, a.Pos.Y + m.Y}
		if !p.g.Blocked(c) && p.dist(a.Goal, c) >= 0 {
			cand = append(cand, c)
		}
	}
	sort.SliceStable(cand, func(i, j int) bool {
		di, dj := p.dist(a.Goal, cand[i]), p.dist(a.Goal, cand[j])
		if di != dj {
			return di < dj
		}
		if cand[i].Y != cand[j].Y {
			return cand[i].Y < cand[j].Y
		}
		return cand[i].X < cand[j].X
	})

	for _, u := range cand {
		if claimed[u] != nil {
			continue // vertex conflict with an already-decided agent
		}
		if pusher != nil && u == pusher.Pos {
			continue // would swap positions with the agent pushing us
		}
		claimed[u] = a
		uCopy := u
		a.next = &uCopy
		if occupant := byPos[u]; occupant != nil && occupant != a && occupant.next == nil {
			// priority inheritance: the occupant must vacate u before we
			// can commit; if it can't (backtracking), it stays on u and we
			// must pick a different cell.
			if !p.plan(occupant, a, claimed, byPos) {
				claimed[u] = occupant
				a.next = nil
				continue
			}
		}
		return true
	}

	// nowhere to go: stay put, tell the pusher to backtrack
	pos := a.Pos
	a.next = &pos
	return false
}
