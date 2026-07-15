# Worlds

`warehouse.world` — the Gazebo warehouse: parallel aisles, shelf blocks, and
pick/drop stations matching `benchmarks/scenarios/*.yaml`.

Build in Week 1–2. Start from a simple layout (2–3 aisles, a handful of shelves)
and grow it. Keep the station coordinates in sync with the scenario files so the
order feeder's pickup/dropoff points are reachable.

Tip: model shelves as static box obstacles; that's enough for navigation and MAPF.
