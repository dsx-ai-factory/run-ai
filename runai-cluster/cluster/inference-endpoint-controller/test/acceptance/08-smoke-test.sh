#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
# Track A step 8 · the page's smoke-test guide as a script.
#   ./08-smoke-test.sh          checks 1-7
set -uo pipefail

NS="${NS:-model-serving}"
GW="${GW:-inference-agw}"
HOST="${HOST:-llama3.test}"
LOCAL_PORT="${LOCAL_PORT:-8080}"
PASS=0; FAIL=0
ok()   { printf '  \033[1;32m✓ %s\033[0m\n' "$*"; PASS=$((PASS+1)); }
bad()  { printf '  \033[1;31m✗ %s\033[0m\n' "$*"; FAIL=$((FAIL+1)); }
step() { printf '\n\033[1m[%s] %s\033[0m\n' "$1" "$2"; }
pf_pids=()
cleanup() { for p in "${pf_pids[@]:-}"; do [ -n "$p" ] && kill "$p" 2>/dev/null; done; }
trap cleanup EXIT

gateway_svc() {
  kubectl get svc -n "$NS" -l "gateway.networking.k8s.io/gateway-name=${GW}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null
}
chat() { # chat <extra curl args...>
  curl -s -m 30 "http://localhost:${LOCAL_PORT}/v1/chat/completions" -H "Host: ${HOST}" -H 'Content-Type: application/json' "$@" \
    -d '{"model":"llama3","messages":[{"role":"user","content":"Hello!"}],"max_tokens":20}'
}

step 1 "Stack health: EPP and sims Running"
kubectl get pods -n "$NS" -o wide
[ "$(kubectl get deploy llama3-epp -n "$NS" -o jsonpath='{.status.availableReplicas}')" = "1" ] && ok "EPP available" || bad "EPP not available"
[ "$(kubectl get deploy vllm-llama3 -n "$NS" -o jsonpath='{.status.availableReplicas}')" = "3" ] && ok "3 sims available" || bad "sims not all available"

step 2 "HTTPRoute backend is an InferencePool, not a Service"
kind=$(kubectl get httproute llama3-route -n "$NS" -o jsonpath='{.spec.rules[0].backendRefs[0].kind}')
[ "$kind" = "InferencePool" ] && ok "backendRef kind=InferencePool" || bad "backendRef kind=$kind"

step 3 "HTTPRoute accepted by agentgateway, refs resolved"
conds=$(kubectl get httproute llama3-route -n "$NS" -o jsonpath='{range .status.parents[*].conditions[*]}{.type}={.status} {end}')
echo "  $conds"
echo "$conds" | grep -q 'Accepted=True'     && ok "Accepted=True"     || bad "Accepted not True"
echo "$conds" | grep -q 'ResolvedRefs=True' && ok "ResolvedRefs=True" || bad "ResolvedRefs not True"

step 3b "EPP runs with TLS (secure-serving true), the page's agentgateway requirement"
flag=$(kubectl logs -n "$NS" deploy/llama3-epp 2>/dev/null | grep -o '"secure-serving":[a-z]*' | head -1)
echo "  ${flag:-<flag line not found>}"
[ "$flag" = '"secure-serving":true' ] && ok "secure-serving true in EPP flags" || echo "  (could not confirm from logs; not fatal, agentgateway will tell us in check 6)"

step 4 "EPP sees the pool's pods (metrics, bearer token required)"
kubectl port-forward -n "$NS" svc/llama3-epp 9090:9090 >/dev/null 2>&1 & pf_pids+=($!); sleep 3
TOKEN=$(kubectl create token llama3-epp -n "$NS" --duration=10m 2>/dev/null)
metrics=$(curl -s -m 10 -H "Authorization: Bearer ${TOKEN}" localhost:9090/metrics)
ready=$(echo "$metrics" | grep -E '^inference_pool_ready_pods' | head -1)
echo "  ${ready:-<no inference_pool_ready_pods line; families: $(echo "$metrics" | grep '^# TYPE inference' | awk '{print $3}' | tr '\n' ' ')>}"
echo "$ready" | grep -qE ' 3(\.0)?$' && ok "inference_pool_ready_pods = 3" || bad "expected 3 ready pods"

step 5 "A sim answers directly (bypass the gateway)"
POD=$(kubectl get pods -n "$NS" -l app=vllm-llama3 -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward -n "$NS" "pod/$POD" 8000:8000 >/dev/null 2>&1 & pf_pids+=($!); sleep 2
r=$(curl -s -m 10 localhost:8000/v1/chat/completions -H 'Content-Type: application/json' -d '{"model":"llama3","messages":[{"role":"user","content":"Hello!"}],"max_tokens":20}')
echo "$r" | grep -q '"choices"' && ok "direct sim response has choices" || bad "direct sim response: $r"

step 6 "End to end: client → agentgateway → HTTPRoute → InferencePool → EPP → sim"
SVC=$(gateway_svc); echo "  gateway service: ${SVC:-<not found>}"
[ -z "$SVC" ] && bad "no Service labeled gateway-name=$GW in $NS" || {
  kubectl port-forward -n "$NS" "svc/$SVC" "${LOCAL_PORT}:80" >/dev/null 2>&1 & pf_pids+=($!); sleep 2
  r=$(chat)
  echo "$r" | grep -q '"choices"' && ok "e2e response has choices" || bad "e2e response: $r"
  kubectl logs -n "$NS" deploy/llama3-epp --tail=100 | grep -qi 'Scheduling succeeded\|scheduling.*succeeded\|selected.*endpoint' \
    && ok "EPP log shows a scheduling decision" || echo "  (no 'Scheduling succeeded' line found at this verbosity)"
}

step 7 "Load awareness: 30 parallel requests, look at which pods were picked"
SVC=$(gateway_svc)
[ -n "$SVC" ] && {
  load_pids=(); for i in $(seq 1 30); do chat >/dev/null & load_pids+=($!); done; wait "${load_pids[@]}"
  echo "  destination endpoints chosen (from EPP logs):"
  kubectl logs -n "$NS" deploy/llama3-epp --tail=400 | grep -oE '10\.[0-9]+\.[0-9]+\.[0-9]+:8000' | sort | uniq -c | sort -rn | head -5
  echo "  a spread that is NOT perfectly even is expected: the EPP prefers pods with shorter queues"
  ok "load test sent; inspect the spread above"
}

printf '\n\033[1mResult: %d passed, %d failed\033[0m\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
