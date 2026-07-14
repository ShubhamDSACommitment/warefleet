# WareFleet — developer entry points.
# These wrap the common workflows so the README "quick start" is real.

ROBOTS ?= 4
SCENARIO ?= benchmarks/scenarios/peak_hour.yaml
LIMIT ?= 0
WORLD ?= sim/worlds/warehouse.world

# Every ROS-touching target sources the env itself, so plain `make <target>`
# works from any fresh terminal.
ROS_ENV = source /opt/ros/jazzy/setup.bash && source ros2_ws/install/setup.bash

.PHONY: help demo demo-headless stop status logs orders \
        metrics up down sim build-ros build-go fmt bench clean

help:
	@echo "WareFleet — the one-command flow:"
	@echo "  make demo             - start EVERYTHING (broker, sim+GUI, agent, bridge, fleet manager)"
	@echo "  make demo-headless    - same, without Gazebo/RViz windows"
	@echo "  make orders           - stream the order scenario (LIMIT=3 for a short run)"
	@echo "  make metrics          - fleet KPIs (tasks completed, makespan, recovery time)"
	@echo "  make logs             - follow all component logs"
	@echo "  make status           - what's running"
	@echo "  make stop             - shut everything down (incl. zombie sweep)"
	@echo ""
	@echo "Building & extras:"
	@echo "  make build-ros / build-go / fmt / bench / clean"
	@echo "  make up / down        - full observability stack (prometheus + grafana) via compose"

# ---- the one-command flow ---------------------------------------------------

demo:
	@bash scripts/demo.sh

demo-headless:
	@bash scripts/demo.sh --headless

stop:
	@bash scripts/stop.sh

status:
	@ps -eo pid,etime,args | grep -E "[g]z sim -r|[a]gent_node|[m]qtt_bridge|[f]leet-manager|[p]arameter_bridge" || echo "nothing running"

logs:
	@tail -n 5 -f .logs/*.log

orders:
	@bash -c '$(ROS_ENV) && ros2 run warefleet_agent order_feeder --scenario $(SCENARIO) --limit $(LIMIT)'

metrics:
	@curl -s localhost:2112/metrics | grep ^warefleet || echo "fleet manager not running (make demo)"

# ---- observability stack ------------------------------------------------------

up:
	docker compose up -d

down:
	docker compose down

# ---- building -----------------------------------------------------------------

build-go:
	cd fleet_manager && go build -o bin/fleet-manager ./cmd/fleet-manager

build-ros:
	@bash -c 'source /opt/ros/jazzy/setup.bash && cd ros2_ws && colcon build'

# Requires a sourced ROS 2 env + built workspace (multi-robot launch is Phase 3+)
sim:
	@bash -c '$(ROS_ENV) && ros2 launch warefleet_bringup single_robot.launch.py'

bench:
	python3 benchmarks/run_benchmarks.py --scenarios benchmarks/scenarios --out benchmarks/results

fmt:
	cd fleet_manager && gofmt -w .
	@command -v black >/dev/null 2>&1 && black ros2_ws benchmarks || echo "black not installed, skipping python fmt"

clean:
	cd ros2_ws && rm -rf build install log
	cd fleet_manager && go clean && rm -rf bin
