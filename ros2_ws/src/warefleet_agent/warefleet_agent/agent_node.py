"""Per-robot agent node.

Executes warehouse Tasks by driving Nav2, and reports state back to the fleet:

    warefleet/tasks (Task)          --> this node --> Nav2 NavigateToPose action
    warefleet/robot_state (RobotState) <-- 1 Hz heartbeat (pose, status, task)

Task execution is a small state machine:
    IDLE -> TO_PICKUP -> PICKING (dwell) -> TO_DROPOFF -> IDLE

Backend analogy: this is the per-robot sidecar/adapter. It consumes a job from
a queue topic and turns it into an async call against Nav2 (the robot's
"driving service"); the heartbeat is the liveness probe the coordinator uses
for dropout detection.

Scope notes (see docs/roadmap.md):
  * Week 3 -> namespacing for N robots; TaskAssignment filtering via MQTT bridge.
  * Week 5 -> follow a MAPF PathPlan (multi-waypoint) instead of point-to-point.
  * A failed/rejected goal drops the task here; re-queueing is the fleet
    manager's job (coordinator.detectDropouts / re-allocation), not the agent's.
"""
import rclpy
from rclpy.action import ActionClient
from rclpy.node import Node

from geometry_msgs.msg import Point, PoseStamped
from nav_msgs.msg import Odometry
from nav2_msgs.action import NavigateToPose
from warefleet_msgs.msg import RobotState, Task

IDLE = 'idle'
BUSY = 'busy'
ERROR = 'error'


class AgentNode(Node):
    def __init__(self):
        super().__init__('warefleet_agent')
        self.declare_parameter('robot_id', 'robot_0')
        self.declare_parameter('dwell_sec', 2.0)  # simulated pick/drop handling time
        self.robot_id = self.get_parameter('robot_id').value
        self.dwell_sec = self.get_parameter('dwell_sec').value

        self.status = IDLE
        self.task = None          # current Task or None
        self.leg = None           # 'to_pickup' | 'to_dropoff'
        self.pose = None          # last odom pose (map ~= odom: SLAM starts at spawn)
        self._dwell_timer = None
        self._nav_goal_handle = None

        self.nav = ActionClient(self, NavigateToPose, 'navigate_to_pose')
        self.create_subscription(Task, 'warefleet/tasks', self._on_task, 10)
        self.create_subscription(Odometry, 'odom', self._on_odom, 10)
        self.state_pub = self.create_publisher(RobotState, 'warefleet/robot_state', 10)
        self.create_timer(1.0, self._publish_state)

        self.get_logger().info(f'WareFleet agent up: {self.robot_id}')

    # ---------- task intake ----------

    def _on_task(self, task: Task):
        if self.status == BUSY:
            self.get_logger().warning(
                f'busy with {self.task.task_id}, rejecting {task.task_id} '
                '(re-allocation is the fleet manager\'s job)')
            return
        self.task = task
        self.status = BUSY
        self.leg = 'to_pickup'
        self.get_logger().info(
            f'task {task.task_id} ({task.kind}): pickup ({task.pickup.x:.1f}, '
            f'{task.pickup.y:.1f}) -> dropoff ({task.dropoff.x:.1f}, {task.dropoff.y:.1f})')
        self._navigate_to(task.pickup)

    # ---------- Nav2 action client ----------

    def _navigate_to(self, point: Point):
        if not self.nav.wait_for_server(timeout_sec=5.0):
            self.get_logger().error('Nav2 action server not available')
            self._fail_task()
            return
        goal = NavigateToPose.Goal()
        pose = PoseStamped()
        pose.header.frame_id = 'map'
        pose.header.stamp = self.get_clock().now().to_msg()
        pose.pose.position.x = point.x
        pose.pose.position.y = point.y
        pose.pose.orientation.w = 1.0
        goal.pose = pose
        self.get_logger().info(f'navigating ({self.leg}) to ({point.x:.1f}, {point.y:.1f})')
        self.nav.send_goal_async(goal).add_done_callback(self._on_goal_response)

    def _on_goal_response(self, future):
        handle = future.result()
        if not handle.accepted:
            self.get_logger().error('Nav2 rejected the goal')
            self._fail_task()
            return
        self._nav_goal_handle = handle
        handle.get_result_async().add_done_callback(self._on_nav_result)

    def _on_nav_result(self, future):
        result = future.result()
        # status 4 = SUCCEEDED (action_msgs/GoalStatus)
        if result.status != 4 or result.result.error_code != 0:
            self.get_logger().error(
                f'navigation failed (status={result.status}, '
                f'error_code={result.result.error_code})')
            self._fail_task()
            return
        if self.leg == 'to_pickup':
            self.get_logger().info(f'at pickup, handling for {self.dwell_sec}s')
            self._dwell_timer = self.create_timer(self.dwell_sec, self._after_dwell)
        else:
            self.get_logger().info(f'task {self.task.task_id} done')
            self._clear_task()

    def _after_dwell(self):
        self._dwell_timer.cancel()
        self._dwell_timer = None
        self.leg = 'to_dropoff'
        self._navigate_to(self.task.dropoff)

    # ---------- state / heartbeat ----------

    def _fail_task(self):
        self.status = ERROR
        self._publish_state()  # heartbeat shows ERROR before we go back to idle
        self._clear_task()

    def _clear_task(self):
        self.task = None
        self.leg = None
        self._nav_goal_handle = None
        self.status = IDLE

    def _on_odom(self, msg: Odometry):
        self.pose = msg.pose.pose

    def _publish_state(self):
        s = RobotState()
        s.robot_id = self.robot_id
        s.status = self.status
        s.current_task_id = self.task.task_id if self.task else ''
        s.battery = 1.0  # TODO: drain model for the failure experiments
        s.stamp = self.get_clock().now().to_msg()
        if self.pose is not None:
            s.pose = self.pose
        self.state_pub.publish(s)


def main(args=None):
    rclpy.init(args=args)
    node = AgentNode()
    try:
        rclpy.spin(node)
    except KeyboardInterrupt:
        pass
    finally:
        node.destroy_node()
        rclpy.shutdown()


if __name__ == '__main__':
    main()
