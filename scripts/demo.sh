#!/usr/bin/env bash
# WareFleet one-command demo orchestrator.
#
#   scripts/demo.sh                     single robot, SLAM, GUI (Phase 1/2 mode)
#   ROBOTS=2 scripts/demo.sh            multi-robot fleet, AMCL on saved map (Phase 3)
#   scripts/demo.sh --headless          either mode without GUI
#   scripts/demo.sh agent-only <name>   (re)start one robot agent — `make revive-robot`
#
# Startup order, each step gated on the previous one being actually ready:
#   1. sweep zombie sim processes (a stale ros_gz_bridge double-publishes
#      /clock and corrupts TF for every later run — hard-won lesson)
#   2. MQTT broker (docker)
#   3. Gazebo + Nav2 (+SLAM single / +AMCL multi) -> wait for lifecycle active
#   4. AMCL initial poses (multi only)
#   5. robot agents + MQTT bridge
#   6. Go fleet manager
# Logs: .logs/<component>.log   PIDs: .logs/pids   Stop: scripts/stop.sh
# (no `set -u`: ROS setup.bash trips over unbound variables)
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOGS="$ROOT/.logs"
mkdir -p "$LOGS"

ROBOTS="${ROBOTS:-1}"
# Spawn poses in the WORLD (gazebo) frame, all open floor, mutually clear.
WORLD_POSES=("-2.0 -0.5" "-2.0 -2.5" "2.0 2.5" "4.5 -0.5")
# map frame = world + (2.0, 0.5) — offset of the SLAM run that built the map
MAP_DX=2.0; MAP_DY=0.5

source /opt/ros/jazzy/setup.bash
if [ ! -f "$ROOT/ros2_ws/install/setup.bash" ]; then
    echo "[demo] workspace not built yet — running colcon build once..."
    (cd "$ROOT/ros2_ws" && colcon build) || exit 1
fi
source "$ROOT/ros2_ws/install/setup.bash"

start() { # start <name> <cmd...>: background a component, record pid, log to .logs/<name>.log
    local name=$1; shift
    "$@" >"$LOGS/$name.log" 2>&1 &
    echo "$! $name" >>"$LOGS/pids"
    echo "[demo] $name up (pid $!, log .logs/$name.log)"
}

start_agent() { # start_agent <robot_name>  (robot_0 = legacy unnamespaced single robot)
    local name=$1 ns_args=()
    [ "$name" != "robot_0" ] && ns_args=(-r __ns:=/"$name")
    pkill -f "[r]obot_id:=$name" 2>/dev/null && sleep 1  # idempotent: never two agents per robot
    start "agent-$name" ros2 run warefleet_agent agent_node --ros-args \
        "${ns_args[@]}" -p robot_id:="$name" -p use_sim_time:=true
}

if [ "${1:-}" = "agent-only" ]; then
    start_agent "${2:-robot_0}"
    exit 0
fi

HEADLESS=False
[ "${1:-}" = "--headless" ] && HEADLESS=True

# --- 1. zombie sweep -------------------------------------------------------
for pat in "gz sim" parameter_bridge component_container_isolated \
           sync_slam_toolbox_node rviz2 "warefleet_agent" "fleet-manager"; do
    if pkill -f "$pat" 2>/dev/null; then echo "[demo] killed stale: $pat"; fi
done
sleep 1
rm -f "$LOGS/pids"

# --- 2. broker --------------------------------------------------------------
(cd "$ROOT" && docker compose up -d mqtt) || exit 1

# --- 3. simulation (retry: nav2 lifecycle activation flakes occasionally) ---
if [ "$ROBOTS" -gt 1 ]; then
    robots_arg=""; names=()
    for i in $(seq 1 "$ROBOTS"); do
        set -- ${WORLD_POSES[$((i-1))]}
        robots_arg+="robot$i={x: $1, y: $2, yaw: 0.0}; "
        names+=("robot$i")
    done
    SIM_CMD=(ros2 launch warefleet_bringup warehouse.launch.py robots:="$robots_arg")
    NEED_ACTIVE=$ROBOTS
else
    names=(robot_0)
    SIM_CMD=(ros2 launch warefleet_bringup single_robot.launch.py \
        headless:="$HEADLESS" use_rviz:="$([ "$HEADLESS" = True ] && echo False || echo True)")
    NEED_ACTIVE=1
fi

publish_initial_poses() { # AMCL needs a pose before costmaps can finish activating
    for i in $(seq 1 "$ROBOTS"); do
        set -- ${WORLD_POSES[$((i-1))]}
        local mx my
        mx=$(python3 -c "print($1 + $MAP_DX)"); my=$(python3 -c "print($2 + $MAP_DY)")
        timeout 10 ros2 topic pub --once "/robot$i/initialpose" geometry_msgs/msg/PoseWithCovarianceStamped \
            "{header: {frame_id: map}, pose: {pose: {position: {x: $mx, y: $my}, orientation: {w: 1.0}},
              covariance: [0.25,0,0,0,0,0, 0,0.25,0,0,0,0, 0,0,0,0,0,0, 0,0,0,0,0,0, 0,0,0,0,0,0, 0,0,0,0,0,0.068]}}" \
            >/dev/null 2>&1
    done
}

sim_ready=0
for attempt in 1 2 3; do
    start sim "${SIM_CMD[@]}"
    echo -n "[demo] waiting for $NEED_ACTIVE Nav2 stack(s) to activate (attempt $attempt)"
    for i in $(seq 1 90); do
        n=$(grep -c "lifecycle_manager_navigation.*Managed nodes are active" "$LOGS/sim.log" 2>/dev/null)
        if [ "${n:-0}" -ge "$NEED_ACTIVE" ]; then echo " — ACTIVE ($n/$NEED_ACTIVE)"; sim_ready=1; break; fi
        if grep -q "Aborting bringup" "$LOGS/sim.log" 2>/dev/null; then echo " — flaked, retrying"; break; fi
        # multi-robot: keep feeding AMCL its initial pose while lifecycle waits on TF
        if [ "$ROBOTS" -gt 1 ] && [ $((i % 3)) = 0 ] && \
           grep -q "Please set the initial pose" "$LOGS/sim.log" 2>/dev/null; then
            publish_initial_poses
        fi
        [ "$i" = 90 ] && echo " — timeout, retrying"
        echo -n "."; sleep 2
    done
    [ "$sim_ready" = 1 ] && break
    pkill -f "warefleet_bringup" 2>/dev/null; sleep 2
    for pat in "gz sim" parameter_bridge component_container_isolated sync_slam_toolbox_node rviz2; do
        pkill -9 -f "$pat" 2>/dev/null
    done
    sleep 3
    grep -v " sim$" "$LOGS/pids" > "$LOGS/pids.tmp" 2>/dev/null; mv -f "$LOGS/pids.tmp" "$LOGS/pids" 2>/dev/null
done
if [ "$sim_ready" != 1 ]; then
    echo "[demo] simulation failed to start after 3 attempts — see .logs/sim.log"
    exit 1
fi

# multi-robot: the nav2 multi launch runs the gz server only — attach a GUI client
if [ "$ROBOTS" -gt 1 ] && [ "$HEADLESS" = False ]; then
    start gz-gui gz sim -g
fi

# --- 4. AMCL initial poses: one final confirmed round (multi only) -----------
if [ "$ROBOTS" -gt 1 ]; then
    echo "[demo] confirming AMCL initial poses"
    publish_initial_poses
fi

# --- 5. robot agents + bridge -------------------------------------------------
for name in "${names[@]}"; do start_agent "$name"; done
robots_yaml=$(printf '%s,' "${names[@]}"); robots_yaml="[${robots_yaml%,}]"
start bridge ros2 run warefleet_agent mqtt_bridge --ros-args -p robots:="$robots_yaml"

# --- 6. fleet manager ----------------------------------------------------------
(cd "$ROOT/fleet_manager" && go build -o bin/fleet-manager ./cmd/fleet-manager) || exit 1
start fleet-manager "$ROOT/fleet_manager/bin/fleet-manager"

echo
echo "🚀 WareFleet is up with ${#names[@]} robot(s): ${names[*]}"
echo "   make orders           stream the scenario (LIMIT=3 for a short run)"
echo "   make kill-robot ROBOT=${names[0]}   H2 experiment: simulate a dropout"
echo "   make revive-robot ROBOT=${names[0]} bring that robot back"
echo "   make metrics          fleet KPIs (completed, makespan, recovery)"
echo "   make logs             follow all component logs"
echo "   make stop             shut everything down"
