# WareFleet ROS 2 workspace

## Prerequisites

- Ubuntu 22.04 (ROS 2 Humble) or 24.04 (ROS 2 Jazzy)
- Nav2: `sudo apt install ros-$ROS_DISTRO-navigation2 ros-$ROS_DISTRO-nav2-bringup`
- Gazebo + bridge: `sudo apt install ros-$ROS_DISTRO-ros-gz`
- TurtleBot3 (for the starter robot model): `sudo apt install ros-$ROS_DISTRO-turtlebot3*`
- `slam_toolbox`: `sudo apt install ros-$ROS_DISTRO-slam-toolbox`

## Build

```bash
cd ros2_ws
colcon build --symlink-install
source install/setup.bash
```

## Packages

| Package | Type | Purpose |
|---|---|---|
| `warefleet_msgs` | ament_cmake | Custom messages (Task, RobotState, TaskAssignment, PathPlan) |
| `warefleet_bringup` | ament_cmake | Launch: Gazebo + N robots + Nav2 |
| `warefleet_nav` | ament_cmake | Nav2 params, warehouse map |
| `warefleet_agent` | ament_python | Per-robot task executor + order feeder |
| `warefleet_mapf` | ament_python | Reference MAPF planners (prioritized/CBS/PIBT) |

## Run

```bash
export TURTLEBOT3_MODEL=burger
ros2 launch warefleet_bringup warehouse.launch.py robots:=4
```

Build order matters: `warefleet_msgs` first (others depend on it). `colcon build`
handles this automatically.
