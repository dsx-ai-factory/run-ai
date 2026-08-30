#!/usr/bin/env bash
#
# Attaches local Grafana to a running cluster: port-forwards Prometheus and starts the
# Grafana container, then cleans both up on exit.
#
# NOTE: a live auto-refreshing dashboard contaminates the container CPU/memory panels —
# every refresh fires all panels as concurrent PromQL queries, and that query load shows
# up in the very numbers you are trying to read. Use this for health checks and for
# watching an env you are debugging; close it before a measurement run.
#
# Usage: ./monitor-live.sh [kubeconfig]
#        KUBECONFIG=~/.kube/my-env ./monitor-live.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GRAFANA_DIR="${SCRIPT_DIR}/../grafana"

if [[ $# -ge 1 ]]; then
  export KUBECONFIG="$1"
fi

NAMESPACE="${PROMETHEUS_NAMESPACE:-runai}"
# 9091 is what grafana/provisioning/datasources/prometheus.yml points at via
# host.docker.internal. Changing it here means changing that file too.
LOCAL_PORT="${LOCAL_PORT:-9091}"

log() { echo "[monitor-live] $*"; }

pod="$(kubectl get pods -n "${NAMESPACE}" -l app.kubernetes.io/name=prometheus \
  -o jsonpath='{.items[0].metadata.name}')"
[[ -n "${pod}" ]] || { echo "[monitor-live] ERROR: no Prometheus pod in ${NAMESPACE}" >&2; exit 1; }

pf_pid=""
cleanup() {
  log "shutting down"
  [[ -n "${pf_pid}" ]] && kill "${pf_pid}" 2>/dev/null || true
  docker compose -f "${GRAFANA_DIR}/docker-compose.yml" down 2>/dev/null || true
}
trap cleanup EXIT INT TERM

log "port-forwarding ${NAMESPACE}/${pod} to localhost:${LOCAL_PORT}"
kubectl port-forward -n "${NAMESPACE}" "pod/${pod}" "${LOCAL_PORT}:9090" >/dev/null 2>&1 &
pf_pid=$!

for _ in $(seq 1 30); do
  curl -sf "http://localhost:${LOCAL_PORT}/-/ready" >/dev/null 2>&1 && break
  sleep 2
done
curl -sf "http://localhost:${LOCAL_PORT}/-/ready" >/dev/null 2>&1 \
  || { echo "[monitor-live] ERROR: Prometheus did not become reachable on ${LOCAL_PORT}" >&2; exit 1; }

log "starting Grafana"
docker compose -f "${GRAFANA_DIR}/docker-compose.yml" up -d

cat <<EOF

  Validity gates : http://localhost:3001/d/validity-gate-monitor
  Workload activity: http://localhost:3001/d/workload-activity

  No login required. Ctrl-C to stop the port-forward and Grafana.

EOF

# Hold the port-forward open until interrupted; the trap tears everything down.
wait "${pf_pid}"
