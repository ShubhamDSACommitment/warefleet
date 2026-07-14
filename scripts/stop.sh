#!/usr/bin/env bash
# Stop everything scripts/demo.sh started, then sweep for survivors.
set -u
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOGS="$ROOT/.logs"

if [ -f "$LOGS/pids" ]; then
    while read -r pid name; do
        if kill "$pid" 2>/dev/null; then echo "[stop] $name (pid $pid)"; fi
    done <"$LOGS/pids"
    rm -f "$LOGS/pids"
fi
sleep 3

# Sweep: a surviving ros_gz_bridge poisons /clock for every future run.
for pat in "gz sim" parameter_bridge component_container_isolated \
           sync_slam_toolbox_node rviz2 "warefleet_agent" "fleet-manager"; do
    if pkill -9 -f "$pat" 2>/dev/null; then echo "[stop] swept survivor: $pat"; fi
done

(cd "$ROOT" && docker compose stop mqtt)
echo "[stop] done"
