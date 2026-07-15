# WareFleet

**An open-source research platform for studying autonomous warehouse fleet orchestration under realistic execution conditions.**

> Built on ROS 2 + Nav2 + Gazebo, coordinated by a distributed Go fleet manager, observed live in Grafana. One command to run the whole thing.

## Positioning

- **Vision (theme).** *How do warehouse fleet systems behave under operational constraints, and how can we measure and improve them?* — WareFleet is the open-source **ROS 2 testbed** I use to explore this, **informed by real warehouse deployments** (studied in simulation, not on live warehouses).
- **Research goal (narrow, testable).** How do **task-allocation strategies** interact with **realistic execution** (congestion, delay, failures) to affect fleet throughput and robustness — the setting LSMART leaves as *random* assignment + centralized control.
- **Hypotheses (reproductions).**
  - **H1** — the allocation strategy best under *idealized* simulation is **not** necessarily best under *realistic* execution.
  - **H2** — telemetry-driven, failure-aware reallocation recovers throughput lost by static allocation.
- **The one extension ("improve").** A telemetry-driven, failure-aware reallocation policy — grounded in LSMART's *random-assignment + centralized* assumptions and the "evolving systems" challenge.

> The **platform is the vehicle; the finding is the contribution.** The component breadth (fleet manager, allocation engine, MAPF layer, MQTT, metrics) is evidence of engineering capability — the research claim is **H1/H2**, not the parts.

> ⚠️ **Not a novelty claim.** This is an active, well-covered area (e.g. the open-source [LSMART](https://github.com/smart-mapf/lifelong-smart) testbed, and joint scheduling+MAPF+failure simulators). WareFleet is a **learning + demonstration + replication** platform built openly on that prior work — see [`docs/related-work.md`](docs/related-work.md). What's distinctive is the *builder's* warehouse-industry + distributed-systems background, not the algorithms.

<!-- Replace with a 30s GIF: robots navigating the warehouse + the Grafana KPI board. This is the single most important asset in the repo. -->
![demo](docs/images/demo.gif)

---

## Why this project

Modern warehouses run fleets of autonomous mobile robots (AMRs). The hard part is rarely a single robot — it's **coordinating many of them**: which robot takes which task, and how they all move without deadlocking. WareFleet is a research-grade testbed for exactly that problem, studying two coupled layers:

1. **Task allocation** — assigning incoming pick/transport orders to robots (greedy, Hungarian, auction).
2. **Multi-Agent Path Finding (MAPF)** — routing all robots to their goals *without collisions or deadlock*, at scale (prioritized planning → CBS/PIBT-style solvers).

…and a third, systems dimension most academic testbeds ignore:

3. **Dependable distributed coordination** — the fleet manager keeps the fleet productive under **robot failures/dropouts**, and we compare **centralized vs. decentralized** coordination.

> 🧭 **Focus:** this project has one North Star — *lifelong, fault-tolerant task-allocation × path-finding*. See [`docs/NORTH-STAR.md`](docs/NORTH-STAR.md) before adding scope.

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
- **MAPF coordination layer** (lifelong/online): prioritized planning → CBS / PIBT / LaCAM (stretch)
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

> Current measured results (both tiers, mean ± std over seeded order streams): [`docs/results.md`](docs/results.md).

## Portable development — no ROS required (idealized tier)

The **idealized tier is pure Go + stdlib Python** and runs on any machine
(macOS included) — only the realistic tier (Gazebo + Nav2) needs the Ubuntu/ROS box.

```bash
git clone https://github.com/ShubhamDSACommitment/warefleet && cd warefleet
# prerequisites: Go 1.22+ and Python 3 with PyYAML
#   macOS:  brew install go python3 && pip3 install pyyaml
cd fleet_manager && go test ./... && cd ..              # all unit tests (allocators, A*, PIBT)
python3 benchmarks/run_benchmarks.py --tier idealized \
  --robots 2,8,24 --orders 19 --rate 4 \
  --planners prioritized,pibt --repeat 5                # full benchmark matrix in seconds
```

Everything strategy/planner-related (greedy, Hungarian, auction, space-time A*,
prioritized planning, PIBT, gridsim, the benchmark harness, `docs/results.md`)
is developable and testable this way. Gazebo verification of changes waits for
the ROS machine.

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
