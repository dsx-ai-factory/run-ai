#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
# Acceptance tests (standalone acceptance suite). Run against a cluster that already has:
#   - the Track A prerequisites (test/acceptance/00-prereqs.sh), Gateway and simulators (01, 02)
#   - our CRD installed (make install) and the controller running (make run, or a deployed manager)
#
#   1. GOLDEN DIFF  apply the llama3 CR → the 10 children equal test/testdata/golden/03,04,05 applied by hand
#   2. DRIFT        delete the EPP Deployment → the controller recreates it
#   3. WEIGHTS      change epp.scheduling.scorers → ConfigMap + config hash change, EPP rolls, new pod loads the config
#   4. REWRITE      add rewrites[] → InferenceModelRewrite owned by the endpoint, canary visible, EPP not restarted
#   5. CLEANUP      delete the CR → the children vanish, Gateway / pods / CRDs stay
#
# The traffic test is test/acceptance/08-smoke-test.sh, run it after step 1 with the same KUBECONFIG.
set -euo pipefail

python3 -c 'import yaml' 2>/dev/null || { echo "PyYAML is required: python3 -m pip install PyYAML" >&2; exit 1; }

NS="${NS:-model-serving}"
GOLDEN_NS="${GOLDEN_NS:-golden}"
GW="${GW:-inference-agw}"
EP="${EP:-llama3}"
SAMPLE="${SAMPLE:-config/samples/inference_v1alpha1_inferenceendpoint.yaml}"
GOLDEN_DIR="${GOLDEN_DIR:-test/testdata/golden}"
READY_TIMEOUT="${READY_TIMEOUT:-180}"
ENDPOINT_LABEL="inference.platform.run.ai/endpoint"

PASS=0; FAIL=0
ok()   { printf '  \033[1;32m✓ %s\033[0m\n' "$*"; PASS=$((PASS+1)); }
bad()  { printf '  \033[1;31m✗ %s\033[0m\n' "$*"; FAIL=$((FAIL+1)); }
step() { printf '\n\033[1m[%s] %s\033[0m\n' "$1" "$2"; }

# kind/name of the ten namespaced children, in apply order
CHILDREN=(
  "inferencepool.inference.networking.k8s.io/${EP}-pool"
  "serviceaccount/${EP}-epp"
  "configmap/${EP}-epp"
  "role/${EP}-epp-non-sa"
  "role/${EP}-epp-sa"
  "rolebinding/${EP}-epp-non-sa"
  "rolebinding/${EP}-epp-sa"
  "service/${EP}-epp"
  "deployment/${EP}-epp"
  "httproute.gateway.networking.k8s.io/${EP}-route"
)

# export_children <namespace> <outfile>: fetch the ten children as JSON and normalise away everything
# the cluster or the owner adds, so two stacks created by different hands can be compared.
export_children() {
  local ns="$1" out="$2"
  : > "$out"
  for ref in "${CHILDREN[@]}"; do
    kubectl get "$ref" -n "$ns" -o json 2>/dev/null \
      | python3 -c '
import json, sys
ns, label = sys.argv[1], sys.argv[2]
obj = json.load(sys.stdin)
meta = obj.get("metadata", {})
for k in ("uid", "resourceVersion", "creationTimestamp", "generation", "managedFields",
          "ownerReferences", "selfLink", "annotations"):
    meta.pop(k, None)
labels = meta.get("labels") or {}
labels.pop("lab.track", None)
labels.pop(label, None)
if labels: meta["labels"] = labels
else: meta.pop("labels", None)
obj.pop("status", None)
if obj.get("kind") == "Deployment":         # the controller stamps a config hash the hand-written golden lacks;
    tmeta = obj.get("spec", {}).get("template", {}).get("metadata", {})   # restartedAt is left by kubectl rollout restart
    ann = tmeta.get("annotations") or {}
    ann.pop("inference.platform.run.ai/config-hash", None)
    ann.pop("kubectl.kubernetes.io/restartedAt", None)
    if ann: tmeta["annotations"] = ann
    else: tmeta.pop("annotations", None)
if obj.get("kind") == "ConfigMap":          # compare YAML values, ignoring comments and scalar quoting
    import yaml
    obj["data"]["default-plugins.yaml"] = yaml.safe_load(obj["data"]["default-plugins.yaml"])
if obj.get("kind") == "Service":            # allocated by the cluster, never equal between two Services
    obj["spec"].pop("clusterIP", None)
    obj["spec"].pop("clusterIPs", None)
text = json.dumps(obj, sort_keys=True, indent=1)
print(text.replace("\"%s\"" % ns, "\"NS\"").replace("%s" % ns, "NS"))
' "$ns" "$ENDPOINT_LABEL" >> "$out" || echo "MISSING $ref" >> "$out"
  done
}

# newest_epp_pod prints the most recently created EPP pod, terminating or not, so a rollout in
# progress never makes us read the old pod by accident.
newest_epp_pod() {
  kubectl get pods -n "$NS" -l inferencepool="${EP}-epp" --sort-by=.metadata.creationTimestamp -o name 2>/dev/null | tail -1
}

wait_ready() {
  local deadline=$((SECONDS + READY_TIMEOUT))
  while (( SECONDS < deadline )); do
    local status reason
    status=$(kubectl get inferenceendpoint "$EP" -n "$NS" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || true)
    reason=$(kubectl get inferenceendpoint "$EP" -n "$NS" -o jsonpath='{.status.conditions[?(@.type=="Ready")].reason}' 2>/dev/null || true)
    if [[ "$status" == "True" ]]; then return 0; fi
    printf '    waiting: Ready=%s reason=%s\n' "${status:-<none>}" "${reason:-<none>}"
    sleep 5
  done
  return 1
}

# apply_golden_namespaced <file> <namespace>: apply only the namespaced documents of a golden file, with the
# lab namespace renamed. The cluster-scoped metrics RBAC in 04-epp.yaml is skipped on purpose: it is not
# part of the v0 diff, and on a lab cluster those names belong to the real RBAC the lab installed.
apply_golden_namespaced() {
  local file="$1" ns="$2"
  sed "s/${NS}/${ns}/g" "$file" \
    | kubectl create --dry-run=client -o json -f - \
    | python3 -c '
import json, sys
# kubectl emits one JSON object per document, concatenated; decode them one by one.
text, pos, items, dec = sys.stdin.read(), 0, [], json.JSONDecoder()
while pos < len(text):
    while pos < len(text) and text[pos].isspace(): pos += 1
    if pos >= len(text): break
    obj, pos = dec.raw_decode(text, pos)
    items.extend(obj.get("items", [obj]))
keep = [i for i in items if i.get("kind") not in ("ClusterRole", "ClusterRoleBinding")]
json.dump({"apiVersion": "v1", "kind": "List", "items": keep}, sys.stdout)
' \
    | kubectl apply -n "$ns" -f - >/dev/null
}

cleanup_golden() {
  kubectl delete namespace "$GOLDEN_NS" --ignore-not-found --wait=false >/dev/null 2>&1 || true
}
trap cleanup_golden EXIT

step 0 "preconditions"
kubectl get gateway "$GW" -n "$NS" >/dev/null && ok "Gateway $GW exists in $NS" || { bad "Gateway $GW missing in $NS, run test/acceptance/00-prereqs.sh and 01"; exit 1; }
kubectl get crd inferenceendpoints.inference.platform.run.ai >/dev/null 2>&1 && ok "InferenceEndpoint CRD installed" || { bad "CRD missing, run make install"; exit 1; }
sims=$(kubectl get pods -n "$NS" -l app=vllm-llama3 --field-selector=status.phase=Running -o name | wc -l | tr -d ' ')
[[ "$sims" -ge 1 ]] && ok "$sims simulator pod(s) running" || bad "no simulator pods with app=vllm-llama3 in $NS"

step 1 "GOLDEN DIFF · one CR in, the Track A stack out"
kubectl apply -f "$SAMPLE" >/dev/null && ok "applied $SAMPLE"
if wait_ready; then
  ok "InferenceEndpoint $EP is Ready=True"
else
  bad "InferenceEndpoint $EP not Ready after ${READY_TIMEOUT}s"
  kubectl get inferenceendpoint "$EP" -n "$NS" -o jsonpath='{.status.conditions}'; echo
fi
for ref in "${CHILDREN[@]}"; do
  if kubectl get "$ref" -n "$NS" -o jsonpath='{.metadata.labels.inference\.platform\.run\.ai/endpoint} {.metadata.ownerReferences[0].kind}' 2>/dev/null | grep -q "^${EP} InferenceEndpoint$"; then
    ok "$ref exists, labelled and owned by the endpoint"
  else
    bad "$ref missing, or not labelled/owned by the endpoint"
  fi
done

kubectl create namespace "$GOLDEN_NS" --dry-run=client -o yaml | kubectl apply -f - >/dev/null
for f in 03-inferencepool.yaml 04-epp.yaml 05-httproute.yaml; do
  apply_golden_namespaced "${GOLDEN_DIR}/${f}" "$GOLDEN_NS"
done
ok "applied the namespaced golden objects by hand into namespace $GOLDEN_NS"
sleep 2
TMP=$(mktemp -d)
export_children "$NS" "$TMP/controller.json"
export_children "$GOLDEN_NS" "$TMP/golden.json"
if diff -u "$TMP/golden.json" "$TMP/controller.json" > "$TMP/diff.txt"; then
  ok "controller output is identical to the hand-written Track A manifests (10 objects)"
else
  bad "controller output differs from golden, see below"
  cat "$TMP/diff.txt"
fi
cleanup_golden

step 2 "DRIFT · delete the EPP Deployment, expect it back"
kubectl delete deployment "${EP}-epp" -n "$NS" --wait=true >/dev/null && ok "deleted deployment ${EP}-epp"
# kubectl wait fails fast on NotFound, and the controller usually recreates the Deployment within a
# couple of seconds, so first wait for it to exist, then for it to become Available.
deadline=$((SECONDS + 60)); back=0
while (( SECONDS < deadline )); do
  kubectl get deployment "${EP}-epp" -n "$NS" >/dev/null 2>&1 && { back=1; break; }
  sleep 2
done
if (( back == 1 )); then
  ok "deployment ${EP}-epp recreated by the controller"
else
  bad "deployment ${EP}-epp not recreated within 60s"
fi
if kubectl wait deployment "${EP}-epp" -n "$NS" --for=condition=Available --timeout=180s >/dev/null 2>&1; then
  ok "deployment ${EP}-epp Available again"
else
  bad "deployment ${EP}-epp not Available within 180s"
fi
wait_ready && ok "InferenceEndpoint $EP Ready=True again" || bad "InferenceEndpoint $EP not Ready after drift"

step 3 "WEIGHTS · change scorer weights on the CR, expect the EPP to roll with the new config"
hash_before=$(kubectl get deployment "${EP}-epp" -n "$NS" -o jsonpath='{.spec.template.metadata.annotations.inference\.platform\.run\.ai/config-hash}')
pod_before=$(newest_epp_pod)
[[ -n "$hash_before" ]] && ok "pod template carries a config hash (${hash_before:0:12}…)" || bad "no config-hash annotation on the EPP Deployment"
kubectl patch inferenceendpoint "$EP" -n "$NS" --type merge -p '{"spec":{"epp":{"scheduling":{"scorers":[{"type":"prefixCache","weight":100}]}}}}' >/dev/null && ok "patched CR: scorers = prefixCache 100"
deadline=$((SECONDS + 60)); hash_after="$hash_before"
while (( SECONDS < deadline )) && [[ "$hash_after" == "$hash_before" ]]; do
  sleep 2; hash_after=$(kubectl get deployment "${EP}-epp" -n "$NS" -o jsonpath='{.spec.template.metadata.annotations.inference\.platform\.run\.ai/config-hash}')
done
[[ "$hash_after" != "$hash_before" ]] && ok "config hash changed (${hash_after:0:12}…)" || bad "config hash did not change within 60s"
config=$(kubectl get configmap "${EP}-epp" -n "$NS" -o jsonpath='{.data.default-plugins\.yaml}')
python3 -c '
import sys, yaml
config = yaml.safe_load(sys.stdin)
profile = next(p for p in config["schedulingProfiles"] if p["name"] == "default")
sys.exit(0 if {"pluginRef": "prefix-cache-scorer", "weight": 100} in profile["plugins"] else 1)
' <<<"$config" && ok "ConfigMap now lists prefix-cache-scorer with weight 100" || bad "ConfigMap not updated"
python3 -c '
import sys, yaml
config = yaml.safe_load(sys.stdin)
present = any(p["type"] == "queue-scorer" for p in config["plugins"]) or any(
    p["pluginRef"] == "queue-scorer" for profile in config["schedulingProfiles"] for p in profile["plugins"])
sys.exit(1 if present else 0)
' <<<"$config" && ok "ConfigMap no longer lists queue-scorer" || bad "ConfigMap still lists queue-scorer or could not be parsed"
kubectl rollout status deployment "${EP}-epp" -n "$NS" --timeout=180s >/dev/null 2>&1 && ok "EPP Deployment rolled out" || bad "EPP rollout did not finish within 180s"
pod_after=$(newest_epp_pod)
[[ -n "$pod_after" && "$pod_after" != "$pod_before" ]] && ok "new EPP pod ${pod_after#pod/}" || bad "EPP pod did not change ($pod_before → $pod_after)"
effective=$(kubectl logs -n "$NS" "$pod_after" 2>/dev/null | grep -m1 'Effective raw configuration' || true)
if grep -q 'prefix-cache-scorer' <<<"$effective" && ! grep -q 'queue-scorer' <<<"$effective"; then
  ok "new EPP pod loaded a config with prefix-cache-scorer and without queue-scorer"
else
  bad "new EPP pod's effective config is not the new one: ${effective:0:160}"
fi
wait_ready && ok "InferenceEndpoint $EP Ready=True with the new weights" || bad "InferenceEndpoint $EP not Ready after the weight change"
kubectl patch inferenceendpoint "$EP" -n "$NS" --type merge -p '{"spec":{"epp":{"scheduling":null}}}' >/dev/null && ok "restored default weights (30/20/10)"
# wait for the controller to write the default hash back before asking about the rollout, otherwise
# rollout status answers for the previous, already finished rollout
deadline=$((SECONDS + 60)); hash_restored=""
while (( SECONDS < deadline )) && [[ "$hash_restored" != "$hash_before" ]]; do
  sleep 2; hash_restored=$(kubectl get deployment "${EP}-epp" -n "$NS" -o jsonpath='{.spec.template.metadata.annotations.inference\.platform\.run\.ai/config-hash}')
done
[[ "$hash_restored" == "$hash_before" ]] && ok "config hash back to the default" || bad "config hash did not return to the default within 60s"
kubectl rollout status deployment "${EP}-epp" -n "$NS" --timeout=180s >/dev/null 2>&1 && ok "EPP rolled back to the default config" || bad "EPP rollback did not finish"

step 4 "REWRITE · add a canary rewrite on the CR, expect a live InferenceModelRewrite and no EPP restart"
gen_before=$(kubectl get deployment "${EP}-epp" -n "$NS" -o jsonpath='{.metadata.generation}')
kubectl patch inferenceendpoint "$EP" -n "$NS" --type merge -p '{"spec":{"rewrites":[{"model":"llama3","targets":[{"model":"llama3","weight":50},{"model":"llama3-v2","weight":50}]}]}}' >/dev/null && ok "patched CR: rewrite llama3 → 50 % llama3 / 50 % llama3-v2"
deadline=$((SECONDS + 30)); rw=""
while (( SECONDS < deadline )) && [[ -z "$rw" ]]; do sleep 2; rw=$(kubectl get inferencemodelrewrite "${EP}-rewrite" -n "$NS" -o jsonpath='{.metadata.ownerReferences[0].kind}' 2>/dev/null || true); done
[[ "$rw" == "InferenceEndpoint" ]] && ok "InferenceModelRewrite ${EP}-rewrite exists and is owned by the endpoint" || bad "InferenceModelRewrite not created within 30s"
kubectl port-forward -n "$NS" "svc/$GW" 18089:80 >/dev/null 2>&1 & pf_pid=$!; sleep 3
since=$(date -u +%Y-%m-%dT%H:%M:%SZ); sleep 1; req_pids=()
for i in $(seq 1 40); do
  curl -s -m 30 http://localhost:18089/v1/chat/completions -H "Host: ${HOSTNAME_HEADER:-llama3.test}" -H 'Content-Type: application/json' \
    -d '{"model":"llama3","messages":[{"role":"user","content":"Hello!"}],"max_tokens":8}' >/dev/null & req_pids+=($!)
done
wait "${req_pids[@]}"; sleep 3; kill "$pf_pid" 2>/dev/null
rewritten=$(kubectl logs -n "$NS" deploy/"${EP}-epp" --since-time="$since" 2>/dev/null | grep -c '"targetModelName":"llama3-v2"' || true)
(( rewritten > 0 )) && ok "$rewritten of the last requests were rewritten to llama3-v2 by the EPP" || bad "no request was rewritten to llama3-v2"
gen_after=$(kubectl get deployment "${EP}-epp" -n "$NS" -o jsonpath='{.metadata.generation}')
[[ "$gen_after" == "$gen_before" ]] && ok "EPP Deployment untouched (generation $gen_after): the rewrite applied live" || bad "EPP Deployment changed on a rewrite (generation $gen_before → $gen_after)"
kubectl patch inferenceendpoint "$EP" -n "$NS" --type merge -p '{"spec":{"rewrites":null}}' >/dev/null && ok "removed the rewrite from the CR"
deadline=$((SECONDS + 30)); gone=0
while (( SECONDS < deadline )); do kubectl get inferencemodelrewrite "${EP}-rewrite" -n "$NS" >/dev/null 2>&1 || { gone=1; break; }; sleep 2; done
(( gone == 1 )) && ok "InferenceModelRewrite deleted with the field" || bad "InferenceModelRewrite still present after 30s"

step 5 "CLEANUP · delete the CR, expect only our objects to vanish"
kubectl delete inferenceendpoint "$EP" -n "$NS" --wait=true >/dev/null && ok "deleted InferenceEndpoint $EP"
deadline=$((SECONDS + 60)); left=1
while (( SECONDS < deadline )); do
  left=0
  for ref in "${CHILDREN[@]}"; do kubectl get "$ref" -n "$NS" >/dev/null 2>&1 && left=$((left+1)); done
  (( left == 0 )) && break
  sleep 3
done
(( left == 0 )) && ok "all 10 children gone" || bad "$left child object(s) still present after 60s"
kubectl get gateway "$GW" -n "$NS" >/dev/null 2>&1 && ok "Gateway $GW untouched" || bad "Gateway $GW disappeared"
kubectl get deployment vllm-llama3 -n "$NS" >/dev/null 2>&1 && ok "simulator Deployment untouched" || bad "simulator Deployment disappeared"
kubectl get crd inferencepools.inference.networking.k8s.io httproutes.gateway.networking.k8s.io >/dev/null 2>&1 && ok "upstream CRDs untouched" || bad "upstream CRDs disappeared"

printf '\n\033[1mresult: %d passed, %d failed\033[0m\n' "$PASS" "$FAIL"
(( FAIL == 0 ))
