"""Order feeder CLI.

Streams a scenario file of warehouse orders into the fleet (publishes Task
messages / MQTT `warefleet/orders`) at the timestamps defined in the scenario.
Used to drive reproducible benchmarks.

Usage:
    ros2 run warefleet_agent order_feeder --scenario benchmarks/scenarios/peak_hour.yaml
"""
import argparse
import sys

# import yaml
# import rclpy
# from warefleet_msgs.msg import Task


def main(argv=None):
    parser = argparse.ArgumentParser(description='WareFleet order feeder')
    parser.add_argument('--scenario', required=True, help='YAML scenario file')
    parser.add_argument('--rate', type=float, default=1.0,
                        help='playback speed multiplier')
    args = parser.parse_args(argv if argv is not None else sys.argv[1:])

    print(f'[order_feeder] scenario={args.scenario} rate={args.rate}x')
    # TODO(week4): load YAML scenario, init rclpy, publish Task messages on schedule
    # TODO(week4): also publish to MQTT topic warefleet/orders for the Go fleet manager


if __name__ == '__main__':
    main()
