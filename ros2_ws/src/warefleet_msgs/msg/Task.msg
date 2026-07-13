# A warehouse task (e.g. move a tote from pick station to drop station).
string task_id
string kind            # "pick" | "transport" | "replenish"
geometry_msgs/Point pickup
geometry_msgs/Point dropoff
int32 priority         # higher = more urgent
builtin_interfaces/Time created_at
