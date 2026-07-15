package allocation

import (
	"math"
	"time"

	"github.com/yourname/warefleet/fleet_manager/internal/model"
)

// Auction implements market-based allocation: an epsilon-scaled Bertsekas
// auction. Bidders repeatedly bid on their cheapest item (cost + price);
// contested items get more expensive until the market clears. Converges to
// within n*epsilon of the optimal assignment.
//
// Why it exists next to Hungarian: the auction is naturally DECENTRALISABLE —
// each robot can compute its own bids from local state — which makes it the
// bridge to the centralized-vs-decentralized comparison in the study.
// Hungarian gives the exact optimum but requires a central solver.
type Auction struct{}

func (a *Auction) Name() string { return "auction" }

// epsilon trades optimality for convergence speed. Warehouse distances are
// metres-scale, so 1cm keeps the result within centimetres of optimal while
// bounding rounds to ~(max_cost/epsilon) per bidder.
const epsilon = 0.01

func (a *Auction) Assign(tasks []model.Task, robots []model.Robot) []model.Assignment {
	free := availableRobots(robots)
	tasks = sortedTasks(tasks)
	if len(free) == 0 || len(tasks) == 0 {
		return nil
	}

	// The SMALLER side bids (every bidder can win an item, so the auction
	// terminates); the larger side are the items being priced.
	robotsBid := len(free) <= len(tasks)
	var nBid, nItem int
	if robotsBid {
		nBid, nItem = len(free), len(tasks)
	} else {
		nBid, nItem = len(tasks), len(free)
	}
	cost := func(b, it int) float64 {
		if robotsBid {
			return math.Sqrt(dist2(free[b].Pose, tasks[it].Pickup))
		}
		return math.Sqrt(dist2(free[it].Pose, tasks[b].Pickup))
	}

	price := make([]float64, nItem)
	ownerOf := make([]int, nItem) // item -> bidder (-1 = unowned)
	itemOf := make([]int, nBid)   // bidder -> item (-1 = unassigned)
	for i := range ownerOf {
		ownerOf[i] = -1
	}
	queue := make([]int, 0, nBid)
	for b := 0; b < nBid; b++ {
		itemOf[b] = -1
		queue = append(queue, b)
	}

	for len(queue) > 0 {
		b := queue[0]
		queue = queue[1:]

		// find the bidder's best and second-best item at current prices
		best := -1
		bestVal, second := math.Inf(1), math.Inf(1)
		for it := 0; it < nItem; it++ {
			v := cost(b, it) + price[it]
			if v < bestVal {
				second = bestVal
				bestVal, best = v, it
			} else if v < second {
				second = v
			}
		}
		if best == -1 {
			continue
		}
		// raise the price by the bidder's margin (Bertsekas bid increment)
		inc := epsilon
		if !math.IsInf(second, 1) {
			inc += second - bestVal
		}
		price[best] += inc

		// take the item; the displaced previous owner re-enters the queue
		if prev := ownerOf[best]; prev != -1 {
			itemOf[prev] = -1
			queue = append(queue, prev)
		}
		ownerOf[best] = b
		itemOf[b] = best
	}

	assignments := make([]model.Assignment, 0, nBid)
	for b := 0; b < nBid; b++ {
		it := itemOf[b]
		if it == -1 {
			continue
		}
		var robotID, taskID string
		if robotsBid {
			robotID, taskID = free[b].ID, tasks[it].ID
		} else {
			robotID, taskID = free[it].ID, tasks[b].ID
		}
		assignments = append(assignments, model.Assignment{
			TaskID:     taskID,
			RobotID:    robotID,
			AssignedAt: time.Now(),
		})
	}
	return assignments
}
