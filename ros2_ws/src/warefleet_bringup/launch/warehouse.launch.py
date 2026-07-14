"""WareFleet Phase 3 — N robots in our warehouse (AMCL on the saved map).

Self-contained multi-robot bringup: Gazebo server with our world, exactly ONE
fleet-wide /clock bridge, and per robot a spawn + sensor bridge (clock-free
config) + robot_state_publisher + Nav2 stack (AMCL against maps/warehouse.yaml).

Why not nav2's cloned_multi_tb3_simulation_launch.py? Its per-robot spawn
bridges /clock once PER ROBOT; N publishers on /clock interleave out of order,
ROS time appears to jump backwards, TF buffers get wiped, and navigation dies
with TF_ERRORs. Owning the spawn lets us bridge the clock exactly once.

Run (the robots list MUST be on the CLI — nav2_common's ParseMultiRobotPose
reads argv, so a default declared here would not be seen):

  ros2 launch warefleet_bringup warehouse.launch.py \
    robots:='robot1={x: -2.0, y: -0.5, yaw: 0.0}; robot2={x: -2.0, y: -2.5, yaw: 0.0}'

Notes:
  * Robot poses are in the WORLD (Gazebo) frame. The map frame is offset:
    map = world + (2.0, 0.5) (the SLAM run that built the map started at
    world (-2.0, -0.5), which became map (0, 0)).
  * AMCL needs an initial pose per robot on /<ns>/initialpose during startup —
    scripts/demo.sh publishes them (repeatedly, until Nav2 activates).
  * Server-only Gazebo; run `gz sim -g` alongside for a GUI (demo.sh does).
"""
import os
import tempfile

from ament_index_python.packages import get_package_share_directory
from launch import LaunchDescription
from launch.actions import (
    AppendEnvironmentVariable,
    DeclareLaunchArgument,
    ExecuteProcess,
    GroupAction,
    IncludeLaunchDescription,
    LogInfo,
    OpaqueFunction,
    RegisterEventHandler,
)
from launch.event_handlers import OnShutdown
from launch.launch_description_sources import PythonLaunchDescriptionSource
from launch.substitutions import Command, FindExecutable, LaunchConfiguration, TextSubstitution
from launch_ros.actions import Node
from nav2_common.launch import ParseMultiRobotPose


def generate_launch_description():
    pkg = get_package_share_directory('warefleet_bringup')
    nav2 = get_package_share_directory('nav2_bringup')
    sim_dir = get_package_share_directory('nav2_minimal_tb3_sim')

    map_yaml = LaunchConfiguration('map')
    params_file = LaunchConfiguration('params_file')

    declare_world = DeclareLaunchArgument(
        'world', default_value=os.path.join(pkg, 'worlds', 'warehouse.sdf'))
    declare_map = DeclareLaunchArgument(
        'map', default_value=os.path.join(pkg, 'maps', 'warehouse.yaml'))
    declare_params = DeclareLaunchArgument(
        'params_file',
        default_value=os.path.join(nav2, 'params', 'nav2_multirobot_params_all.yaml'))

    # --- Gazebo server with our world (xacro pass-through, like nav2 does) ---
    world_sdf = tempfile.mktemp(prefix='warefleet_', suffix='.sdf')
    world_xacro = ExecuteProcess(
        cmd=['xacro', '-o', world_sdf, ['headless:=', 'False'],
             LaunchConfiguration('world')])
    gz_server = ExecuteProcess(cmd=['gz', 'sim', '-r', '-s', world_sdf], output='screen')
    cleanup_world = RegisterEventHandler(event_handler=OnShutdown(
        on_shutdown=[OpaqueFunction(function=lambda _: os.remove(world_sdf))]))

    # --- exactly ONE clock bridge for the whole fleet ---
    clock_bridge = Node(
        package='ros_gz_bridge', executable='parameter_bridge', name='clock_bridge',
        arguments=['/clock@rosgraph_msgs/msg/Clock[gz.msgs.Clock'],
        output='screen')

    urdf = os.path.join(sim_dir, 'urdf', 'turtlebot3_waffle.urdf')
    with open(urdf) as f:
        robot_description = f.read()
    remappings = [('/tf', 'tf'), ('/tf_static', 'tf_static')]

    robots = ParseMultiRobotPose('robots').value()

    groups = []
    for name, pose in robots.items():
        groups.append(GroupAction([
            LogInfo(msg=['warefleet: bringing up ', name, ' at ', str(pose)]),
            # sensor/cmd bridge — namespaced topics, NO clock entry
            Node(
                package='ros_gz_bridge', executable='parameter_bridge',
                namespace=name,
                parameters=[{
                    'config_file': os.path.join(pkg, 'config', 'tb3_bridge_noclock.yaml'),
                    'expand_gz_topic_names': True,
                    'use_sim_time': True,
                }],
                output='screen'),
            # spawn the robot model
            Node(
                package='ros_gz_sim', executable='create', namespace=name,
                output='screen',
                arguments=[
                    '-name', name,
                    '-string', Command([
                        FindExecutable(name='xacro'), ' ', 'namespace:=',
                        TextSubstitution(text=name), ' ',
                        os.path.join(sim_dir, 'urdf', 'gz_waffle.sdf.xacro')]),
                    '-x', str(pose['x']), '-y', str(pose['y']), '-z', '0.01',
                    '-R', str(pose['roll']), '-P', str(pose['pitch']),
                    '-Y', str(pose['yaw'])]),
            # TF from the robot model
            Node(
                package='robot_state_publisher', executable='robot_state_publisher',
                name='robot_state_publisher', namespace=name, output='screen',
                parameters=[{'use_sim_time': True,
                             'robot_description': robot_description}],
                remappings=remappings),
            # Nav2: AMCL localization + navigation servers
            IncludeLaunchDescription(
                PythonLaunchDescriptionSource(
                    os.path.join(nav2, 'launch', 'bringup_launch.py')),
                launch_arguments={
                    'namespace': name,
                    'use_namespace': 'True',
                    'slam': 'False',
                    'map': map_yaml,
                    'use_sim_time': 'True',
                    'params_file': params_file,
                    'autostart': 'True',
                }.items()),
        ]))

    ld = LaunchDescription()
    ld.add_action(AppendEnvironmentVariable(
        'GZ_SIM_RESOURCE_PATH', os.path.join(sim_dir, 'models')))
    ld.add_action(declare_world)
    ld.add_action(declare_map)
    ld.add_action(declare_params)
    ld.add_action(world_xacro)
    ld.add_action(gz_server)
    ld.add_action(cleanup_world)
    ld.add_action(clock_bridge)
    ld.add_action(LogInfo(msg=['warefleet: fleet size = ', str(len(robots))]))
    for g in groups:
        ld.add_action(g)
    return ld
