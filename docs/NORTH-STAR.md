# 🧭 North Star — Project Focus & Scope Guard

> Read this before adding anything to WareFleet. If a feature doesn't serve the focus below, it's a stretch goal or it's cut.

## Positioning (canonical — say it in this order everywhere)

- **Vision (theme).** *How do warehouse fleet systems behave under operational constraints, and how can we measure and improve them?* — explored in an open ROS 2 testbed **informed by real deployments** (studied in simulation, not live warehouses).
- **Research goal (narrow, testable).** How do **task-allocation strategies** interact with **realistic execution** (congestion, delay, failures) to affect throughput/robustness — the setting LSMART leaves as *random* assignment + centralized.
- **Hypotheses (reproductions):**
  - **H1** — best-under-idealized ≠ best-under-realistic.
  - **H2** — telemetry-driven, failure-aware reallocation recovers throughput lost by static allocation.
- **The one extension ("improve"):** telemetry-driven failure-aware reallocation, grounded in LSMART's random-assignment + centralized assumptions.

> **Platform = vehicle; finding = contribution.** Component breadth is an engineering-capability signal, not the research claim. When asked "what's the research?", answer H1/H2 — not the component list.

> ⚠️ **Not novel — and that's fine for a Master's applicant.** This area is active and populated (LSMART, WareRover, "It Takes Two to Tango", the lifelong-MAPF realism papers). WareFleet is a **learning/demonstration/replication** platform built openly on that work; see [`related-work.md`](related-work.md). The genuine contribution is deferred to the thesis, with a supervisor. **Never present WareFleet as novel** — cite the prior work instead.

## The one focus
**Lifelong, fault-tolerant task-allocation × path-finding (TAPF) for warehouse AMR fleets.**

WareFleet exists to answer: *how do task-allocation and MAPF strategies interact to affect throughput and reliability in a **lifelong (online)** warehouse fleet — and can a fault-tolerant coordinator sustain performance when robots fail?*

Everything is judged by one question: **does it help prove that?**

## Definition of done (the MVP that proves the focus)
A continuous order stream **+** 3 allocation strategies (greedy / Hungarian / auction) **+** 1 MAPF planner (prioritized) **+** a dropout-recovery result **+** a benchmark table comparing them.

## Required core components
1. Lifelong order stream — `warefleet_agent/order_feeder.py`, `benchmarks/scenarios/`
2. Online task queue — `coordination/coordinator.go`
3. Allocation × 3 — `allocation/` (greedy ✓, hungarian, auction)
4. One MAPF planner — `coordination/planner.go` (`prioritized`), `warefleet_mapf/prioritized.py`
5. Coordinator loop coupling allocation + MAPF online — `coordinator.go`
6. Fault injection + dropout recovery — `coordinator.go` (`detectDropouts`)
7. Metrics + benchmark harness — `metrics/`, `benchmarks/run_benchmarks.py`
8. Minimal sim + N robots executing a path — `warefleet_bringup/`, `warefleet_nav/`, `warefleet_agent/agent_node.py`

## Supporting (keep light)
MQTT · Prometheus/Grafana (the results **table** matters more than the dashboard) · Docker Compose.

## Stretch / cut — NOT required for the focus
- CBS / PIBT / LaCAM (one extra planner only if ahead)
- SLAM (use a static map instead)
- Grid→AGV execution-gap analysis
- Centralized vs. decentralized comparison
- React web fleet map

## Drift test
If you have three half-finished MAPF algorithms and no lifelong benchmark, you drifted into a MAPF-algorithm project. Refocus: **the study is the allocation×MAPF comparison under lifelong operation + failure — not the planners themselves.**
