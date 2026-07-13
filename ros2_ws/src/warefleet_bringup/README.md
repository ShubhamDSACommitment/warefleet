# warefleet_bringup

Launch files for WareFleet simulation.

## Phase 1 — one robot in our warehouse (`single_robot.launch.py`)

Reuses Nav2's proven TB3 simulation, pointed at **our** `worlds/warehouse.sdf`, running **SLAM live** (no pre-built map needed yet).

### Run (Ubuntu / ROS 2 Jazzy) — ✅ verified 2026-07-13
```bash
cd <warefleet>/ros2_ws
source /opt/ros/jazzy/setup.bash
colcon build
source install/setup.bash

ros2 launch warefleet_bringup single_robot.launch.py
```
Then in **RViz**: click **Nav2 Goal** → the robot plans a path and drives through the
warehouse. No **2D Pose Estimate** needed — SLAM is on, so the robot localizes from
where it spawns. ✅ = Phase 1 gate passed.

Notes from the verified run:
- `TURTLEBOT3_MODEL` is **not** needed — Jazzy's `nav2_bringup` uses its own
  `nav2_minimal_tb3_sim` waffle, not the `turtlebot3_gazebo` models.
- The world **must load `gz-sim-imu-system`** (now in `warehouse.sdf`) or the robot's
  IMU never publishes.
- Headless smoke test (no GUI):
  `ros2 launch warefleet_bringup single_robot.launch.py headless:=True use_rviz:=False`
  then send a goal from a second terminal:
  ```bash
  ros2 action send_goal /navigate_to_pose nav2_msgs/action/NavigateToPose \
    "{pose: {header: {frame_id: map}, pose: {position: {x: 4.0, y: 0.5}, orientation: {w: 1.0}}}}"
  ```
  (Map frame ≈ world frame shifted by the spawn pose (-2, -0.5); goal (4.0, 0.5) = world (2, 0), mid-aisle.)

### Fallback (if our world misbehaves)
Confirm the machinery still works with Nav2's default world, then reintroduce ours:
```bash
ros2 launch nav2_bringup tb3_simulation_launch.py slam:=True
```

## Next (Phase 2+)
`warehouse.launch.py` (stub) grows into the **multi-robot** bringup — N namespaced robots — once single-robot is solid.
