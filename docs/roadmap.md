# Roadmap — Phased Build

Ordered by **dependency + risk-retirement** (do the scary unknowns first). Aligned to the North Star ([`NORTH-STAR.md`](NORTH-STAR.md)) and the research question:

> *Does the ranking of task-allocation strategies change once realistic execution (congestion, delay, failures) is modeled — and can a telemetry-driven, failure-aware reallocation policy recover throughput that static allocation loses?*

**Critical-path logic:** the risky unknowns are Layer 1–2 (ROS 2 / Nav2 / MAPF) — front-loaded. Layer 3 (MQTT / telemetry / metrics) is the builder's strength and slots in cheaply later.

Each phase lists its **exit criterion** (when to move on) and the **risk it retires**.

---

## Phase 0 — Setup & first motion · Days 1–3
- **Build:** install ROS 2 + Gazebo + Nav2 + TurtleBot3; one robot driving via teleop; hand-make a **static warehouse map** (no SLAM).
- **Components:** `sim/worlds/warehouse.world`, `ros2_ws/src/warefleet_nav/maps/`.
- **Exit when:** one robot drives in Gazebo; map loads.
- **Retires:** toolchain/environment risk.

## Phase 1 — Single-robot navigation · Week 1
- **Build:** Nav2 drives one robot to goal poses; `agent_node` accepts a goal, reports arrival.
- **Components:** `warefleet_nav/config/nav2_params.yaml`, `warefleet_agent/agent_node.py`.
- **Exit when:** publish a goal → robot navigates there autonomously.
- **Retires:** the biggest ROS unknown ("can I make Nav2 work at all").

## Phase 2 — Multi-robot + telemetry backbone · Week 2
- **Build:** spawn N namespaced robots each taking goals; MQTT `RobotState` heartbeats flowing. *(Builder's home turf — fast.)*
- **Components:** `warefleet_bringup/launch/warehouse.launch.py`, `fleet_manager/internal/mqtt/`.
- **Exit when:** N robots move; live state visible over MQTT.
- **Retires:** multi-robot plumbing + the telemetry spine the study depends on.

## Phase 3 — Lifelong pipeline + Fleet Manager v1 · Week 3 ⭐ first full loop
- **Build:** `order_feeder` streams tasks continuously (lifelong); coordinator queue + **greedy** allocation → assignment → robot executes.
- **Components:** `warefleet_agent/order_feeder.py`, `coordination/coordinator.go`, `allocation/greedy.go`.
- **Exit when:** an order enters → a robot completes it, unattended, on a stream.
- **Retires:** "is the pipeline real?" — now it's a working system, not parts.

## Phase 4 — One MAPF planner + second allocation policy · Week 4
- **Build:** **prioritized-planning MAPF** (space-time A* + reservation) — prototype in Python, port to Go. Add a second allocation policy (prefer the **telemetry-aware** one over Hungarian).
- **Components:** `warefleet_mapf/prioritized.py`, `coordination/planner.go`, `allocation/`.
- **Exit when:** robots follow deconflicted paths; allocation policy switchable by config.
- **Retires:** coordination correctness. **Stop at one planner** (North Star).

## Phase 5 — The realism harness · Week 5 ⭐ the research instrument
- **Build:** three controllable realism knobs — (1) **congestion** (density/aisle width), (2) **execution delay/uncertainty** (idealized grid vs. real Nav2 timing), (3) **failure injection** (dropout). Instrument Prometheus metrics.
- **Components:** `benchmarks/scenarios/`, `coordinator.go` (`detectDropouts`), `metrics/metrics.go`.
- **Exit when:** same order stream runs across an "idealized → realistic" sweep.
- **Retires:** "can I actually run the experiment?"

## Phase 6 — Telemetry-driven reallocation · Week 6 ⭐ the differentiator
- **Build:** coordinator consumes live failure/congestion telemetry to **reassign** tasks (failure-aware). The part only this background can build credibly.
- **Components:** `coordinator.go` (reallocation on dropout), `mqtt/` + `metrics/`.
- **Exit when:** kill a robot mid-task → task reassigned → throughput visibly recovers.
- **Retires:** the novel contribution itself.

## Phase 7 — Benchmark + the finding · Week 7 ⭐ turns it into research
- **Build:** run allocation-policy × realism-factor matrix. Two money plots: (a) does the allocation ranking change idealized → realistic? (b) does telemetry-aware reallocation recover post-failure throughput? Fill `docs/results.md`.
- **Exit when:** a **claim with evidence** exists (either outcome is a result).
- **Retires:** "is there a finding?" — demo vs. thesis.

## Phase 8 — Demo, narrative, outreach · Week 8
- **Build:** 3–4 min demo video; README polish; proposal updated with preliminary results; send professor emails.
- **Exit when:** repo public + presentable; first emails sent.

---

## Application-MVP line
You do **not** need all 8 phases before emailing. **Through Phase 6 with preliminary numbers (~70%)** is enough for a strong, specific email (reference a working telemetry-driven reallocation demo + "early results suggest X"). Finish Phase 7 while awaiting replies. Emailing at 100% risks missing the outreach window.

## De-risking escape hatch (decide by end of Week 1)
If Gazebo/Nav2 threatens the timeline, **answer the research question on a 2-D grid simulator** (pure Python/Go; model delay/failure abstractly) — the RQ does **not** depend on Gazebo. Then Gazebo/Nav2 becomes the credibility + demo layer, added later. If Phase 1 slips past ~10 days, fork a grid-sim track in parallel so the *science* isn't hostage to ROS tooling.

## Stays cut (North Star discipline)
SLAM · CBS/PIBT/LaCAM (one planner only) · grid→AGV deep dive · decentralized mode · React map · arXiv writeup — all stretch, only if ahead after Phase 7.

## Learning order (only what phases need)
A* → ROS 2 basics → TF/URDF (minimal) → Gazebo → Nav2 → multi-robot namespacing → fleet manager → allocation → prioritized MAPF → metrics → benchmark. Concepts: see `docs/LEARNING.md`.
