#!/usr/bin/env bash
# Assemble one report from every Linux node matching the configured selector.
set -euo pipefail

api=https://kubernetes.default.svc
service_account_dir=${SERVICE_ACCOUNT_DIR:-/var/run/secrets/kubernetes.io/serviceaccount}
token=$(cat "$service_account_dir/token")
ca=$service_account_dir/ca.crt
namespace=${POD_NAMESPACE:?}
owner=${OWNER_LABEL:?}
selector=${NODE_SELECTOR:?}
checker_version=${CHECKER_VERSION:?}
dispatcher_job=${DISPATCHER_JOB_NAME:-$owner-dispatcher}

get() {
  curl -fsS --cacert "$ca" -H "Authorization: Bearer $token" "$api$1"
}
list() {
  curl -fsS --cacert "$ca" -H "Authorization: Bearer $token" \
    --get --data-urlencode "labelSelector=$2" "$api$1"
}

nodes=$(list /api/v1/nodes "$selector")
jobs=$(list "/apis/batch/v1/namespaces/$namespace/jobs" "$owner/owner=$owner")
dispatcher=$(get "/apis/batch/v1/namespaces/$namespace/jobs/$dispatcher_job")
if [[ ${WAIT_FOR_DISPATCHER:-false} == true ]]; then
  for ((attempt=0; attempt<600; attempt++)); do
    if jq -e 'any(.status.conditions[]?; .type == "Failed" and .status == "True")' <<<"$dispatcher" >/dev/null; then
      echo "Dispatcher Job $dispatcher_job failed; pre-flight collection is incomplete" >&2
      exit 1
    fi
    if jq -e 'any(.status.conditions[]?; .type == "Complete" and .status == "True")' <<<"$dispatcher" >/dev/null; then break; fi
    sleep 2
    dispatcher=$(get "/apis/batch/v1/namespaces/$namespace/jobs/$dispatcher_job")
  done
  if ! jq -e 'any(.status.conditions[]?; .type == "Complete" and .status == "True")' <<<"$dispatcher" >/dev/null; then
    echo "Dispatcher Job $dispatcher_job did not complete before the collection deadline" >&2
    exit 1
  fi
  # Probe Jobs can be created while waiting; collect the final set.
  jobs=$(list "/apis/batch/v1/namespaces/$namespace/jobs" "$owner/owner=$owner")
fi
dispatcher_uid=$(jq -r '.metadata.uid' <<<"$dispatcher")
expected=$(jq -c '[.items[].metadata.name] | sort' <<<"$nodes")
inventory=$(jq -c '[.items[] | {name:.metadata.name,uid:.metadata.uid,machineID:.status.nodeInfo.machineID,bootID:.status.nodeInfo.bootID}] | sort_by(.name)' <<<"$nodes")
if ! jq -e 'all(.[]; [.name,.uid,.machineID,.bootID] | all(.[]; type == "string" and length > 0))' <<<"$inventory" >/dev/null; then
  echo 'Selected nodes have incomplete machine or boot identities' >&2
  exit 1
fi
if [[ $(jq length <<<"$expected") == 0 ]]; then
  echo "No Linux nodes match the pre-flight selector: $selector" >&2
  exit 1
fi

results='[]'
while IFS= read -r job; do
  name=$(jq -r '.metadata.name' <<<"$job")
  node=$(jq -r '.spec.template.spec.nodeName' <<<"$job")
  if [[ $(jq -r '.status.succeeded // 0' <<<"$job") != 1 ]]; then
    echo "Node Job $name did not succeed" >&2
    exit 1
  fi
  pods=$(list "/api/v1/namespaces/$namespace/pods" "batch.kubernetes.io/job-name=$name")
  if [[ $(jq '.items | length' <<<"$pods") != 1 ]]; then
    echo "Expected one pod for $name" >&2
    exit 1
  fi
  pod=$(jq -r '.items[0].metadata.name' <<<"$pods")
  report=$(get "/api/v1/namespaces/$namespace/pods/$pod/log?container=precheck")
  identity=$(jq -c --arg node "$node" '.[] | select(.name == $node)' <<<"$inventory")
  if [[ -z $identity ]] || ! jq -e --argjson identity "$identity" --arg version "$checker_version" \
    '.kind == "krab-node-precheck" and .schemaVersion == 2 and
     .node.machineID == $identity.machineID and .node.bootID == $identity.bootID and
     .node.checkerVersion == $version' <<<"$report" >/dev/null; then
    echo "Invalid report from $node" >&2
    exit 1
  fi
  results=$(jq -c --arg node "$node" --argjson report "$report" \
    '. + [{nodeName:$node,report:$report}]' <<<"$results")
done < <(jq -c --arg uid "$dispatcher_uid" '.items[] | select(any(.metadata.ownerReferences[]?; .uid == $uid))' <<<"$jobs")

actual=$(jq -c '[.[].nodeName] | sort' <<<"$results")
if [[ $actual != "$expected" ]]; then
  echo "Node coverage mismatch: expected $expected, got $actual" >&2
  exit 1
fi
current=$(list /api/v1/nodes "$selector")
current=$(jq -c '[.items[] | {name:.metadata.name,uid:.metadata.uid,machineID:.status.nodeInfo.machineID,bootID:.status.nodeInfo.bootID}] | sort_by(.name)' <<<"$current")
if [[ $current != "$inventory" ]]; then
  echo 'Selected nodes or their identities changed during pre-flight; rerun the check' >&2
  exit 1
fi
jq -n --argjson results "$results" --arg selector "$selector" --argjson inventory "$inventory" \
  '{kind:"krab-cluster-precheck",schemaVersion:2,scope:{nodeSelector:$selector,nodes:$inventory},nodes:$results}'
