# Conflict-free path for one robot, produced by the MAPF layer.
string robot_id
string task_id
geometry_msgs/Point[] waypoints
# Optional time-stamped schedule for space-time deconfliction (MAPF).
builtin_interfaces/Time[] arrival_times
