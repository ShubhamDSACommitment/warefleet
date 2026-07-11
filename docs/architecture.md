# Architecture

WareFleet has three planes:

1. **Simulation plane** — Gazebo + N robots, each running ROS 2 / Nav2 (runs on host).
2. **Coordination plane** — the Go Fleet Manager: allocation + MAPF + fault-tolerant coordination (Docker).
3. **Observability plane** — MQTT → Prometheus → Grafana (Docker).

## System diagram

```
                          ┌───────────────────────────────────────────────┐
                          │            FLEET MANAGER  (Go service)          │
        orders  ────────► │                                                 │
    (scenario feeder)     │  ┌───────────┐  ┌────────────┐  ┌────────────┐ │
                          │  │ Task Queue│─►│ Allocation │─►│    MAPF     │ │
                          │  │ + intake  │  │ greedy /    │  │ prioritized│ │
                          │  │           │  │ hungarian / │  │ / cbs /    │ │
                          │  │           │  │ auction     │  │ pibt/lacam │ │
                          │  └───────────┘  └────────────┘  └─────┬──────┘ │
                          │  ┌──────────────────────────────┐     │        │
                          │  │ Coordinator (centralized OR   │◄────┘        │
                          │  │ decentralized) + robot health │              │
                          │  │ registry + dropout re-assign  │              │
                          │  └───────────────┬──────────────┘              │
                          └──────────────────┼─────────────────────────────┘
                                             │ ROS 2 (via rosbridge/DDS) + MQTT
              ┌──────────────────────────────┼──────────────────────────────┐
              │                              │                               │
        ┌─────▼─────┐                  ┌─────▼─────┐                   ┌─────▼─────┐
        │  Robot 1  │                  │  Robot 2  │      ...          │  Robot N  │
        │  Nav2     │                  │  Nav2     │                   │  Nav2     │
        │  +AMCL/   │                  │  agent    │                   │  agent    │
        │  SLAM     │                  │  node     │                   │  node     │
        └─────┬─────┘                  └─────┬─────┘                   └─────┬─────┘
              └───────────  Gazebo warehouse world (physics + sensors) ──────┘
                                             │
                                     MQTT telemetry (robot state, task events)
                                             │
                                   ┌─────────▼─────────┐      ┌──────────────┐
                                   │    Prometheus     │─────►│   Grafana    │
                                   │ (scrapes :2112)   │      │  dashboards  │
                                   └───────────────────┘      └──────────────┘
```

## Component responsibilities

### Fleet Manager (Go) — `fleet_manager/`
- **Task Queue / intake** — receives orders (from the scenario feeder or MQTT) as a **continuous, lifelong stream**, maintaining the pending-task set as new orders keep arriving.
- **Allocation** (`internal/allocation`) — assigns tasks to robots. Pluggable strategy: `greedy`, `hungarian`, `auction`.
- **MAPF** (`internal/coordination`) — plans collision-free paths for the assigned robots on the warehouse graph, **re-planning online as new tasks arrive (lifelong)**. Pluggable: `prioritized` (baseline), `cbs`, `pibt`, `lacam` (stretch).
- **Coordinator** — the control loop. Owns the robot health registry, detects **dropouts** (missed heartbeats), **re-allocates** their tasks, and runs either **centralized** (one planner) or **decentralized** (per-robot local negotiation) mode for the comparison study.
- **Metrics** (`internal/metrics`) — exposes Prometheus counters/histograms at `:2112/metrics`.

### ROS 2 workspace — `ros2_ws/src/`
- `warefleet_msgs` — custom messages: `Task`, `RobotState`, `TaskAssignment`, `PathPlan`.
- `warefleet_bringup` — launch files spawning Gazebo + N namespaced robots + Nav2 per robot.
- `warefleet_nav` — Nav2 params, warehouse map (`maps/`), costmap config.
- `warefleet_agent` — per-robot node: receives a `TaskAssignment` + `PathPlan`, drives Nav2 to waypoints, publishes `RobotState`; also the `order_feeder` CLI.
- `warefleet_mapf` — Python reference implementation of the MAPF planners (for prototyping/validation before/besides the Go version).

### Observability — `monitoring/`
- `mosquitto.conf` — MQTT broker config.
- `prometheus/prometheus.yml` — scrape config (targets the fleet manager).
- `grafana/` — provisioned datasource + dashboards (throughput, makespan, utilization, conflicts, active robots).

## Data flow (one task lifecycle)

1. Order enters the **Task Queue** (feeder or MQTT `warefleet/orders`).
2. **Allocation** assigns it to a robot → emits `TaskAssignment`.
3. **MAPF** computes a conflict-free path for that robot given all others → `PathPlan`.
4. The robot's **agent node** executes the path through Nav2.
5. The agent publishes `RobotState` + task events over **MQTT**.
6. The **Coordinator** updates the health registry; on dropout it re-queues the task.
7. **Prometheus** scrapes fleet metrics; **Grafana** renders KPIs live.

## Prerequisites (simulation plane, host)

- Ubuntu 22.04/24.04
- ROS 2 (Humble or Jazzy) + Nav2 + `turtlebot3` (or the provided diff-drive model)
- Gazebo (Fortress/Harmonic) with `ros_gz`
- `colcon`

Coordination + observability planes need only Docker + Docker Compose.
