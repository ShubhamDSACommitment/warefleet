// Command gridsim is the IDEALIZED simulator tier: the warehouse as a grid,
// robots moving one cell per tick, no physics — the fast half of the
// two-simulator design (the realistic half is ROS 2 + Gazebo).
//
// It reuses the exact same allocation.Strategy implementations and the
// prioritized MAPF planner as the real fleet manager: the strategies cannot
// tell which world they are in. That symmetry is what makes the H1 comparison
// honest — one variable (execution realism), everything else identical.
//
//	gridsim -schedule sched.json -strategy hungarian -robots 8
//
// Reads the order schedule dumped by the order feeder (--dump), runs the
// lifelong loop (orders arrive by timestamp; idle robots get assigned; paths
// are planned conflict-free; robots advance), prints one JSON result line.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/yourname/warefleet/fleet_manager/internal/allocation"
	"github.com/yourname/warefleet/fleet_manager/internal/coordination"
	"github.com/yourname/warefleet/fleet_manager/internal/gridworld"
	"github.com/yourname/warefleet/fleet_manager/internal/model"
)

const (
	tickSeconds = 0.5 // one cell (0.5 m) per tick ~= 1 m/s idealized robot
	dwellTicks  = 4   // 2 s handling at pickup, mirrors the agent's dwell_sec
)

type schedOrder struct {
	T        float64                `json:"t"`
	TaskID   string                 `json:"task_id"`
	Kind     string                 `json:"kind"`
	Pickup   struct{ X, Y float64 } `json:"pickup"`
	Dropoff  struct{ X, Y float64 } `json:"dropoff"`
	Priority int                    `json:"priority"`
}

type simRobot struct {
	id      string
	cell    gridworld.Cell
	path    []gridworld.Cell // remaining cells (leg in progress)
	dwell   int              // ticks left standing at pickup
	leg2    []gridworld.Cell // dropoff leg, starts after dwell
	task    *schedOrder
	pickupT int // tick the task was picked up for makespan bookkeeping
}

func main() {
	schedPath := flag.String("schedule", "", "order schedule JSON (from order_feeder --dump)")
	strategy := flag.String("strategy", "greedy", "greedy | hungarian | auction")
	nRobots := flag.Int("robots", 4, "fleet size")
	maxOrders := flag.Int("orders", 0, "cap orders (0 = whole schedule)")
	rate := flag.Float64("rate", 1.0, "time compression: order timestamps are divided by this "+
		"(must match the realistic tier's playback --rate for a fair comparison)")
	plannerName := flag.String("planner", "prioritized",
		"prioritized (full-path space-time A* at assignment) | pibt (per-tick stepping, AIJ 2022)")
	flag.Parse()

	raw, err := os.ReadFile(*schedPath)
	if err != nil {
		log.Fatalf("schedule: %v", err)
	}
	var orders []schedOrder
	if err := json.Unmarshal(raw, &orders); err != nil {
		log.Fatalf("schedule: %v", err)
	}
	sort.SliceStable(orders, func(i, j int) bool { return orders[i].T < orders[j].T })
	if *maxOrders > 0 && len(orders) > *maxOrders {
		orders = orders[:*maxOrders]
	}
	for i := range orders {
		orders[i].T /= *rate
	}

	alloc, err := allocation.New(*strategy)
	if err != nil {
		log.Fatal(err)
	}
	grid := gridworld.NewWarehouse()
	planner := &coordination.Prioritized{Grid: grid}

	// spawn robots spread along the west wall, mirroring the Gazebo poses
	robots := make([]*simRobot, *nRobots)
	for i := range robots {
		c := grid.NearestFree(gridworld.Cell{X: 2 + (i%2)*2, Y: 3 + (i/2)*3})
		robots[i] = &simRobot{id: fmt.Sprintf("robot%d", i+1), cell: c}
	}

	var completed int
	var makespanSum, simSeconds float64
	switch *plannerName {
	case "pibt":
		completed, makespanSum, simSeconds = runPIBT(grid, alloc, robots, orders)
	case "prioritized":
		completed, makespanSum, simSeconds = runPrioritized(grid, planner, alloc, robots, orders)
	default:
		log.Fatalf("unknown planner %q", *plannerName)
	}

	out := map[string]any{
		"tier":               "idealized",
		"strategy":           *strategy,
		"planner":            *plannerName,
		"robots":             *nRobots,
		"orders":             len(orders),
		"completed":          completed,
		"sim_s":              simSeconds,
		"throughput_per_min": round2(float64(completed) / (simSeconds / 60.0)),
		"makespan_mean_s":    round2(makespanSum / float64(max(completed, 1))),
	}
	json.NewEncoder(os.Stdout).Encode(out)
}

func runPrioritized(grid *gridworld.Grid, planner *coordination.Prioritized,
	alloc allocation.Strategy, robots []*simRobot, orders []schedOrder) (int, float64, float64) {

	pending := []schedOrder{}
	completed := 0
	makespanSum := 0.0
	next := 0
	tick := 0

	for completed < len(orders) {
		tick++
		now := float64(tick) * tickSeconds

		for next < len(orders) && orders[next].T <= now {
			pending = append(pending, orders[next])
			next++
		}

		// allocation over idle robots, exactly like the coordinator's step()
		if len(pending) > 0 {
			var idle []model.Robot
			for _, r := range robots {
				if r.task == nil {
					x, y := grid.ToMap(r.cell)
					idle = append(idle, model.Robot{
						ID: r.id, Pose: model.Point{X: x, Y: y}, Status: model.StatusIdle})
				}
			}
			var tasks []model.Task
			for _, o := range pending {
				tasks = append(tasks, model.Task{
					ID: o.TaskID, Kind: o.Kind, Priority: o.Priority,
					Pickup:  model.Point{X: o.Pickup.X, Y: o.Pickup.Y},
					Dropoff: model.Point{X: o.Dropoff.X, Y: o.Dropoff.Y},
				})
			}
			for _, a := range alloc.Assign(tasks, idle) {
				r := findRobot(robots, a.RobotID)
				o := takeOrder(&pending, a.TaskID)
				if r == nil || o == nil {
					continue
				}
				if !planTask(grid, planner, robots, r, o) {
					pending = append(pending, *o) // congested now; retry next tick
					continue
				}
			}
		}

		// advance the world one tick
		for _, r := range robots {
			switch {
			case len(r.path) > 0:
				r.cell, r.path = r.path[0], r.path[1:]
				if len(r.path) == 0 && r.leg2 != nil {
					r.dwell = dwellTicks
				}
			case r.dwell > 0:
				r.dwell--
				if r.dwell == 0 {
					r.path, r.leg2 = r.leg2, nil
				}
			case r.task != nil:
				makespanSum += now - r.task.T
				completed++
				r.task = nil
			}
		}

		if tick > 200000 {
			log.Fatalf("gridsim: no progress (completed %d/%d)", completed, len(orders))
		}
	}
	return completed, makespanSum, float64(tick) * tickSeconds
}

// runPIBT executes the lifelong loop with per-tick PIBT stepping (AIJ 2022):
// no precomputed paths — every tick, all robots take one conflict-free step
// toward their current target (idle robots target their own cell and get
// pushed aside as needed). This is PIBT's intended lifelong usage.
func runPIBT(grid *gridworld.Grid, alloc allocation.Strategy,
	robots []*simRobot, orders []schedOrder) (int, float64, float64) {

	stepper := gridworld.NewPIBT(grid)
	agents := make([]*gridworld.PIBTAgent, len(robots))
	targets := make([]struct {
		pickup, dropoff gridworld.Cell
		leg             string // "" | "pickup" | "dropoff"
	}, len(robots))
	for i, r := range robots {
		agents[i] = &gridworld.PIBTAgent{ID: r.id, Pos: r.cell, Goal: r.cell}
	}

	pending := []schedOrder{}
	completed := 0
	makespanSum := 0.0
	next := 0
	tick := 0

	for completed < len(orders) {
		tick++
		now := float64(tick) * tickSeconds

		for next < len(orders) && orders[next].T <= now {
			pending = append(pending, orders[next])
			next++
		}

		if len(pending) > 0 {
			var idle []model.Robot
			for i, r := range robots {
				if r.task == nil {
					x, y := grid.ToMap(agents[i].Pos)
					idle = append(idle, model.Robot{
						ID: r.id, Pose: model.Point{X: x, Y: y}, Status: model.StatusIdle})
				}
			}
			var tasks []model.Task
			for _, o := range pending {
				tasks = append(tasks, model.Task{
					ID: o.TaskID, Kind: o.Kind, Priority: o.Priority,
					Pickup:  model.Point{X: o.Pickup.X, Y: o.Pickup.Y},
					Dropoff: model.Point{X: o.Dropoff.X, Y: o.Dropoff.Y},
				})
			}
			for _, a := range alloc.Assign(tasks, idle) {
				o := takeOrder(&pending, a.TaskID)
				ri := robotIndex(robots, a.RobotID)
				if o == nil || ri < 0 {
					continue
				}
				robots[ri].task = o
				targets[ri].pickup = grid.NearestFree(grid.ToCell(o.Pickup.X, o.Pickup.Y))
				targets[ri].dropoff = grid.NearestFree(grid.ToCell(o.Dropoff.X, o.Dropoff.Y))
				targets[ri].leg = "pickup"
				agents[ri].Goal = targets[ri].pickup
			}
		}

		stepper.Step(agents)

		for i, r := range robots {
			if r.task == nil {
				agents[i].Goal = agents[i].Pos // idle: hold position, yield to movers
				continue
			}
			switch targets[i].leg {
			case "pickup":
				if agents[i].Pos == targets[i].pickup {
					if r.dwell == 0 {
						r.dwell = dwellTicks
					}
					r.dwell-- // dwell only counts while standing at the pickup
					if r.dwell == 0 {
						targets[i].leg = "dropoff"
						agents[i].Goal = targets[i].dropoff
					}
				}
			case "dropoff":
				if agents[i].Pos == targets[i].dropoff {
					makespanSum += now - r.task.T
					completed++
					r.task = nil
					targets[i].leg = ""
					agents[i].Goal = agents[i].Pos
				}
			}
		}

		if tick > 200000 {
			log.Fatalf("gridsim(pibt): no progress (completed %d/%d)", completed, len(orders))
		}
	}
	return completed, makespanSum, float64(tick) * tickSeconds
}

func robotIndex(rs []*simRobot, id string) int {
	for i, r := range rs {
		if r.id == id {
			return i
		}
	}
	return -1
}

// planTask plans start->pickup and pickup->dropoff with every other robot's
// position/trajectory reserved; returns false if no conflict-free path exists
// right now (caller re-queues).
func planTask(grid *gridworld.Grid, planner *coordination.Prioritized,
	robots []*simRobot, r *simRobot, o *schedOrder) bool {

	res := gridworld.NewReservation()
	parked := map[gridworld.Cell]bool{}
	for _, other := range robots {
		if other == r {
			continue
		}
		if len(other.path) > 0 {
			full := append([]gridworld.Cell{other.cell}, other.path...)
			res.ReservePath(full, 0)
			if other.leg2 != nil {
				// robot dwells at the pickup between legs — keep it reserved
				pickupCell := full[len(full)-1]
				for t := len(other.path); t <= len(other.path)+other.dwell; t++ {
					res.Occupied[gridworld.TimedCell(pickupCell, t)] = true
				}
				res.ReservePath(other.leg2, len(other.path)+other.dwell)
			}
		} else {
			res.Park(other.cell, 0)
			parked[other.cell] = true
		}
	}

	// a goal under a parked robot is unreachable forever — aim next to it
	// (real fleets nudge idle robots aside; adjacent-cell delivery is the
	// idealized equivalent and keeps the sim deadlock-free)
	resolve := func(c gridworld.Cell) gridworld.Cell {
		c = grid.NearestFree(c)
		for r := 0; parked[c] && r < 6; r++ {
			c = grid.NearestFreeExcluding(c, parked)
		}
		return c
	}
	pickup := resolve(grid.ToCell(o.Pickup.X, o.Pickup.Y))
	dropoff := resolve(grid.ToCell(o.Dropoff.X, o.Dropoff.Y))

	leg1 := gridworld.FindPath(grid, res, r.cell, pickup, 0, planHorizon)
	if leg1 == nil {
		return false
	}
	res.ReservePath(leg1, 0)
	leg2 := gridworld.FindPath(grid, res, pickup, dropoff, len(leg1)-1+dwellTicks, planHorizon)
	if leg2 == nil {
		return false
	}
	_ = planner       // reserved for future multi-request batch planning
	r.path = leg1[1:] // current cell excluded
	r.leg2 = leg2[1:]
	r.task = o
	if len(r.path) == 0 { // already standing at the pickup
		r.dwell = dwellTicks
	}
	return true
}

const planHorizon = 200 // ticks; grid diameter ~40, generous for waits

func findRobot(rs []*simRobot, id string) *simRobot {
	for _, r := range rs {
		if r.id == id {
			return r
		}
	}
	return nil
}

func takeOrder(pending *[]schedOrder, id string) *schedOrder {
	for i, o := range *pending {
		if o.TaskID == id {
			*pending = append((*pending)[:i], (*pending)[i+1:]...)
			return &o
		}
	}
	return nil
}

func round2(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
