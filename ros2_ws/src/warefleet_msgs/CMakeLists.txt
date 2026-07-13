cmake_minimum_required(VERSION 3.8)
project(warefleet_msgs)

find_package(ament_cmake REQUIRED)
find_package(rosidl_default_generators REQUIRED)
find_package(geometry_msgs REQUIRED)
find_package(builtin_interfaces REQUIRED)

rosidl_generate_interfaces(${PROJECT_NAME}
  "msg/Task.msg"
  "msg/RobotState.msg"
  "msg/TaskAssignment.msg"
  "msg/PathPlan.msg"
  DEPENDENCIES geometry_msgs builtin_interfaces
)

ament_package()
