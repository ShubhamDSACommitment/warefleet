// Package model holds the core domain types shared across the fleet manager.
package model

import "time"

// Point is a 2D position in the warehouse map frame (metres).
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Task is a unit of warehouse work.
type Task struct {
	ID        string    `json:"task_id"`
	Kind      string    `json:"kind"` // pick | transport | replenish
	Pickup    Point     `json:"pickup"`
	Dropoff   Point     `json:"dropoff"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

// RobotStatus enumerates the lifecycle states of a robot.
type RobotStatus string

const (
	StatusIdle     RobotStatus = "idle"
	StatusBusy     RobotStatus = "busy"
	StatusCharging RobotStatus = "charging"
	StatusError    RobotStatus = "error"
	StatusOffline  RobotStatus = "offline"
)

// Robot is the fleet manager's view of a single robot, updated from heartbeats.
type Robot struct {
	ID            string      `json:"robot_id"`
	Pose          Point       `json:"pose"`
	Status        RobotStatus `json:"status"`
	CurrentTaskID string      `json:"current_task_id"`
	Battery       float64     `json:"battery"`
	LastSeen      time.Time   `json:"-"` // for dropout detection
}

// Assignment is the output of the allocation layer.
type Assignment struct {
	TaskID     string    `json:"task_id"`
	RobotID    string    `json:"robot_id"`
	AssignedAt time.Time `json:"assigned_at"`
}

// PathPlan is the output of the MAPF layer: a conflict-free path for one robot.
type PathPlan struct {
	RobotID   string  `json:"robot_id"`
	TaskID    string  `json:"task_id"`
	Waypoints []Point `json:"waypoints"`
}
