# Periodic state heartbeat published by each robot agent.
string robot_id
geometry_msgs/Pose pose
string status          # "idle" | "busy" | "charging" | "error" | "offline"
string current_task_id # empty if idle
float32 battery        # 0.0 - 1.0
builtin_interfaces/Time stamp
