package coordination

import (
	"fmt"
	"sort"

	"github.com/yourname/warefleet/fleet_manager/internal/gridworld"
	"github.com/yourname/warefleet/fleet_manager/internal/model"
)

// Planner is the pluggable MAPF layer: given each robot's start and its assigned
// goal, produce conflict-free paths for all of them. Second axis of the study.
type Planner interface {
	Name() string
	// Plan returns one PathPlan per request; err if no conflict-free solution.
	Plan(reqs []PlanRequest) ([]model.PathPlan, error)
}

// PlanRequest pairs a robot's current position with the task it must serve.
type PlanRequest struct {
	RobotID string
	TaskID  string
	Start   model.Point
	Goal    model.Point
}

// NewPlanner returns the planner named by `name`
// ("prioritized" | "cbs" | "pibt" | "lacam").
func NewPlanner(name string) (Planner, error) {
	switch name {
	case "prioritized":
		return &Prioritized{}, nil
	case "cbs":
		return &CBS{}, nil
	case "pibt":
		return &PIBT{}, nil
	case "lacam":
		return &LaCAM{}, nil
	default:
		return nil, &UnknownPlannerError{Name: name}
	}
}

// UnknownPlannerError is returned by NewPlanner for an unrecognised name.
type UnknownPlannerError struct{ Name string }

func (e *UnknownPlannerError) Error() string {
	return "unknown coordination planner: " + e.Name
}

// Prioritized is the baseline MAPF planner (Silver 2005): plan robots one at
// a time with space-time A*, treating already-planned robots' trajectories as
// moving obstacles via a reservation table. Fast and incomplete — the standard
// baseline the CBS/PIBT stretch goals would be compared against.
type Prioritized struct {
	Grid *gridworld.Grid // defaults to the warehouse grid
}

func (p *Prioritized) Name() string { return "prioritized" }

const planHorizon = 400 // ticks; 24x16 grid paths are <100 even with waits

// Plan plans reqs in order (callers sort by priority; we keep input order but
// make it deterministic by RobotID for equal inputs). Waypoints are map-frame
// cell centres, one per tick — arrival_times are implicit (index = tick).
func (p *Prioritized) Plan(reqs []PlanRequest) ([]model.PathPlan, error) {
	g := p.Grid
	if g == nil {
		g = gridworld.NewWarehouse()
	}
	ordered := make([]PlanRequest, len(reqs))
	copy(ordered, reqs)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].RobotID < ordered[j].RobotID })

	res := gridworld.NewReservation()
	// every robot's start is occupied until its own plan moves it
	starts := map[string]gridworld.Cell{}
	for _, r := range ordered {
		c := g.NearestFree(g.ToCell(r.Start.X, r.Start.Y))
		starts[r.RobotID] = c
		res.Park(c, 0)
	}

	plans := make([]model.PathPlan, 0, len(ordered))
	for _, r := range ordered {
		start := starts[r.RobotID]
		goal := g.NearestFree(g.ToCell(r.Goal.X, r.Goal.Y))
		res.Unpark(start)
		path := gridworld.FindPath(g, res, start, goal, 0, planHorizon)
		if path == nil {
			return nil, fmt.Errorf("prioritized: no conflict-free path for %s (%v -> %v)",
				r.RobotID, start, goal)
		}
		res.ReservePath(path, 0)

		wp := make([]model.Point, len(path))
		for i, c := range path {
			x, y := g.ToMap(c)
			wp[i] = model.Point{X: x, Y: y}
		}
		plans = append(plans, model.PathPlan{RobotID: r.RobotID, TaskID: r.TaskID, Waypoints: wp})
	}
	return plans, nil
}

// CBS — Conflict-Based Search (optimal, complete). Stretch goal.
type CBS struct{}

func (c *CBS) Name() string { return "cbs" }
func (c *CBS) Plan(reqs []PlanRequest) ([]model.PathPlan, error) {
	panic("CBS.Plan: stretch goal")
}

// PIBT — Priority Inheritance with Backtracking (Okumura, Machida, Défago,
// Tamura; AIJ 2022). Core rule implemented in gridworld/pibt.go. Here it is
// exposed through the one-shot Planner interface by stepping all agents until
// everyone reaches its goal (the paper's "iterative use for one-shot MAPF").
// PIBT is complete on biconnected graphs; the step cap catches the rest.
type PIBT struct {
	Grid *gridworld.Grid // defaults to the warehouse grid
}

func (p *PIBT) Name() string { return "pibt" }

func (p *PIBT) Plan(reqs []PlanRequest) ([]model.PathPlan, error) {
	g := p.Grid
	if g == nil {
		g = gridworld.NewWarehouse()
	}
	stepper := gridworld.NewPIBT(g)

	agents := make([]*gridworld.PIBTAgent, len(reqs))
	paths := make(map[string][]gridworld.Cell, len(reqs))
	for i, r := range reqs {
		start := g.NearestFree(g.ToCell(r.Start.X, r.Start.Y))
		goal := g.NearestFree(g.ToCell(r.Goal.X, r.Goal.Y))
		agents[i] = &gridworld.PIBTAgent{ID: r.RobotID, Pos: start, Goal: goal}
		paths[r.RobotID] = []gridworld.Cell{start}
	}

	for step := 0; step < planHorizon; step++ {
		stepper.Step(agents)
		done := true
		for _, a := range agents {
			paths[a.ID] = append(paths[a.ID], a.Pos)
			if a.Pos != a.Goal {
				done = false
			}
		}
		if done {
			break
		}
	}

	plans := make([]model.PathPlan, 0, len(reqs))
	for _, r := range reqs {
		cells := paths[r.RobotID]
		if cells[len(cells)-1] != agents[indexOf(reqs, r.RobotID)].Goal {
			return nil, fmt.Errorf("pibt: %s did not reach its goal within %d steps",
				r.RobotID, planHorizon)
		}
		wp := make([]model.Point, len(cells))
		for i, c := range cells {
			x, y := g.ToMap(c)
			wp[i] = model.Point{X: x, Y: y}
		}
		plans = append(plans, model.PathPlan{RobotID: r.RobotID, TaskID: r.TaskID, Waypoints: wp})
	}
	return plans, nil
}

func indexOf(reqs []PlanRequest, robotID string) int {
	for i, r := range reqs {
		if r.RobotID == robotID {
			return i
		}
	}
	return -1
}

// LaCAM — fast, near-optimal, lifelong-capable MAPF (the Coord Lab reference
// line). Stretch goal; strongest signal for MAPF-focused labs.
type LaCAM struct{}

func (l *LaCAM) Name() string { return "lacam" }
func (l *LaCAM) Plan(reqs []PlanRequest) ([]model.PathPlan, error) {
	panic("LaCAM.Plan: stretch goal")
}
