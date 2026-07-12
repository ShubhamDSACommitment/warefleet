"""Bring up the warehouse simulation with N robots.

Launches:
  * Gazebo with the warehouse world
  * N namespaced robots (robot_0 .. robot_{N-1}), each with a Nav2 stack
  * one warefleet_agent per robot

Usage:
  ros2 launch warefleet_bringup warehouse.launch.py robots:=4 world:=sim/worlds/warehouse.world

Scaffold — fill in incrementally per docs/roadmap.md:
  Week 1-2 -> single robot + Gazebo + Nav2
  Week 3   -> loop to N namespaced robots + agents
"""
from launch import LaunchDescription
from launch.actions import DeclareLaunchArgument, OpaqueFunction
from launch.substitutions import LaunchConfiguration
from launch_ros.actions import Node


def spawn_fleet(context, *args, **kwargs):
    n = int(LaunchConfiguration('robots').perform(context))
    nodes = []
    for i in range(n):
        robot_id = f'robot_{i}'
        nodes.append(Node(
            package='warefleet_agent',
            executable='agent_node',
            namespace=robot_id,
            name='agent',
            parameters=[{'robot_id': robot_id}],
            output='screen',
        ))
        # TODO(week2-3): add per-robot Nav2 bringup + Gazebo spawn_entity here,
        #                each remapped into the robot's namespace.
    return nodes


def generate_launch_description():
    return LaunchDescription([
        DeclareLaunchArgument('robots', default_value='4',
                              description='number of robots to spawn'),
        DeclareLaunchArgument('world', default_value='sim/worlds/warehouse.world',
                              description='Gazebo world file'),
        # TODO(week1): launch Gazebo (ros_gz_sim) with the warehouse world
        OpaqueFunction(function=spawn_fleet),
    ])
