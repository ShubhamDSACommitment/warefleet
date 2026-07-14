package allocation

import "github.com/yourname/warefleet/fleet_manager/internal/model"

// Auction implements market-based allocation: each robot "bids" on tasks based
// on its cost to serve them; tasks go to the lowest bidder. Naturally maps to a
// DECENTRALISED coordinator (robots compute their own bids), which is why it is
// the bridge to the centralized-vs-decentralized comparison in the study.
type Auction struct{}

func (a *Auction) Name() string { return "auction" }

func (a *Auction) Assign(tasks []model.Task, robots []model.Robot) []model.Assignment {
	// TODO(week5): sequential single-item auction:
	//   for each task (priority order):
	//     each free robot bids cost = travel cost to pickup (dist2)
	//     award task to lowest bidder; mark robot busy
	//   (extension: iterative/CBBA-style for multi-task bundles)
	_ = tasks
	_ = robots
	panic("Auction.Assign: implement in Week 5")
}
