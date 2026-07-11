# Research proposal (seed)

> This is the seed that grows into your MEXT research plan and SOP. Keep it here in the repo — it signals research intent to anyone browsing, and doubles as your working draft.

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

## Expected contribution

An open, reproducible testbed and an empirical study of how allocation and MAPF interact under failure — a small but concrete step toward dependable large-scale warehouse fleets, and a foundation extensible to a master's thesis.
