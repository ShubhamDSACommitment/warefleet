package allocation

import (
	"math"
	"testing"
	"time"

	"github.com/yourname/warefleet/fleet_manager/internal/model"
)

func robot(id string, x, y float64) model.Robot {
	return model.Robot{ID: id, Pose: model.Point{X: x, Y: y}, Status: model.StatusIdle}
}

func task(id string, x, y float64) model.Task {
	return model.Task{ID: id, Kind: "transport", Pickup: model.Point{X: x, Y: y},
		Priority: 1, CreatedAt: time.Unix(0, 0)}
}

// totalCost sums euclidean robot->pickup distance over the assignments.
func totalCost(t *testing.T, as []model.Assignment, robots []model.Robot, tasks []model.Task) float64 {
	t.Helper()
	rByID := map[string]model.Robot{}
	for _, r := range robots {
		rByID[r.ID] = r
	}
	tByID := map[string]model.Task{}
	for _, tk := range tasks {
		tByID[tk.ID] = tk
	}
	sum := 0.0
	for _, a := range as {
		sum += math.Sqrt(dist2(rByID[a.RobotID].Pose, tByID[a.TaskID].Pickup))
	}
	return sum
}

// assertValid checks structural invariants: no robot or task used twice,
// only idle robots, only known tasks.
func assertValid(t *testing.T, as []model.Assignment, robots []model.Robot, tasks []model.Task) {
	t.Helper()
	seenR, seenT := map[string]bool{}, map[string]bool{}
	idle := map[string]bool{}
	for _, r := range robots {
		if r.Status == model.StatusIdle {
			idle[r.ID] = true
		}
	}
	known := map[string]bool{}
	for _, tk := range tasks {
		known[tk.ID] = true
	}
	for _, a := range as {
		if seenR[a.RobotID] {
			t.Fatalf("robot %s assigned twice", a.RobotID)
		}
		if seenT[a.TaskID] {
			t.Fatalf("task %s assigned twice", a.TaskID)
		}
		if !idle[a.RobotID] {
			t.Fatalf("non-idle robot %s got an assignment", a.RobotID)
		}
		if !known[a.TaskID] {
			t.Fatalf("unknown task %s assigned", a.TaskID)
		}
		seenR[a.RobotID], seenT[a.TaskID] = true, true
	}
}

// The classic case where greedy is suboptimal: greedy grabs the nearest robot
// for the first task and forces a long trip for the second. Hungarian and the
// auction must both find the cheaper global matching.
func TestHungarianAndAuctionBeatGreedyOnCrossing(t *testing.T) {
	robots := []model.Robot{robot("r1", 0, 0), robot("r2", 2, 0)}
	tasks := []model.Task{task("t1", 1.9, 0), task("t2", 2.1, 0)}

	greedyCost := totalCost(t, (&Greedy{}).Assign(tasks, robots), robots, tasks)
	hungCost := totalCost(t, (&Hungarian{}).Assign(tasks, robots), robots, tasks)
	auctCost := totalCost(t, (&Auction{}).Assign(tasks, robots), robots, tasks)

	// greedy: t1->r2 (0.1) then t2->r1 (2.1) = 2.2; optimal: 1.9 + 0.1 = 2.0
	if greedyCost < 2.19 || greedyCost > 2.21 {
		t.Fatalf("greedy cost = %.3f, expected ~2.2 (test premise broken)", greedyCost)
	}
	if hungCost > 2.01 {
		t.Fatalf("hungarian cost = %.3f, expected optimal ~2.0", hungCost)
	}
	if auctCost > hungCost+4*epsilon {
		t.Fatalf("auction cost = %.3f, expected within n*eps of optimal %.3f", auctCost, hungCost)
	}
}

func TestAllStrategiesStructuralInvariants(t *testing.T) {
	robots := []model.Robot{
		robot("r1", 0, 0), robot("r2", 5, 0), robot("r3", 0, 5),
		{ID: "r4", Pose: model.Point{X: 1, Y: 1}, Status: model.StatusBusy},
		{ID: "r5", Pose: model.Point{X: 2, Y: 2}, Status: model.StatusOffline},
	}
	tasks := []model.Task{task("t1", 1, 0), task("t2", 4, 1)}

	for _, name := range []string{"greedy", "hungarian", "auction"} {
		s, err := New(name)
		if err != nil {
			t.Fatal(err)
		}
		as := s.Assign(tasks, robots)
		assertValid(t, as, robots, tasks)
		if len(as) != 2 { // 3 idle robots, 2 tasks -> both tasks assigned
			t.Fatalf("%s: expected 2 assignments, got %d", name, len(as))
		}
	}
}

func TestMoreTasksThanRobots(t *testing.T) {
	robots := []model.Robot{robot("r1", 0, 0)}
	tasks := []model.Task{task("t1", 1, 0), task("t2", 10, 0), task("t3", 5, 5)}

	for _, name := range []string{"greedy", "hungarian", "auction"} {
		s, _ := New(name)
		as := s.Assign(tasks, robots)
		assertValid(t, as, robots, tasks)
		if len(as) != 1 {
			t.Fatalf("%s: 1 robot, 3 tasks -> want 1 assignment, got %d", name, len(as))
		}
	}
}

func TestNoIdleRobotsNoAssignments(t *testing.T) {
	robots := []model.Robot{{ID: "r1", Status: model.StatusBusy}}
	tasks := []model.Task{task("t1", 1, 0)}
	for _, name := range []string{"greedy", "hungarian", "auction"} {
		s, _ := New(name)
		if as := s.Assign(tasks, robots); len(as) != 0 {
			t.Fatalf("%s: expected no assignments, got %d", name, len(as))
		}
	}
}

// Hungarian must match the brute-force optimum on random-ish fixed instances.
func TestHungarianOptimalVsBruteForce(t *testing.T) {
	robots := []model.Robot{
		robot("r1", 0, 0), robot("r2", 3, 1), robot("r3", -2, 4), robot("r4", 6, -3),
	}
	tasks := []model.Task{
		task("t1", 1, 1), task("t2", -1, 3), task("t3", 5, -2), task("t4", 2, -2),
	}

	hungCost := totalCost(t, (&Hungarian{}).Assign(tasks, robots), robots, tasks)

	// brute force all 4! permutations
	best := math.Inf(1)
	perm := []int{0, 1, 2, 3}
	var rec func(k int)
	rec = func(k int) {
		if k == len(perm) {
			sum := 0.0
			for i, j := range perm {
				sum += math.Sqrt(dist2(robots[i].Pose, tasks[j].Pickup))
			}
			if sum < best {
				best = sum
			}
			return
		}
		for i := k; i < len(perm); i++ {
			perm[k], perm[i] = perm[i], perm[k]
			rec(k + 1)
			perm[k], perm[i] = perm[i], perm[k]
		}
	}
	rec(0)

	if math.Abs(hungCost-best) > 1e-9 {
		t.Fatalf("hungarian cost %.6f != brute-force optimum %.6f", hungCost, best)
	}
}

func TestDeterministicAcrossShuffledInput(t *testing.T) {
	robots := []model.Robot{robot("r1", 0, 0), robot("r2", 3, 1), robot("r3", -2, 4)}
	tasks := []model.Task{task("t1", 1, 1), task("t2", -1, 3)}
	shuffledR := []model.Robot{robots[2], robots[0], robots[1]}
	shuffledT := []model.Task{tasks[1], tasks[0]}

	for _, name := range []string{"greedy", "hungarian", "auction"} {
		s, _ := New(name)
		a := s.Assign(tasks, robots)
		b := s.Assign(shuffledT, shuffledR)
		if len(a) != len(b) {
			t.Fatalf("%s: nondeterministic count", name)
		}
		pairs := map[string]string{}
		for _, x := range a {
			pairs[x.TaskID] = x.RobotID
		}
		for _, x := range b {
			if pairs[x.TaskID] != x.RobotID {
				t.Fatalf("%s: nondeterministic: %s -> %s vs %s",
					name, x.TaskID, pairs[x.TaskID], x.RobotID)
			}
		}
	}
}
