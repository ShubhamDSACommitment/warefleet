# Maps

Put the warehouse occupancy map here:

- `warehouse.yaml` — Nav2 map metadata (resolution, origin, thresholds)
- `warehouse.pgm` — the occupancy grid image

Generate it in Week 2 by driving one robot around `sim/worlds/warehouse.world`
with `slam_toolbox`, then saving:

```bash
ros2 run nav2_map_server map_saver_cli -f warehouse
```

Commit both files here so the map is reproducible.
