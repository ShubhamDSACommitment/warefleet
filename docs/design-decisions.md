# Design decisions

This document explains *why* WareFleet is built the way it is. For a reviewer (or a professor), this is the most revealing file in the repo — it shows engineering judgment, not just working code.

## 1. Why separate allocation from path finding?

Allocation ("who does what") and MAPF ("how they move without colliding") are often conflated. Separating them lets us **study their interaction independently** — the central research question. A good allocation can be ruined by poor path coordination, and vice versa. Two pluggable layers → a clean experiment matrix (allocation × coordination).

## 2. Why Go for the Fleet Manager (not Python)?

- The coordinator is a **long-running, concurrent, networked service** — Go's goroutines/channels model the "many robots, many events, health monitoring" workload naturally.
- It mirrors how real warehouse fleet orchestrators are built (services, not scripts) — the project should read as *industrial software*, not a homework notebook.
- It plays to existing backend/distributed-systems strength.
- Python (`warefleet_mapf`) is kept for **algorithm prototyping and validation** — fast to iterate, easy to unit-test a planner against known MAPF instances, then port the validated logic to Go.

## 3. Why a distributed-systems framing (fault tolerance, centralized vs. decentralized)?

Most academic MAPF testbeds assume perfect, always-on robots. Real fleets don't work that way — robots stall, drop off the network, and get pulled for charging. Modeling **robot dropout + re-allocation** and comparing **centralized vs. decentralized** coordination:

- makes the study *realistic* and *novel-adjacent*, and
- aligns with dependable-distributed-systems research (a direct bridge to target labs such as the Coord Lab at Institute of Science Tokyo).

## 4. Why MQTT for telemetry?

It's the de-facto messaging layer between warehouse management systems and robots in industry. Using it (rather than only ROS topics) keeps the architecture honest to production and decouples the observability plane from ROS/DDS.

## 5. Why Nav2 rather than a custom local planner?

Reinventing single-robot navigation adds no research value and burns the timeline. Nav2 is the modern, credible standard; WareFleet's contribution is the **fleet layer on top**, not the per-robot stack.

## 6. Coordination strategies — the experiment matrix

| Allocation | Coordination | Rationale |
|---|---|---|
| Greedy | Prioritized | Cheapest baseline; establishes the floor |
| Hungarian | Prioritized | Optimal one-shot assignment vs. greedy |
| Auction | Prioritized | Decentralized-friendly assignment |
| (best allocation) | CBS | Optimal/complete conflict resolution (stretch) |
| (best allocation) | PIBT | Scalable, near-real-time MAPF (stretch) |
| (best allocation) | LaCAM | Fast, near-optimal, lifelong-capable MAPF (stretch; the reference line for MAPF-focused labs) |

## 7. Metrics we optimize / report

- **Throughput** — completed tasks/hour (primary business KPI).
- **Makespan / avg task completion time** — latency per task.
- **Conflicts / re-plans** — coordination quality.
- **Robot utilization** — fleet efficiency.
- **Recovery time after dropout** — dependability KPI (the distributed-systems angle).

## 8. Scope guardrails (what we deliberately do NOT do)

- No real hardware — simulation only (timeline).
- No custom perception/SLAM in the core (SLAM module is a stretch goal for a specific lab pitch).
- MAPF core target is **prioritized planning**; CBS/PIBT/LaCAM are stretch. Prioritized planning alone is enough to produce a meaningful comparison and ship on time.

## 9. Lifelong / online operation is the target regime

Real warehouses never stop taking orders. WareFleet therefore treats coordination as **lifelong (online)** — tasks arrive continuously and the coordinator re-allocates + re-plans on the fly — not as a one-shot offline batch. The offline case is only a baseline. This framing (a) matches real deployments (and my Addverb experience), and (b) aligns with the research frontier (e.g. lifelong MAPF / LaCAM). Concretely: the order feeder streams tasks over time; the coordinator's control loop runs continuously; metrics are measured as steady-state rates, not single-batch totals.

## 10. We evaluate the grid-plan → real-AGV execution gap

MAPF plans on an idealized time-stepped grid; real robots execute continuously through Nav2 with acceleration limits and delays. A distinctive contribution of WareFleet is to **measure the divergence** between the planned (grid) cost and the realized (executed) cost — the gap that separates "MAPF on paper" from a working fleet. This is why WareFleet executes plans through Nav2 in Gazebo rather than only simulating on a grid.
