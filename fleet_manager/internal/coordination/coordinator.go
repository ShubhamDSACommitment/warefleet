// Package coordination runs the fleet control loop: it turns pending tasks into
// conflict-free, executable plans, and keeps the fleet productive under failures.
//
// This is the heart of WareFleet and the differentiator vs. a plain demo: it
// couples allocation + MAPF and adds dependability (dropout detection + task
// re-allocation) and a centralized/decentralized mode switch.
package coordination

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yourname/warefleet/fleet_manager/internal/allocation"
	"github.com/yourname/warefleet/fleet_manager/internal/metrics"
	"github.com/yourname/warefleet/fleet_manager/internal/model"
)

// Mode selects the coordination architecture under study.
type Mode string

const (
	Centralized   Mode = "centralized"
	Decentralized Mode = "decentralized"
)

// Publisher sends plans/assignments out to the robots (implemented by the MQTT client).
type Publisher interface {
	PublishAssignment(model.Assignment) error
	PublishPathPlan(model.PathPlan) error
}

// Coordinator owns fleet state and drives the control loop.
//
// Task lifecycle: tasks (pending) --assign--> inflight --heartbeat says done--> completed.
type Coordinator struct {
	mu     sync.RWMutex
	tasks  map[string]model.Task  // pending tasks
	robots map[string]model.Robot // last-known robot state

	inflight map[string]model.Task // assigned, not yet completed (by task ID)

	alloc     allocation.Strategy
	planner   Planner
	pub       Publisher
	mode      Mode
	heartbeat time.Duration // dropout threshold
}

func New(alloc allocation.Strategy, planner Planner, pub Publisher, mode Mode) *Coordinator {
	return &Coordinator{
		tasks:     make(map[string]model.Task),
		robots:    make(map[string]model.Robot),
		inflight:  make(map[string]model.Task),
		alloc:     alloc,
		planner:   planner,
		pub:       pub,
		mode:      mode,
		heartbeat: 5 * time.Second,
	}
}

// AddTask enqueues a new order (called from the MQTT orders subscription).
func (c *Coordinator) AddTask(t model.Task) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tasks[t.ID] = t
	log.Printf("order received: %s (%s), queue=%d", t.ID, t.Kind, len(c.tasks))
}

// UpdateRobot records the latest heartbeat and reacts to state transitions:
// busy->idle completes the in-flight task.
func (c *Coordinator) UpdateRobot(r model.Robot) {
	r.LastSeen = time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	prev, known := c.robots[r.ID]

	// Completion: the robot was working task X and now reports it gone.
	if known && prev.CurrentTaskID != "" && r.CurrentTaskID == "" {
		if t, ok := c.inflight[prev.CurrentTaskID]; ok && r.Status != model.StatusError {
			delete(c.inflight, t.ID)
			metrics.TasksCompleted.Inc()
			metrics.TaskMakespan.Observe(time.Since(t.CreatedAt).Seconds())
			log.Printf("task %s completed by %s (makespan %.1fs, inflight=%d)",
				t.ID, r.ID, time.Since(t.CreatedAt).Seconds(), len(c.inflight))
		}
	}

	c.robots[r.ID] = r
}

// Run is the control loop: allocate -> plan -> dispatch, and check for dropouts.
func (c *Coordinator) Run(ctx context.Context, tick time.Duration) {
	t := time.NewTicker(tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.detectDropouts()
			c.step()
		}
	}
}

// step runs one allocation + planning + dispatch cycle.
func (c *Coordinator) step() {
	c.mu.Lock()
	tasks := mapValues(c.tasks)
	robots := mapValues(c.robots)
	c.mu.Unlock()

	if len(tasks) == 0 {
		return
	}

	assignments := c.alloc.Assign(tasks, robots)
	if len(assignments) == 0 {
		return
	}

	// TODO(week5): build the MAPF problem from assignments + robot poses,
	// call c.planner.Plan(...) for conflict-free PathPlans, publish both.
	for _, a := range assignments {
		if err := c.pub.PublishAssignment(a); err != nil {
			log.Printf("publish assignment %s: %v", a.TaskID, err)
			continue
		}
		log.Printf("assigned task %s -> robot %s", a.TaskID, a.RobotID)
		c.mu.Lock()
		if t, ok := c.tasks[a.TaskID]; ok {
			c.inflight[a.TaskID] = t
			delete(c.tasks, a.TaskID)
		}
		// Mark the robot busy immediately rather than waiting for its next
		// heartbeat: otherwise a task arriving within the heartbeat interval
		// would be assigned to the same robot, rejected by the agent, and lost.
		if r, ok := c.robots[a.RobotID]; ok {
			r.Status = model.StatusBusy
			r.CurrentTaskID = a.TaskID
			c.robots[a.RobotID] = r
		}
		c.mu.Unlock()
	}
}

// detectDropouts marks robots that missed their heartbeat window as offline and
// re-queues their in-flight task. This is the dependability KPI in the study (H2).
func (c *Coordinator) detectDropouts() {
	// TODO(H2): for each robot, if now-LastSeen > heartbeat and status != offline:
	//   mark offline, re-queue its in-flight task, record recovery-time metrics.
}

func mapValues[T any](m map[string]T) []T {
	out := make([]T, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
