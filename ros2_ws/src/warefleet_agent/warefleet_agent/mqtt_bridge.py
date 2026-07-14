"""MQTT <-> ROS 2 bridge: plugs the Go fleet manager into the robot side.

Topic map (MQTT contract defined by fleet_manager/internal/mqtt/client.go,
JSON field names by fleet_manager/internal/model/types.go):

    MQTT  warefleet/orders                     --cache--> (task lookup table)
    MQTT  warefleet/robots/<id>/assignment     ---------> ROS warefleet/tasks (Task)
    ROS   warefleet/robot_state (RobotState)   ---------> MQTT warefleet/robots/<id>/state

An Assignment only carries task_id + robot_id, so the bridge subscribes to the
order stream too and joins assignment->task before handing the robot a
self-contained Task (the agent never needs fleet-manager state).

Backend analogy: this is the protocol adapter between the two buses — DDS on
the robot side, MQTT on the fleet side — like a Kafka Connect connector.
Heartbeats go out QoS 0 (lossy is fine at 1 Hz), orders/assignments QoS 1.

Multi-robot: the `robots` parameter lists the fleet's robot names. Convention:
a robot's ROS namespace equals its robot_id — except the legacy single-robot
name 'robot_0', which runs unnamespaced (Phase 1/2 compatibility). Assignments
are routed to /<robot_id>/warefleet/tasks by the robot_id in the message.
"""
import json
import threading

import paho.mqtt.client as mqtt
import rclpy
from rclpy.node import Node

from warefleet_msgs.msg import RobotState, Task

ORDERS_TOPIC = 'warefleet/orders'
ASSIGNMENT_TOPIC = 'warefleet/robots/+/assignment'
STATE_TOPIC = 'warefleet/robots/{robot_id}/state'


def _ns_prefix(robot_id: str) -> str:
    return '' if robot_id == 'robot_0' else f'/{robot_id}/'


class MqttBridge(Node):
    def __init__(self):
        super().__init__('warefleet_mqtt_bridge')
        self.declare_parameter('broker_host', 'localhost')
        self.declare_parameter('broker_port', 1883)
        self.declare_parameter('robots', ['robot_0'])
        host = self.get_parameter('broker_host').value
        port = self.get_parameter('broker_port').value
        robots = self.get_parameter('robots').value

        self._tasks = {}  # task_id -> order JSON dict, filled from ORDERS_TOPIC
        self._lock = threading.Lock()

        # ROS side: one task publisher + one state subscription per robot
        self.task_pubs = {}
        for r in robots:
            ns = _ns_prefix(r)
            self.task_pubs[r] = self.create_publisher(Task, f'{ns}warefleet/tasks', 10)
            self.create_subscription(
                RobotState, f'{ns}warefleet/robot_state', self._on_robot_state, 10)
        self.get_logger().info(f'bridging robots: {robots}')

        # MQTT side (paho runs its own network thread via loop_start)
        self.mq = mqtt.Client(client_id='warefleet-ros-bridge')
        self.mq.on_connect = self._on_mqtt_connect
        self.mq.on_message = self._on_mqtt_message
        self.mq.connect(host, port, keepalive=30)
        self.mq.loop_start()
        self.get_logger().info(f'MQTT bridge up, broker {host}:{port}')

    # ---------- MQTT -> ROS ----------

    def _on_mqtt_connect(self, client, userdata, flags, rc):
        client.subscribe([(ORDERS_TOPIC, 1), (ASSIGNMENT_TOPIC, 1)])
        self.get_logger().info(f'connected (rc={rc}), subscribed to orders + assignments')

    def _on_mqtt_message(self, client, userdata, msg):
        try:
            payload = json.loads(msg.payload)
        except json.JSONDecodeError as e:
            self.get_logger().error(f'bad JSON on {msg.topic}: {e}')
            return
        if msg.topic == ORDERS_TOPIC:
            with self._lock:
                self._tasks[payload['task_id']] = payload
            self.get_logger().info(f'order cached: {payload["task_id"]}')
        else:  # warefleet/robots/<id>/assignment
            self._dispatch_assignment(payload)

    def _dispatch_assignment(self, assignment):
        task_id = assignment.get('task_id', '')
        robot_id = assignment.get('robot_id', '')
        pub = self.task_pubs.get(robot_id)
        if pub is None:
            self.get_logger().error(
                f'assignment {task_id} for unknown robot {robot_id!r} '
                f'(bridge knows: {list(self.task_pubs)})')
            return
        with self._lock:
            order = self._tasks.get(task_id)
        if order is None:
            self.get_logger().error(f'assignment for unknown task {task_id!r} (no order seen)')
            return
        t = Task()
        t.task_id = task_id
        t.kind = order.get('kind', 'transport')
        t.pickup.x = float(order['pickup']['x'])
        t.pickup.y = float(order['pickup']['y'])
        t.dropoff.x = float(order['dropoff']['x'])
        t.dropoff.y = float(order['dropoff']['y'])
        t.priority = int(order.get('priority', 0))
        t.created_at = self.get_clock().now().to_msg()
        pub.publish(t)
        self.get_logger().info(f'assignment {task_id} -> robot {robot_id}: dispatched to ROS')

    # ---------- ROS -> MQTT ----------

    def _on_robot_state(self, s: RobotState):
        # field names match fleet_manager/internal/model/types.go (model.Robot)
        payload = json.dumps({
            'robot_id': s.robot_id,
            'pose': {'x': s.pose.position.x, 'y': s.pose.position.y},
            'status': s.status,
            'current_task_id': s.current_task_id,
            'battery': round(float(s.battery), 3),
            'stamp': s.stamp.sec + s.stamp.nanosec * 1e-9,
        })
        self.mq.publish(STATE_TOPIC.format(robot_id=s.robot_id), payload, qos=0)

    def destroy_node(self):
        self.mq.loop_stop()
        self.mq.disconnect()
        super().destroy_node()


def main(args=None):
    rclpy.init(args=args)
    node = MqttBridge()
    try:
        rclpy.spin(node)
    except KeyboardInterrupt:
        pass
    finally:
        node.destroy_node()
        rclpy.shutdown()


if __name__ == '__main__':
    main()
