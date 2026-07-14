// Package allocation assigns pending tasks to available robots.
//
// The Strategy interface makes the allocation method pluggable so the benchmark
// can swap greedy / hungarian / auction on identical inputs. This is one axis of
// the study (allocation x coordination); see docs/research-proposal.md.
package allocation

import "github.com/yourname/warefleet/fleet_manager/internal/model"

// Strategy assigns a set of pending tasks to a set of available robots.
// It returns the assignments it could make (a task may be left unassigned if no
// suitable robot is free).
type Strategy interface {
	Name() string
	Assign(tasks []model.Task, robots []model.Robot) []model.Assignment
}

// New returns the strategy named by `name` ("greedy" | "hungarian" | "auction").
func New(name string) (Strategy, error) {
	switch name {
	case "greedy":
		return &Greedy{}, nil
	case "hungarian":
		return &Hungarian{}, nil
	case "auction":
		return &Auction{}, nil
	default:
		return nil, &UnknownStrategyError{Name: name}
	}
}

// UnknownStrategyError is returned by New for an unrecognised strategy name.
type UnknownStrategyError struct{ Name string }

func (e *UnknownStrategyError) Error() string {
	return "unknown allocation strategy: " + e.Name
}

// dist2 is the squared Euclidean distance — enough for comparisons, avoids sqrt.
func dist2(a, b model.Point) float64 {
	dx, dy := a.X-b.X, a.Y-b.Y
	return dx*dx + dy*dy
}

// availableRobots filters the fleet down to robots that can take new work.
func availableRobots(robots []model.Robot) []model.Robot {
	out := make([]model.Robot, 0, len(robots))
	for _, r := range robots {
		if r.Status == model.StatusIdle {
			out = append(out, r)
		}
	}
	return out
}
