# Results

> Populated in Week 7. This file is what turns WareFleet from "a project" into "a study." Treat it like a mini paper: setup → results → discussion.

## Experimental setup

- **World:** `sim/worlds/warehouse.world` (aisles: TBD, shelves: TBD, pick/drop stations: TBD)
- **Fleet sizes:** N ∈ {2, 4, 8, 16}
- **Order stream:** `benchmarks/scenarios/peak_hour.yaml` (fixed seed for reproducibility)
- **Hardware:** TBD
- **Each config run:** TBD repetitions; report mean ± std.

## Metrics

| Metric | Definition |
|---|---|
| Throughput | Completed tasks per simulated hour |
| Makespan | Time from order intake to completion (avg) |
| Conflicts | Path conflicts requiring re-plan |
| Utilization | Fraction of time robots are on-task |
| Recovery time | Time to restore throughput after a dropout |
| Execution gap | Divergence between planned (grid/MAPF) cost and realized (Nav2 execution) cost |

> Primary regime is **lifelong/online**: metrics are steady-state rates over a continuous order stream, not single-batch totals. Report throughput as tasks/hour at steady state.

## Results: allocation × coordination

_(fill after Week 7)_

| Allocation | Coordination | N | Throughput | Makespan (s) | Conflicts | Utilization |
|---|---|---|---|---|---|---|
| Greedy | Prioritized | 8 | – | – | – | – |
| Hungarian | Prioritized | 8 | – | – | – | – |
| Auction | Prioritized | 8 | – | – | – | – |
| Hungarian | CBS | 8 | – | – | – | – |
| Hungarian | PIBT | 8 | – | – | – | – |

## Results: failure injection

_(fill after stretch goal)_

| Coordination mode | Dropout % | Throughput loss | Recovery time (s) |
|---|---|---|---|
| Centralized | 25% | – | – |
| Decentralized | 25% | – | – |

## Plots

_(embed from `benchmarks/results/`)_

- `throughput_vs_fleetsize.png`
- `makespan_by_strategy.png`
- `recovery_after_dropout.png`

## Discussion

_(2–4 paragraphs: which strategy won and why; how allocation choice affected MAPF difficulty; the centralized/decentralized trade-off; limitations; future work.)_
