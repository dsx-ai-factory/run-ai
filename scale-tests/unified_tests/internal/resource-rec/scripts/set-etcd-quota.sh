#!/usr/bin/env bash
# Raise etcd's backend quota above the 2 GiB default, for scale levels expected to exceed it.
#
# A Run:ai workload costs ~23 KiB of etcd across six keys (pod, job, podgroup,
# workloadrunner, externalworkload, bindrequest). Once the quota is hit etcd rejects every
# write and the cluster wedges, so this is run at setup time, before any load.
#
# etcd is a kubeadm static pod, so the flag lives in a manifest on the control-plane node
# rather than in any API object. This writes it through a privileged pod; kubelet notices
# the manifest change and restarts etcd on its own.
#
# Usage: ./set-etcd-quota.sh [kubeconfig] [quota-bytes]
set -euo pipefail

export KUBECONFIG="${1:-${KUBECONFIG:-$HOME/.kube/config}}"
QUOTA_BYTES="${2:-8589934592}" # 8 GiB
MANIFEST=/etc/kubernetes/manifests/etcd.yaml
NS=kube-system

quota_flag() {
  kubectl get po -n "$NS" -l component=etcd -o jsonpath='{.items[0].spec.containers[0].command}' 2>/dev/null \
    | tr ',' '\n' | grep -o 'quota-backend-bytes=[0-9]*' | cut -d= -f2 || true
}

NODE=$(kubectl get po -n "$NS" -l component=etcd -o jsonpath='{.items[0].spec.nodeName}')
echo "etcd on node $NODE, target quota $QUOTA_BYTES bytes"

if [[ "$(quota_flag)" == "$QUOTA_BYTES" ]]; then
  echo "quota already set, nothing to do"
  exit 0
fi

# Base64 so the patch survives the trip through JSON and two shells without quoting games.
# kubelet parses every file in staticPodPath, so a stray .bak or .tmp there is a second
# definition of the same pod and wins over the edit. Stage outside the directory and
# overwrite the manifest in place.
PATCH=$(base64 <<PATCH_EOF | tr -d '\n'
set -e
f=/host$MANIFEST
cp \$f /host/etc/kubernetes/etcd.yaml.bak
grep -v quota-backend-bytes \$f > /tmp/etcd.stripped
awk 'index(\$0,"--data-dir=/var/lib/etcd") && !seen {print "    - --quota-backend-bytes=$QUOTA_BYTES"; seen=1} {print}' /tmp/etcd.stripped > /tmp/etcd.new
cat /tmp/etcd.new > \$f
grep -q quota-backend-bytes=$QUOTA_BYTES \$f
PATCH_EOF
)

kubectl run "etcd-quota-$$" -n "$NS" --restart=Never --rm -i --quiet --image=busybox:1.36 \
  --overrides="$(cat <<JSON
{
  "spec": {
    "nodeName": "$NODE",
    "tolerations": [{"operator": "Exists"}],
    "containers": [{
      "name": "patch",
      "image": "busybox:1.36",
      "securityContext": {"privileged": true},
      "command": ["sh","-c","echo $PATCH | base64 -d | sh"],
      "volumeMounts": [{"name":"host","mountPath":"/host"}]
    }],
    "volumes": [{"name":"host","hostPath":{"path":"/"}}]
  }
}
JSON
)" || echo "patch pod did not attach cleanly; verifying against the running etcd instead"

echo "waiting for etcd to restart with the new flag"
for _ in $(seq 1 60); do
  sleep 10
  ready=$(kubectl get po -n "$NS" -l component=etcd \
    -o jsonpath='{.items[0].status.containerStatuses[0].ready}' 2>/dev/null || true)
  if [[ "$(quota_flag)" == "$QUOTA_BYTES" && "$ready" == "true" ]]; then
    echo "etcd restarted with quota $QUOTA_BYTES"
    exit 0
  fi
done

echo "ERROR: etcd did not come back with quota $QUOTA_BYTES"
exit 1
