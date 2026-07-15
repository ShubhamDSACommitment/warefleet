package gridworld

import (
	"fmt"
	"testing"
)

// assertStepInvariants checks PIBT's per-step guarantees: no vertex conflicts
// (two agents on one cell) and no edge conflicts (a swap).
func assertStepInvariants(t *testing.T, before map[string]Cell, agents []*PIBTAgent) {
	t.Helper()
	seen := map[Cell]string{}
	for _, a := range agents {
		if other, dup := seen[a.Pos]; dup {
			t.Fatalf("vertex conflict: %s and %s share %v", other, a.ID, a.Pos)
		}
		seen[a.Pos] = a.ID
	}
	for _, a := range agents {
		for _, b := range agents {
			if a.ID < b.ID && a.Pos == before[b.ID] && b.Pos == before[a.ID] &&
				a.Pos != b.Pos {
				t.Fatalf("edge conflict: %s and %s swapped %v<->%v",
					a.ID, b.ID, before[a.ID], before[b.ID])
			}
		}
	}
}

func run(t *testing.T, g *Grid, agents []*PIBTAgent, maxSteps int) int {
	t.Helper()
	p := NewPIBT(g)
	for step := 1; step <= maxSteps; step++ {
		before := map[string]Cell{}
		for _, a := range agents {
			before[a.ID] = a.Pos
		}
		p.Step(agents)
		assertStepInvariants(t, before, agents)
		done := true
		for _, a := range agents {
			if a.Pos != a.Goal {
				done = false
				break
			}
		}
		if done {
			return step
		}
	}
	return -1
}

func TestPIBTHeadOnInAisle(t *testing.T) {
	// two agents traversing the same aisle in opposite directions — the case
	// plain shortest-path following deadlocks on
	g := NewWarehouse()
	agents := []*PIBTAgent{
		{ID: "a", Pos: Cell{4, 8}, Goal: Cell{18, 8}},
		{ID: "b", Pos: Cell{18, 8}, Goal: Cell{4, 8}},
	}
	if steps := run(t, g, agents, 200); steps < 0 {
		t.Fatal("head-on agents never resolved")
	}
}

func TestPIBTPushesIdleAgentAside(t *testing.T) {
	// idle agent (goal = own cell) sits in the mover's path; priority
	// inheritance must push it out and it must return afterwards
	g := NewWarehouse()
	idle := &PIBTAgent{ID: "idle", Pos: Cell{10, 8}, Goal: Cell{10, 8}}
	mover := &PIBTAgent{ID: "mover", Pos: Cell{4, 8}, Goal: Cell{16, 8}}
	agents := []*PIBTAgent{idle, mover}

	if steps := run(t, g, agents, 200); steps < 0 {
		t.Fatal("mover never reached goal (idle agent not pushed aside)")
	}
	if idle.Pos != idle.Goal {
		t.Fatalf("idle agent did not return home: at %v", idle.Pos)
	}
}

func TestPIBTManyAgentsConvergeNoConflicts(t *testing.T) {
	// 12 agents, crossing assignments across the warehouse
	g := NewWarehouse()
	var agents []*PIBTAgent
	starts := []Cell{{2, 2}, {2, 5}, {2, 8}, {2, 11}, {21, 2}, {21, 5},
		{21, 8}, {21, 11}, {8, 2}, {8, 13}, {15, 2}, {15, 13}}
	for i, s := range starts {
		goal := starts[(i+6)%len(starts)] // swap sides
		agents = append(agents, &PIBTAgent{
			ID: fmt.Sprintf("a%02d", i), Pos: g.NearestFree(s), Goal: g.NearestFree(goal)})
	}
	if steps := run(t, g, agents, 500); steps < 0 {
		t.Fatal("12-agent instance did not converge")
	}
}

func TestPIBTDeterministic(t *testing.T) {
	g := NewWarehouse()
	mk := func() []*PIBTAgent {
		return []*PIBTAgent{
			{ID: "a", Pos: Cell{4, 8}, Goal: Cell{18, 8}},
			{ID: "b", Pos: Cell{18, 8}, Goal: Cell{4, 8}},
			{ID: "c", Pos: Cell{10, 3}, Goal: Cell{10, 12}},
		}
	}
	run1, run2 := mk(), mk()
	p1, p2 := NewPIBT(g), NewPIBT(g)
	for i := 0; i < 60; i++ {
		p1.Step(run1)
		p2.Step(run2)
		for k := range run1 {
			if run1[k].Pos != run2[k].Pos {
				t.Fatalf("nondeterministic at step %d agent %s: %v vs %v",
					i, run1[k].ID, run1[k].Pos, run2[k].Pos)
			}
		}
	}
}
