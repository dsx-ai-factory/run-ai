#!/usr/bin/env bash
#
# Prepares a cluster for validity-gate monitoring. Idempotent — safe to re-run.
#
#   1. Applies the etcd metrics proxy and the runai-backend ServiceMonitor.
#   2. Enables the Prometheus admin API (needed for TSDB snapshots) and widens retention.
#   3. Verifies every scrape target the validity gates depend on is actually up.
#
# Usage: ./setup-monitoring.sh [kubeconfig]
#        KUBECONFIG=~/.kube/my-env ./setup-monitoring.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFESTS_DIR="${SCRIPT_DIR}/../manifests"

if [[ $# -ge 1 ]]; then
  export KUBECONFIG="$1"
fi

NAMESPACE="${PROMETHEUS_NAMESPACE:-runai}"
RUNAICONFIG_NAME="${RUNAICONFIG_NAME:-runai}"
# Run:ai ships 2h, which is shorter than a run: by the time a 4h run ends its first
# half has already been dropped, so whole-run queries and any TSDB snapshot see only
# the tail. 10h covers any run we make plus room either side.
PROMETHEUS_RETENTION="${PROMETHEUS_RETENTION:-10h}"
ROLLOUT_TIMEOUT="${ROLLOUT_TIMEOUT:-10m}"
RECONCILE_TIMEOUT_SECS="${RECONCILE_TIMEOUT_SECS:-300}"
TARGET_TIMEOUT_SECS="${TARGET_TIMEOUT_SECS:-300}"

# Must stay in sync with CriticalMetricSources in ../prometheus.go — the jobs feeding
# the validity gate panels. Verifying all of them here, not just the two this script
# installs, means a dead kube-state-metrics (gates 5-7) or node-exporter (gate 12a) is
# caught at setup rather than hours into a run.
CRITICAL_JOBS=(
  apiserver
  etcd-metrics-proxy
  kube-state-metrics
  runai-scheduler-default
  runai-node-exporter
)

API_TIMEOUT_SECS="${API_TIMEOUT_SECS:-600}"

log() { echo "[setup-monitoring] $*"; }
die() { echo "[setup-monitoring] ERROR: $*" >&2; exit 1; }

# ---------------------------------------------------------------------------
# 0. Wait for the API server
# ---------------------------------------------------------------------------
# A freshly provisioned cluster reports ready before its API server load balancer
# has finished registering targets, so the first kubectl call can fail on a TCP
# timeout even though DNS resolves. Everything below assumes a live API, so wait
# for it here rather than letting an unrelated step fail with a connection error.
log "waiting for the cluster API to become reachable (timeout ${API_TIMEOUT_SECS}s)"
deadline=$((SECONDS + API_TIMEOUT_SECS))
until kubectl get --raw /readyz --request-timeout=10s >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    die "cluster API not reachable after ${API_TIMEOUT_SECS}s: $(kubectl get --raw /readyz --request-timeout=10s 2>&1 | tail -1)"
  fi
  sleep 10
done
log "API server is ready"

# ---------------------------------------------------------------------------
# 1. Manifests
# ---------------------------------------------------------------------------
# etcd listens for metrics on loopback only, so without the socat proxy Prometheus
# cannot scrape it and validity gates 2-4 are permanently UNAVAILABLE.
log "applying etcd metrics proxy and runai-backend ServiceMonitor"
kubectl apply -f "${MANIFESTS_DIR}/etcd-metrics-proxy.yaml"
kubectl apply -f "${MANIFESTS_DIR}/runai-backend-servicemonitor.yaml"

# ---------------------------------------------------------------------------
# 2. Prometheus admin API
# ---------------------------------------------------------------------------
# This must be set on the RunaiConfig, not on the Prometheus CR. The operator's
# SetCommonPrometheusFields assigns the whole Prometheus spec from
# RunaiConfig.spec.prometheus.spec on every reconcile, so a direct patch of the
# Prometheus CR is silently reverted.
current="$(kubectl get runaiconfig "${RUNAICONFIG_NAME}" -n "${NAMESPACE}" \
  -o jsonpath='{.spec.prometheus.spec.enableAdminAPI}' 2>/dev/null || true)"
current_retention="$(kubectl get runaiconfig "${RUNAICONFIG_NAME}" -n "${NAMESPACE}" \
  -o jsonpath='{.spec.prometheus.spec.retention}' 2>/dev/null || true)"

log "waiting for the Prometheus StatefulSet to exist"
for _ in $(seq 1 60); do
  sts="$(kubectl get statefulset -n "${NAMESPACE}" -l operator.prometheus.io/name \
    -o name 2>/dev/null | head -1)"
  [[ -n "${sts:-}" ]] && break
  sleep 5
done
[[ -n "${sts:-}" ]] || die "no Prometheus StatefulSet appeared in namespace ${NAMESPACE}"

# Both settings ride one patch, so a cluster missing either takes a single rollout.
if [[ "${current}" == "true" && "${current_retention}" == "${PROMETHEUS_RETENTION}" ]]; then
  log "admin API and ${PROMETHEUS_RETENTION} retention already set on runaiconfig/${RUNAICONFIG_NAME}"
else
  # The patch reaches the StatefulSet via two controllers, so rollout status straight
  # after it returns success against the old spec. metadata.generation only moves once
  # the change actually lands, so wait for that first; rollout status is then valid.
  gen_before="$(kubectl get "${sts}" -n "${NAMESPACE}" -o jsonpath='{.metadata.generation}')"
  log "setting admin API and ${PROMETHEUS_RETENTION} retention via runaiconfig/${RUNAICONFIG_NAME} (${sts} generation ${gen_before}, retention was ${current_retention:-unset})"
  kubectl patch runaiconfig "${RUNAICONFIG_NAME}" -n "${NAMESPACE}" --type=merge \
    -p "{\"spec\":{\"prometheus\":{\"spec\":{\"enableAdminAPI\":true,\"retention\":\"${PROMETHEUS_RETENTION}\"}}}}"

  log "waiting for the operator to apply it to ${sts} (timeout ${RECONCILE_TIMEOUT_SECS}s)"
  deadline=$((SECONDS + RECONCILE_TIMEOUT_SECS))
  while [[ "$(kubectl get "${sts}" -n "${NAMESPACE}" -o jsonpath='{.metadata.generation}')" == "${gen_before}" ]]; do
    if (( SECONDS >= deadline )); then
      die "${sts} generation never moved past ${gen_before}. The Run:ai operator did not reconcile runaiconfig into the Prometheus CR — check its logs."
    fi
    sleep 5
  done
fi

log "waiting for the Prometheus StatefulSet to roll out (timeout ${ROLLOUT_TIMEOUT})"
kubectl rollout status "${sts}" -n "${NAMESPACE}" --timeout="${ROLLOUT_TIMEOUT}"

# The CR is the input; the pod is the proof. Verify the flags actually reached the
# container rather than trusting that the patch stuck.
log "verifying --web.enable-admin-api and --storage.tsdb.retention.time reached the Prometheus container"
pod="$(kubectl get pods -n "${NAMESPACE}" -l app.kubernetes.io/name=prometheus \
  -o jsonpath='{.items[0].metadata.name}')"
[[ -n "${pod}" ]] || die "no Prometheus pod found in namespace ${NAMESPACE}"

pod_yaml="$(kubectl get pod "${pod}" -n "${NAMESPACE}" -o yaml)"
if ! grep -q -- '--web.enable-admin-api' <<<"${pod_yaml}"; then
  die "admin API flag is not present on pod ${pod}. The RunaiConfig patch did not reach the Prometheus spec; inspect the operator logs before continuing."
fi
if ! grep -q -- "--storage.tsdb.retention.time=${PROMETHEUS_RETENTION}" <<<"${pod_yaml}"; then
  die "retention is not ${PROMETHEUS_RETENTION} on pod ${pod}. A run longer than the retention window loses its own early data, so stop here rather than measure against it."
fi
log "admin API and ${PROMETHEUS_RETENTION} retention confirmed on pod ${pod}"

# ---------------------------------------------------------------------------
# 3. Scrape target verification
# ---------------------------------------------------------------------------
log "waiting for the etcd-metrics-proxy DaemonSet"
kubectl rollout status daemonset/etcd-metrics-proxy -n "${NAMESPACE}" --timeout=5m

log "verifying scrape targets are up (timeout ${TARGET_TIMEOUT_SECS}s)"
local_port="$(python3 -c 'import socket;s=socket.socket();s.bind(("",0));print(s.getsockname()[1]);s.close()')"
kubectl port-forward -n "${NAMESPACE}" "pod/${pod}" "${local_port}:9090" >/dev/null 2>&1 &
pf_pid=$!
trap 'kill "${pf_pid}" 2>/dev/null || true' EXIT

for _ in $(seq 1 30); do
  curl -sf "http://localhost:${local_port}/-/ready" >/dev/null 2>&1 && break
  sleep 2
done

# Returns True only when the query matched at least one series and every one is up.
# An absent series is a failure: it means the target is not being scraped at all.
query_up() {
  curl -sf --get "http://localhost:${local_port}/api/v1/query" \
    --data-urlencode "query=$1" 2>/dev/null | python3 -c \
    'import json,sys
d=json.load(sys.stdin)
r=d["data"]["result"]
print(bool(r) and all(float(x["value"][1])==1 for x in r))' 2>/dev/null || echo False
}

deadline=$((SECONDS + TARGET_TIMEOUT_SECS))
while true; do
  down=()
  for job in "${CRITICAL_JOBS[@]}"; do
    [[ "$(query_up "up{job=\"${job}\"}")" == "True" ]] || down+=("${job}")
  done
  # runai-backend services are selected by namespace: the ServiceMonitor matches any
  # service exposing a "metrics" port, so their job names are not fixed.
  [[ "$(query_up 'up{namespace="runai-backend"}')" == "True" ]] || down+=("runai-backend")

  if [[ ${#down[@]} -eq 0 ]]; then
    log "all required scrape targets are up"
    break
  fi
  if (( SECONDS >= deadline )); then
    die "scrape targets still down after ${TARGET_TIMEOUT_SECS}s: ${down[*]}. The validity gates fed by these are UNAVAILABLE for the whole run."
  fi
  log "waiting on: ${down[*]}"
  sleep 10
done

log "done"
