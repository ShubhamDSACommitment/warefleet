package gridworld

import "testing"

func TestPathAvoidsRacks(t *testing.T) {
	g := NewWarehouse()
	// across the warehouse, past both racks (map frame: spawn-ish -> dock)
	start := g.NearestFree(g.ToCell(0, 0))
	goal := g.NearestFree(g.ToCell(4.0, 0.5))
	path := FindPath(g, NewReservation(), start, goal, 0, 400)
	if path == nil {
		t.Fatal("no path across the warehouse")
	}
	if path[0] != start || path[len(path)-1] != goal {
		t.Fatalf("path endpoints wrong: %v ... %v", path[0], path[len(path)-1])
	}
	for _, c := range path {
		if g.Blocked(c) {
			t.Fatalf("path passes through blocked cell %v", c)
		}
	}
}

func TestReservationForcesWaitOrDetour(t *testing.T) {
	g := NewWarehouse()
	res := NewReservation()

	// robot A goes right along a free row; reserve its path
	a := FindPath(g, res, Cell{2, 8}, Cell{10, 8}, 0, 400)
	if a == nil {
		t.Fatal("A found no path")
	}
	res.ReservePath(a, 0)

	// robot B crosses A head-on along the same row
	b := FindPath(g, res, Cell{10, 8}, Cell{2, 8}, 0, 400)
	if b == nil {
		t.Fatal("B found no path (should wait or detour)")
	}

	// no vertex conflicts: same cell at same tick
	occA := map[timedCell]bool{}
	for i, c := range a {
		occA[timedCell{c, i}] = true
	}
	for i, c := range b {
		if occA[timedCell{c, i}] {
			t.Fatalf("vertex conflict at %v t=%d", c, i)
		}
	}
	// no edge swaps
	for i := 1; i < len(b) && i < len(a); i++ {
		if b[i-1] == a[i] && b[i] == a[i-1] {
			t.Fatalf("head-on swap between t=%d and t=%d", i-1, i)
		}
	}
}

func TestParkedRobotIsObstacle(t *testing.T) {
	g := NewWarehouse()
	res := NewReservation()
	res.Park(Cell{6, 8}, 0) // robot standing mid-row

	path := FindPath(g, res, Cell{2, 8}, Cell{10, 8}, 0, 400)
	if path == nil {
		t.Fatal("no path around parked robot")
	}
	for _, c := range path {
		if c == (Cell{6, 8}) {
			t.Fatal("path drives through the parked robot")
		}
	}
}
