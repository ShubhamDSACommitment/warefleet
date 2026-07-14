# Fleet Manager (Go)

The distributed coordination service: task intake → allocation → MAPF → dispatch,
with dropout detection + re-allocation and a centralized/decentralized mode switch.

## Layout

```
cmd/fleet-manager/     entrypoint; wiring + config from env
internal/model/        domain types (Task, Robot, Assignment, PathPlan)
internal/allocation/   pluggable allocation (greedy | hungarian | auction)
internal/coordination/ control loop + pluggable MAPF (prioritized | cbs | pibt)
internal/mqtt/         robot I/O over MQTT
internal/metrics/      Prometheus KPIs
```

## Config (env)

| Var | Default | Meaning |
|---|---|---|
| `WAREFLEET_MQTT_BROKER` | `tcp://localhost:1883` | broker URL |
| `WAREFLEET_ALLOCATION` | `greedy` | greedy \| hungarian \| auction |
| `WAREFLEET_COORDINATION` | `prioritized` | prioritized \| cbs \| pibt \| lacam |
| `WAREFLEET_MODE` | `centralized` | centralized \| decentralized |
| `WAREFLEET_METRICS_ADDR` | `:2112` | Prometheus scrape addr |

## Develop

```bash
go mod tidy      # populate go.sum
go build ./...
go test ./...
go run ./cmd/fleet-manager
```

## Implementation order

Greedy allocation (wk4) → Hungarian + Auction (wk5) → prioritized MAPF (wk5, port
from the Python reference) → dropout handling (wk6) → CBS/PIBT + decentralized mode (stretch).
