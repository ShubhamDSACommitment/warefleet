"""WareFleet Phase 1 — one robot in our warehouse, navigating.

Strategy (see docs/focus-and-framing.md): DON'T reinvent the spawn/bridge/Nav2
wiring — reuse Nav2's proven TB3 simulation launch, but point it at OUR
warehouse world and run SLAM live (so no pre-built map is needed yet).

Run:
  ros2 launch warefleet_bringup single_robot.launch.py

Then in RViz: "Nav2 Goal" -> watch it plan + drive. (No "2D Pose Estimate" needed:
SLAM is on, the robot localizes from its spawn pose. TURTLEBOT3_MODEL is not used
on Jazzy — nav2_bringup spawns its own nav2_minimal_tb3_sim waffle.)

Verified working on ROS 2 Jazzy, 2026-07-13 (see README for the headless smoke test).
"""
import os
from ament_index_python.packages import get_package_share_directory
from launch import LaunchDescription
from launch.actions import DeclareLaunchArgument, IncludeLaunchDescription
from launch.launch_description_sources import PythonLaunchDescriptionSource
from launch.substitutions import LaunchConfiguration

TB3_SIM_LAUNCH = 'tb3_simulation_launch.py'   # verify this exists for your Nav2 version


def generate_launch_description():
    warehouse = os.path.join(
        get_package_share_directory('warefleet_bringup'), 'worlds', 'warehouse.sdf')
    nav2_launch = os.path.join(
        get_package_share_directory('nav2_bringup'), 'launch', TB3_SIM_LAUNCH)

    return LaunchDescription([
        DeclareLaunchArgument('slam', default_value='True',
                              description='build the map live (no pre-made map needed)'),
        DeclareLaunchArgument('headless', default_value='False',
                              description='False = show the Gazebo GUI'),
        IncludeLaunchDescription(
            PythonLaunchDescriptionSource(nav2_launch),
            launch_arguments={
                'world': warehouse,                       # <-- OUR warehouse
                'slam': LaunchConfiguration('slam'),
                'headless': LaunchConfiguration('headless'),
            }.items(),
        ),
    ])
