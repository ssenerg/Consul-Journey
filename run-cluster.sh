#!/usr/bin/env bash
# Launch N instances of consul-journey locally, each on its own HTTP port.
# They register with the Consul agent, discover each other, and elect a leader.
#
#   ./run-cluster.sh            # 3 instances on ports 8080..8082
#   ./run-cluster.sh 5          # 5 instances
#   COUNT=4 BASE_PORT=9000 ./run-cluster.sh
#
# Ctrl-C stops all instances (each deregisters + releases leadership cleanly).
set -euo pipefail

COUNT="${1:-${COUNT:-3}}"
BASE_PORT="${BASE_PORT:-8080}"
CONSUL_ADDR="${CJ_CONSUL_ADDR:-127.0.0.1:8500}"

# Consul runs in Docker (see deployment/docker/docker-compose.yml) and scrapes
# each instance's /health from inside its container, so instances must bind all
# interfaces and advertise a host address the container can reach.
#   - Docker Consul (default): advertise host.docker.internal, bind 0.0.0.0
#   - Native Consul on host  : CJ_ADVERTISE_ADDR=127.0.0.1 ./run-cluster.sh
ADVERTISE_ADDR="${CJ_ADVERTISE_ADDR:-host.docker.internal}"
BIND_ADDR="${CJ_BIND_ADDR:-0.0.0.0}"

echo "building..."
go build -o ./.bin/consul-journey .

pids=()
cleanup() {
  echo
  echo "stopping ${#pids[@]} instance(s)..."
  for pid in "${pids[@]}"; do
    kill -TERM "$pid" 2>/dev/null || true
  done
  # Wait so each instance can deregister + release its lock before we exit.
  for pid in "${pids[@]}"; do
    wait "$pid" 2>/dev/null || true
  done
  echo "all stopped."
}
trap cleanup INT TERM

echo "starting $COUNT instances against Consul at $CONSUL_ADDR"
echo "  (bind=$BIND_ADDR advertise=$ADVERTISE_ADDR)"
for i in $(seq 0 $((COUNT - 1))); do
  port=$((BASE_PORT + i))
  CJ_CONSUL_ADDR="$CONSUL_ADDR" \
  CJ_BIND_ADDR="$BIND_ADDR" \
  CJ_ADVERTISE_ADDR="$ADVERTISE_ADDR" \
    ./.bin/consul-journey -port "$port" -id "node-$port" 2>&1 \
    | sed "s/^/[$port] /" &
  pids+=($!)
  echo "  -> instance node-$port on http://127.0.0.1:$port  (pid $!)"
  sleep 0.3
done

echo
echo "dashboards: $(for i in $(seq 0 $((COUNT - 1))); do printf 'http://127.0.0.1:%d  ' $((BASE_PORT + i)); done)"
echo "consul ui : http://127.0.0.1:8500"
echo "press Ctrl-C to stop."
wait
