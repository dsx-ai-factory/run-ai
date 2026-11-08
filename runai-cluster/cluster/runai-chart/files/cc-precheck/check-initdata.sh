#!/usr/bin/env bash
set -euo pipefail

api=https://kubernetes.default.svc
service_account_dir=${SERVICE_ACCOUNT_DIR:-/var/run/secrets/kubernetes.io/serviceaccount}
token=$(cat "$service_account_dir/token")
ca=$service_account_dir/ca.crt
namespace=${POD_NAMESPACE:?}
name=${INITDATA_CONFIGMAP:?}

if ! configmap=$(curl -fsS --cacert "$ca" -H "Authorization: Bearer $token" \
  "$api/api/v1/namespaces/$namespace/configmaps/$name" 2>/dev/null); then
  echo "ConfigMap '$name' was not found in namespace '$namespace'." >&2
  echo "global.confidentialComputing.initdataConfigMap names the initdata document that carries the attestation configuration. Create it before installing." >&2
  exit 1
fi

if ! jq -e '(.data["initdata.toml"] // "") | length > 0' <<<"$configmap" >/dev/null; then
  echo "ConfigMap '$name' in namespace '$namespace' has no 'initdata.toml' key, or it is empty." >&2
  echo "The key must hold the initdata document as plain text." >&2
  exit 1
fi
