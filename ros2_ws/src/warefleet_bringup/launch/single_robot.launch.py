"""WareFleet Phase 1 — one robot in our warehouse, navigating.

Strategy (see docs/focus-and-framing.md): DON'T reinvent the spawn/bridge/Nav2
wiring — reuse Nav2's proven TB3 simulation launch, but point it at OUR
warehouse world and run SLAM live (so no pre-built map is needed yet).

Run:
  export TURTLEBOT3_MODEL=waffle
  ros2 launch warefleet_bringup single_robot.launch.py

Then in RViz: "2D Pose Estimate", then "Nav2 Goal" -> watch it plan + drive.

NOTE: nav2_bringup's launch filename/args vary by version. Verify with:
  ros2 launch nav2_bringup tb3_simulation_launch.py --show-args
  ls $(ros2 pkg prefix nav2_bringup)/share/nav2_bringup/launch/
and adjust `TB3_SIM_LAUNCH` / arg names below if needed. We'll iterate on errors.
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
