package allocation

import (
	"time"

	"github.com/yourname/warefleet/fleet_manager/internal/model"
)

// Greedy assigns each task (highest priority first) to the nearest idle robot.
// Cheapest baseline; establishes the performance floor for the study.
type Greedy struct{}

func (g *Greedy) Name() string { return "greedy" }

func (g *Greedy) Assign(tasks []model.Task, robots []model.Robot) []model.Assignment {
	free := availableRobots(robots)
	tasks = sortedTasks(tasks) // highest priority, then oldest, first

	used := make(map[string]bool)
	assignments := make([]model.Assignment, 0, len(tasks))
	for _, t := range tasks {
		best := -1
		var bestD float64
		for i, r := range free {
			if used[r.ID] {
				continue
			}
			d := dist2(r.Pose, t.Pickup)
			if best == -1 || d < bestD {
				best, bestD = i, d
			}
		}
		if best == -1 {
			continue // no free robot left
		}
		used[free[best].ID] = true
		assignments = append(assignments, model.Assignment{
			TaskID:     t.ID,
			RobotID:    free[best].ID,
			AssignedAt: time.Now(),
		})
	}
	return assignments
}
