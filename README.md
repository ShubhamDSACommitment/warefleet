# WareFleet

**An open-source multi-robot warehouse fleet-management stack: task allocation, multi-agent path finding (MAPF), and a fault-tolerant distributed coordinator — built by a warehouse-automation engineer.**

> Simulated in ROS 2 + Nav2 + Gazebo. Coordinated by a distributed Go fleet manager. Observed live in Grafana. One command to run the whole thing.

<!-- Replace with a 30s GIF: robots navigating the warehouse + the Grafana KPI board. This is the single most important asset in the repo. -->
![demo](docs/images/demo.gif)

---

## Why this project

Modern warehouses run fleets of autonomous mobile robots (AMRs). The hard part is rarely a single robot — it's **coordinating many of them**: which robot takes which task, and how they all move without deadlocking. WareFleet is a research-grade testbed for exactly that problem, studying two coupled layers:

1. **Task allocation** — assigning incoming pick/transport orders to robots (greedy, Hungarian, auction).
2. **Multi-Agent Path Finding (MAPF)** — routing all robots to their goals *without collisions or deadlock*, at scale (prioritized planning → CBS/PIBT-style solvers).

…and a third, systems dimension most academic testbeds ignore:

3. **Dependable distributed coordination** — the fleet manager keeps the fleet productive under **robot failures/dropouts**, and we compare **centralized vs. decentralized** coordination.

## Research question

> *How do task-allocation and multi-robot path-finding strategies interact to affect throughput and reliability in large-scale warehouse AMR fleets — and can a partly-decentralized, fault-tolerant coordinator maintain performance when robots fail?*

See [`docs/research-proposal.md`](docs/research-proposal.md).

## Architecture

See [`docs/architecture.md`](docs/architecture.md) for the full diagram and data flow. In one line:

```
Orders → Fleet Manager [allocation + MAPF + fault-tolerant coordinator] → ROS 2 → N × (Nav2 robot in Gazebo)
                              │
                              └── MQTT telemetry → Prometheus → Grafana KPIs
```

## Key features

- **N-robot warehouse simulation** (Gazebo world with aisles, shelves, pick/drop stations)
- **Per-robot autonomy** via ROS 2 **Nav2** (localization, planning, control)
- **Pluggable task allocation**: greedy · Hungarian (optimal assignment) · market/auction
- **MAPF coordination layer**: prioritized planning → conflict-based (CBS) / PIBT-style (stretch)
- **Distributed Fleet Manager (Go)**: task queue, assignment, conflict resolution, **robot-dropout handling**
- **MQTT** event/telemetry bus (mirrors real WMS↔robot comms)
- **Prometheus + Grafana** live KPIs: throughput, makespan, utilization, conflicts
- **Reproducible benchmarks**: compare strategies on identical scenarios; results in [`docs/results.md`](docs/results.md)
- **One-command bring-up** via Docker Compose

## Quick start

```bash
# 1. Bring up broker + monitoring + fleet manager
docker compose up -d

# 2. Build & launch the ROS 2 simulation (see ros2_ws/README for prerequisites)
make sim ROBOTS=4

# 3. Feed the fleet some orders
make orders SCENARIO=benchmarks/scenarios/peak_hour.yaml

# 4. Open the dashboards
#    Grafana:    http://localhost:3000  (admin/admin)
#    Prometheus: http://localhost:9090
```

Full setup, including ROS 2 / Gazebo prerequisites, is in [`docs/architecture.md`](docs/architecture.md) and [`ros2_ws/README.md`](ros2_ws/README.md).

## Results (preview)

Benchmark comparing allocation × coordination strategies on the same warehouse and order stream. _(Populated in Week 7 — see [`docs/results.md`](docs/results.md).)_

| Allocation | Coordination | Throughput (tasks/hr) | Avg. makespan (s) | Conflicts | Utilization |
|---|---|---|---|---|---|
| Greedy | Prioritized | _tbd_ | _tbd_ | _tbd_ | _tbd_ |
| Hungarian | Prioritized | _tbd_ | _tbd_ | _tbd_ | _tbd_ |
| Auction | CBS/PIBT | _tbd_ | _tbd_ | _tbd_ | _tbd_ |

## Design decisions

Why Go for the coordinator, why a separate MAPF layer, centralized vs. decentralized trade-offs, message schema choices — [`docs/design-decisions.md`](docs/design-decisions.md).

## Roadmap

Week-by-week build plan and stretch goals — [`docs/roadmap.md`](docs/roadmap.md).

## Tech stack

`ROS 2 (Humble/Jazzy)` · `Nav2` · `Gazebo` · `Go` · `Python (rclpy)` · `MQTT (Mosquitto)` · `Prometheus` · `Grafana` · `Docker Compose`

## About

Built by a software engineer with warehouse-automation / ASRS industry experience (Addverb, Takeoff), as a research prototype toward graduate study in multi-robot warehouse systems.
`GitHub: <you>` · `LinkedIn: <you>`

## License

MIT — see [LICENSE](LICENSE).
