"""Order feeder CLI.

Streams a scenario of warehouse orders into the fleet by publishing Task JSON
to MQTT `warefleet/orders` (the Go fleet manager's intake). Explicit orders in
the scenario play back at their timestamps; after the last one, the feeder
keeps generating orders from the scenario's RNG seed until `duration_s` —
that continuous stream is what makes operation *lifelong* (the study setting).

Reproducible by construction: same scenario + seed => identical order stream.

Usage:
    ros2 run warefleet_agent order_feeder --scenario benchmarks/scenarios/peak_hour.yaml
    # useful flags: --rate 2.0 (2x speed)  --limit 5 (stop after 5 orders)

Plain MQTT publisher on purpose — no rclpy: the feeder is fleet-side load
generation (think k6/JMeter), not a robot component.
"""
import argparse
import json
import random
import sys
import time

import paho.mqtt.client as mqtt
import yaml

ORDERS_TOPIC = 'warefleet/orders'


def load_scenario(path):
    with open(path) as f:
        sc = yaml.safe_load(f)
    stations = {s['id']: s for kind in sc['stations'].values() for s in kind}
    return sc, stations


def explicit_orders(sc, stations):
    for i, o in enumerate(sc.get('orders', [])):
        yield o['t'], {
            'task_id': f"{sc['name']}-{i:03d}",
            'kind': o.get('kind', 'transport'),
            'pickup': {'x': stations[o['pickup']]['x'], 'y': stations[o['pickup']]['y']},
            'dropoff': {'x': stations[o['dropoff']]['x'], 'y': stations[o['dropoff']]['y']},
            'priority': o.get('priority', 1),
        }


def generated_orders(sc, stations, start_t, start_seq):
    """Continue the stream from the seed: exponential inter-arrival times,
    uniform random pick->drop station pairs."""
    rng = random.Random(sc['seed'])
    picks = sc['stations']['pick']
    drops = sc['stations']['drop']
    mean = sc.get('mean_interval_s', 40.0)
    t, seq = start_t, start_seq
    while t < sc['duration_s']:
        t += rng.expovariate(1.0 / mean)
        p, d = rng.choice(picks), rng.choice(drops)
        yield t, {
            'task_id': f"{sc['name']}-{seq:03d}",
            'kind': 'transport',
            'pickup': {'x': p['x'], 'y': p['y']},
            'dropoff': {'x': d['x'], 'y': d['y']},
            'priority': rng.randint(1, 3),
        }
        seq += 1


def main(argv=None):
    parser = argparse.ArgumentParser(description='WareFleet order feeder')
    parser.add_argument('--scenario', required=True, help='YAML scenario file')
    parser.add_argument('--rate', type=float, default=1.0, help='playback speed multiplier')
    parser.add_argument('--limit', type=int, default=0, help='stop after N orders (0 = full scenario)')
    parser.add_argument('--broker', default='localhost', help='MQTT broker host')
    parser.add_argument('--port', type=int, default=1883)
    parser.add_argument('--dump', default=None, metavar='FILE',
                        help='write the schedule as JSON and exit (no MQTT). '
                             'This is the shared input for the idealized tier '
                             '(gridsim) — one schedule, two simulators.')
    args = parser.parse_args(argv if argv is not None else sys.argv[1:])

    sc, stations = load_scenario(args.scenario)
    schedule = list(explicit_orders(sc, stations))
    last_t = schedule[-1][0] if schedule else 0.0
    schedule += list(generated_orders(sc, stations, last_t, len(schedule)))
    if args.limit:
        schedule = schedule[:args.limit]

    if args.dump:
        with open(args.dump, 'w') as f:
            json.dump([dict(order, t=t) for t, order in schedule], f, indent=1)
        print(f'[order_feeder] schedule ({len(schedule)} orders) -> {args.dump}')
        return

    mq = mqtt.Client(client_id='warefleet-order-feeder')
    mq.connect(args.broker, args.port, keepalive=30)
    mq.loop_start()

    print(f"[order_feeder] scenario={sc['name']} orders={len(schedule)} "
          f"duration={sc['duration_s']}s rate={args.rate}x broker={args.broker}:{args.port}")
    t0 = time.monotonic()
    try:
        for t, order in schedule:
            delay = t / args.rate - (time.monotonic() - t0)
            if delay > 0:
                time.sleep(delay)
            mq.publish(ORDERS_TOPIC, json.dumps(order), qos=1)
            print(f"[{t:7.1f}s] {order['task_id']}: {order['kind']} "
                  f"({order['pickup']['x']:.1f},{order['pickup']['y']:.1f}) -> "
                  f"({order['dropoff']['x']:.1f},{order['dropoff']['y']:.1f}) prio={order['priority']}")
    except KeyboardInterrupt:
        print('\n[order_feeder] interrupted')
    finally:
        mq.loop_stop()
        mq.disconnect()
    print('[order_feeder] scenario complete')


if __name__ == '__main__':
    main()
