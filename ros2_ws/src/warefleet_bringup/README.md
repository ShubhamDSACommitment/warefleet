# warefleet_bringup

Launch files for WareFleet simulation.

## Phase 1 — one robot in our warehouse (`single_robot.launch.py`)

Reuses Nav2's proven TB3 simulation, pointed at **our** `worlds/warehouse.sdf`, running **SLAM live** (no pre-built map needed yet).

### Run (Ubuntu / ROS 2 Jazzy)
```bash
cd <warefleet>/ros2_ws
source /opt/ros/jazzy/setup.bash
colcon build
source install/setup.bash

export TURTLEBOT3_MODEL=waffle
ros2 launch warefleet_bringup single_robot.launch.py
```
Then in **RViz**: click **2D Pose Estimate** (place the robot), then **Nav2 Goal** → the robot plans a path and drives through the warehouse. ✅ = Phase 1 gate passed.

### ⚠️ Expect 1–2 debug iterations (Gazebo integration is finicky, and this was authored without a test run)
Before/if it errors, verify the reused launch matches your Nav2 version:
```bash
ls $(ros2 pkg prefix nav2_bringup)/share/nav2_bringup/launch/     # find the real tb3 sim launch filename
ros2 launch nav2_bringup tb3_simulation_launch.py --show-args      # confirm 'world', 'slam', 'headless' args exist
```
If the filename differs, update `TB3_SIM_LAUNCH` in `single_robot.launch.py`. If an arg name differs, update the `launch_arguments`. **Paste any error and we'll fix it together.**

### Fallback (if our world misbehaves)
Confirm the machinery still works with Nav2's default world, then reintroduce ours:
```bash
ros2 launch nav2_bringup tb3_simulation_launch.py slam:=True
```

## Next (Phase 2+)
`warehouse.launch.py` (stub) grows into the **multi-robot** bringup — N namespaced robots — once single-robot is solid.
