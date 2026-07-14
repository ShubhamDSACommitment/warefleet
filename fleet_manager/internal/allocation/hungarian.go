package allocation

import "github.com/yourname/warefleet/fleet_manager/internal/model"

// Hungarian computes the cost-minimising one-shot assignment of tasks to robots
// via the Kuhn-Munkres algorithm on a cost matrix of robot->task travel cost.
//
// vs. Greedy: greedy is locally optimal per task; Hungarian is globally optimal
// for the batch. Comparing the two quantifies the value of optimal assignment.
type Hungarian struct{}

func (h *Hungarian) Name() string { return "hungarian" }

func (h *Hungarian) Assign(tasks []model.Task, robots []model.Robot) []model.Assignment {
	// TODO(week5):
	//   1. free := availableRobots(robots)
	//   2. build cost[i][j] = travel cost robot i -> task j pickup (dist2)
	//   3. pad to a square matrix (dummy rows/cols) if len(free) != len(tasks)
	//   4. run Kuhn-Munkres to get the min-cost perfect matching
	//   5. emit Assignment for each real (robot, task) pair in the matching
	//
	// Prototype/validate the algorithm in warefleet_mapf tests, or unit-test here.
	_ = tasks
	_ = robots
	panic("Hungarian.Assign: implement in Week 5")
}
