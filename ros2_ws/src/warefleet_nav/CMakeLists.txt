cmake_minimum_required(VERSION 3.8)
project(warefleet_nav)

find_package(ament_cmake REQUIRED)

install(DIRECTORY config maps
  DESTINATION share/${PROJECT_NAME}
)

ament_package()
