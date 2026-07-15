package allocation

import (
	"math"
	"time"

	"github.com/yourname/warefleet/fleet_manager/internal/model"
)

// Hungarian computes the cost-minimising one-shot assignment of tasks to robots
// via the Kuhn-Munkres algorithm on a cost matrix of robot->task travel cost.
//
// vs. Greedy: greedy is locally optimal per task; Hungarian is globally optimal
// for the batch. Comparing the two quantifies the value of optimal assignment.
//
// Cost is true euclidean distance (not squared): the algorithm minimises the
// SUM of costs, and summing squared distances would over-penalise long legs
// and yield a different (wrong) optimum.
type Hungarian struct{}

func (h *Hungarian) Name() string { return "hungarian" }

func (h *Hungarian) Assign(tasks []model.Task, robots []model.Robot) []model.Assignment {
	free := availableRobots(robots)
	tasks = sortedTasks(tasks)
	if len(free) == 0 || len(tasks) == 0 {
		return nil
	}

	// Square cost matrix padded with zero-cost dummies (dropped afterwards).
	n := max(len(free), len(tasks))
	cost := make([][]float64, n)
	for i := range cost {
		cost[i] = make([]float64, n)
		for j := range cost[i] {
			if i < len(free) && j < len(tasks) {
				cost[i][j] = math.Sqrt(dist2(free[i].Pose, tasks[j].Pickup))
			}
		}
	}

	rowOfCol := kuhnMunkres(cost) // rowOfCol[j] = robot index assigned to task j

	assignments := make([]model.Assignment, 0, min(len(free), len(tasks)))
	for j := 0; j < len(tasks); j++ {
		i := rowOfCol[j]
		if i < len(free) { // skip dummy robots
			assignments = append(assignments, model.Assignment{
				TaskID:     tasks[j].ID,
				RobotID:    free[i].ID,
				AssignedAt: time.Now(),
			})
		}
	}
	return assignments
}

// kuhnMunkres solves the square min-cost assignment problem in O(n^3) using
// the shortest-augmenting-path formulation with row/column potentials.
// Returns rowOfCol: for each column j, the row assigned to it.
func kuhnMunkres(cost [][]float64) []int {
	n := len(cost)
	inf := math.Inf(1)
	u := make([]float64, n+1) // row potentials (1-indexed)
	v := make([]float64, n+1) // column potentials (1-indexed)
	p := make([]int, n+1)     // p[j] = row matched to column j (0 = none)
	way := make([]int, n+1)

	for i := 1; i <= n; i++ {
		p[0] = i
		j0 := 0
		minv := make([]float64, n+1)
		used := make([]bool, n+1)
		for j := range minv {
			minv[j] = inf
		}
		for {
			used[j0] = true
			i0, j1, delta := p[j0], 0, inf
			for j := 1; j <= n; j++ {
				if used[j] {
					continue
				}
				cur := cost[i0-1][j-1] - u[i0] - v[j]
				if cur < minv[j] {
					minv[j] = cur
					way[j] = j0
				}
				if minv[j] < delta {
					delta = minv[j]
					j1 = j
				}
			}
			for j := 0; j <= n; j++ {
				if used[j] {
					u[p[j]] += delta
					v[j] -= delta
				} else {
					minv[j] -= delta
				}
			}
			j0 = j1
			if p[j0] == 0 {
				break
			}
		}
		// augment along the found path
		for {
			j1 := way[j0]
			p[j0] = p[j1]
			j0 = j1
			if j0 == 0 {
				break
			}
		}
	}

	rowOfCol := make([]int, n)
	for j := 1; j <= n; j++ {
		rowOfCol[j-1] = p[j] - 1
	}
	return rowOfCol
}
