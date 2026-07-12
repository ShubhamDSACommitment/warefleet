# Research proposal (seed)

> This is the seed that grows into your MEXT research plan and SOP. Keep it here in the repo — it signals research intent to anyone browsing, and doubles as your working draft.

## Vision (theme)
*How do warehouse fleet systems behave under operational constraints, and how can we measure and improve them?* — explored in an open ROS 2 testbed informed by real deployments. The narrow, testable question below sits **under** this theme. ("Measure" = the reproductions H1/H2; "improve" = the one extension.)

## Working title

**Lifelong, dependable multi-robot coordination for large-scale warehouse AMR fleets: coupling task allocation with multi-agent path finding (TAPF) under continuous order streams and failures.**

## Problem

Warehouse throughput is increasingly limited not by individual robot capability but by **fleet-level coordination**: how tasks are allocated across robots and how their paths are deconflicted at scale (jointly, the **target-assignment-and-path-finding / TAPF** problem). Two gaps separate current research from deployment: (1) most joint allocation+MAPF results are **offline**, but real warehouses face an endless **online (lifelong)** stream of orders; and (2) robots **fail or drop off the network**, yet most coordination research assumes an idealized, always-available fleet. Closing both is fundamentally a dependable-distributed-systems problem.

## Research question

> How do task-allocation strategies and multi-agent path-finding (MAPF) methods interact to determine throughput and reliability in large-scale, **lifelong (online)** warehouse AMR fleets — and can a partly-decentralized, fault-tolerant coordinator sustain performance when robots fail?

## Sub-questions

1. How does the *choice* of allocation (greedy / Hungarian / auction) change the difficulty of the downstream MAPF problem?
2. What is the throughput/latency trade-off between complete MAPF (CBS) and scalable MAPF (PIBT) as fleet size grows?
3. How much throughput is lost, and how fast is it recovered, when a fraction of robots drop out — under centralized vs. decentralized coordination?

## Approach

- Build a reproducible warehouse simulation (ROS 2 + Nav2 + Gazebo) with a configurable fleet.
- Implement the allocation × coordination matrix in a distributed Go coordinator.
- Inject controlled robot failures and measure recovery.
- Benchmark on fixed order streams; report throughput, makespan, conflicts, utilization, recovery time.

## Why me

Industry experience with warehouse automation and ASRS fleets (Addverb, Takeoff) — I have operated the real systems this research abstracts, and bring the distributed-systems engineering (Go, MQTT, telemetry) that fleet coordination actually requires. WareFleet is my working prototype of these ideas.

## Fit with target labs (tune per professor before sending)

- **Défago / Coord Lab (Institute of Science Tokyo)** — emphasize dependable distributed coordination + MAPF (their core).
- **Miura (Toyohashi Univ. of Technology)** — emphasize mobile-robot navigation + the SLAM stretch module.
- **Ota (University of Tokyo)** — emphasize multi-agent systems + warehouse/manufacturing framing.

## Relationship to existing work (honest)
This is an **active, well-covered area** — not a gap I discovered. Open testbeds already study MAPF/fleet performance under realistic execution and failures: **LSMART** (kinodynamics, comm delays, execution uncertainty, fail policies; [arXiv 2602.15721](https://arxiv.org/abs/2602.15721)), **"It Takes Two to Tango"** (joint scheduling+MAPF; [arXiv 2602.13999](https://arxiv.org/pdf/2602.13999)), **WareRover**, and the "realistic settings" agenda ([arXiv 2404.16162](https://arxiv.org/abs/2404.16162)). WareFleet **builds openly on these** — it re-implements a compact slice in ROS 2/Nav2 to engage the problem hands-on and reproduce known findings. No claim of algorithmic novelty. See `docs/related-work.md`.

## Expected contribution

A compact, open, ROS 2-based reproduction that lets me engage this literature hands-on and reproduce the finding that allocation rankings shift under realistic execution — evidence I can build and reason about these systems. **It is a starting point, not a discovery**; the genuinely novel direction would be defined with a supervisor during the Master's. What I bring that the cited authors do not is real warehouse/ASRS operations experience + production distributed-systems engineering.
