package coordination

import "github.com/yourname/warefleet/fleet_manager/internal/model"

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

// Prioritized is the baseline MAPF planner (space-time A* + reservation table).
// Port the validated logic from ros2_ws/.../warefleet_mapf/prioritized.py.
type Prioritized struct{}

func (p *Prioritized) Name() string { return "prioritized" }
func (p *Prioritized) Plan(reqs []PlanRequest) ([]model.PathPlan, error) {
	panic("Prioritized.Plan: implement in Week 5 (port from Python reference)")
}

// CBS — Conflict-Based Search (optimal, complete). Stretch goal.
type CBS struct{}

func (c *CBS) Name() string { return "cbs" }
func (c *CBS) Plan(reqs []PlanRequest) ([]model.PathPlan, error) {
	panic("CBS.Plan: stretch goal")
}

// PIBT — scalable near-real-time MAPF. Stretch goal; direct match to MAPF labs.
type PIBT struct{}

func (p *PIBT) Name() string { return "pibt" }
func (p *PIBT) Plan(reqs []PlanRequest) ([]model.PathPlan, error) {
	panic("PIBT.Plan: stretch goal")
}

// LaCAM — fast, near-optimal, lifelong-capable MAPF (the Coord Lab reference
// line). Stretch goal; strongest signal for MAPF-focused labs.
type LaCAM struct{}

func (l *LaCAM) Name() string { return "lacam" }
func (l *LaCAM) Plan(reqs []PlanRequest) ([]model.PathPlan, error) {
	panic("LaCAM.Plan: stretch goal")
}
