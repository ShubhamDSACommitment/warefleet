#!/usr/bin/env python3
"""WareFleet benchmark harness.

Sweeps the allocation x coordination x fleet-size matrix on fixed scenarios,
collects KPIs from Prometheus, and writes a results table + plots. This is what
turns the project into a study (docs/results.md).

Usage:
    python3 benchmarks/run_benchmarks.py --scenarios benchmarks/scenarios --out benchmarks/results

Plan (Week 7):
  for allocation in [greedy, hungarian, auction]:
    for coordination in [prioritized, cbs, pibt]:
      for n in [2, 4, 8, 16]:
        - restart fleet-manager with these env vars
        - launch sim with n robots
        - stream the scenario via order_feeder
        - wait for completion
        - query Prometheus for throughput/makespan/conflicts/utilization
        - append a row to results.csv
  then render plots (matplotlib) into --out.
"""
import argparse

ALLOCATIONS = ["greedy", "hungarian", "auction"]
COORDINATIONS = ["prioritized", "cbs", "pibt"]
FLEET_SIZES = [2, 4, 8, 16]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--scenarios", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--prometheus", default="http://localhost:9090")
    args = ap.parse_args()

    print("WareFleet benchmark matrix:")
    for a in ALLOCATIONS:
        for c in COORDINATIONS:
            for n in FLEET_SIZES:
                print(f"  [ ] allocation={a:10s} coordination={c:12s} robots={n}")
    print("\nTODO(week7): implement run loop + Prometheus queries + plotting.")


if __name__ == "__main__":
    main()
