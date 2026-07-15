# Related Work & Honest Positioning

> **WareFleet is a learning + demonstration + replication platform — not a novel research contribution.** Fleet-level MAPF, task allocation, and their evaluation under realistic execution + failures is an **active, well-populated area**. This file records the work WareFleet builds on, so the project (and any email/proposal) never overclaims.

## The area is already well-covered

**Realistic-execution testbeds (this is essentially WareFleet's stated purpose):**
- **LSMART** — *Lifelong Scalable Multi-Agent Realistic Testbed*: an open-source simulator to evaluate any MAPF algorithm in a fleet-management system under **kinodynamics, communication delays, and execution uncertainties**, with pluggable **fail policies**, scaling to 1000 robots. [github.com/smart-mapf/lifelong-smart](https://github.com/smart-mapf/lifelong-smart) · [arXiv 2602.15721](https://arxiv.org/abs/2602.15721). *(From the smart-mapf group — Okumura/Défago's ecosystem.)*
- **"Scaling Lifelong MAPF to More Realistic Settings: Research Challenges and Opportunities"** — names the realism gap as the agenda. [arXiv 2404.16162](https://arxiv.org/abs/2404.16162)

**Joint scheduling + MAPF + failure (the "allocation × path-finding under failure" angle):**
- **"It Takes Two to Tango: A Holistic Simulator for Joint Order Scheduling and MAPF in Robotic Warehouses"** [arXiv 2602.13999](https://arxiv.org/pdf/2602.13999)
- **WareRover** — scheduler + MAPF joint loop with a failure-simulation-and-recovery module.
- **"MAPF with Real Robot Dynamics and Interdependent Tasks for Automated Warehouses"** [arXiv 2408.14527](https://arxiv.org/pdf/2408.14527)
- **DeepFleet** (Amazon, 2025) — multi-agent foundation models for warehouse robots [arXiv 2508.08574](https://arxiv.org/html/2508.08574v1)

**Failure-aware task reallocation (MRTA under failures):**
- "Fault-Tolerant Framework for Dynamic Task Reassignment in Multi-Robot Systems" [MDPI](https://www.mdpi.com/2673-4591/120/1/22)
- "Multi-Robot Preemptive Task Scheduling with Fault Recovery" [NCBI](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC8512959/)

**Allocation-for-throughput comparison (academic + commercial):**
- "Improving Automatic Warehouse Throughput by Optimizing Task Allocation… Simulation" [MDPI](https://www.mdpi.com/2673-4052/2/3/7)
- AnyLogic large-scale AMR fleet case studies [AnyLogic](https://www.anylogic.com/resources/case-studies/optimizing-large-scale-amr-fleet-operations/)

**Foundational MAPF / MRTA:**
- PIBT (Okumura et al., AIJ 2022); LaCAM (Okumura, AAAI 2023); CBS (Sharon et al., 2015); Gerkey & Matarić MRTA taxonomy (2004).

## How WareFleet relates (honestly)
WareFleet **re-implements a small slice** of the above in ROS 2 + Nav2 to:
1. **Learn** the stack hands-on,
2. **Demonstrate** that I can build and reason about exactly the systems these labs study,
3. **Engage** the literature (reproduce a known finding, e.g. that allocation rankings shift under realistic execution).

**What is genuinely distinctive is about *me*, not the research:** real warehouse/ASRS operations experience + production distributed-systems/observability engineering (Go, MQTT, telemetry). No claim of algorithmic novelty.

## Rule
- **Do not claim novelty.** Cite this prior work; position WareFleet as a demonstration built openly on it.
- The **novel contribution is deferred to the Master's thesis**, defined with a supervisor.
- If a fresher angle emerges (e.g. telemetry-driven reallocation on real ROS 2/Nav2 execution vs. grid sims), treat it as *a direction to discuss with a supervisor* — assume partial prior work until verified.
