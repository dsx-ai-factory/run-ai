#!/usr/bin/env bash

set -euo pipefail

ref="${1:?usage: $0 <KAI tag, branch, or commit, for example v0.16.1, main, or 6ee3494ea357>}"
module="github.com/kai-scheduler/KAI-scheduler"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/../.." && pwd)"

tmp_mod_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_mod_dir}"' EXIT
resolved="$(
  cd "${tmp_mod_dir}"
  go mod init kai-version-resolver >/dev/null 2>&1
  go list -m -f '{{.Version}}|{{.Origin.Hash}}' "${module}@${ref}"
)"
go_module_version="${resolved%%|*}"
commit_hash="${resolved#*|}"

if [[ -z "${go_module_version}" || -z "${commit_hash}" || "${resolved}" != *"|"* ]]; then
  echo "KAI ref ${ref} could not be resolved" >&2
  exit 1
fi

if [[ "${go_module_version}" =~ [.-][0-9]{14}-[0-9a-f]{12}$ ]]; then
  chart_version="0.0.0-${commit_hash:0:7}"
else
  chart_version="${go_module_version}"
fi

export ref commit_hash go_module_version chart_version
echo "Resolved KAI ${ref} to ${commit_hash} (Go ${go_module_version}, chart ${chart_version})"

while IFS= read -r gomod; do
  module_dir="$(dirname "${gomod}")"
  echo "Updating ${gomod}"
  (cd "${module_dir}" && go get "${module}@${go_module_version}" && go mod tidy)
done < <(grep -rl --include=go.mod "${module}" "${repo_root}")

yq -i '(.dependencies[] | select(.name == "kai-scheduler").version) = strenv(chart_version)' \
  "${repo_root}/runai-cluster/cluster/runai-chart/Chart.yaml"
yq -i '.kai.human-readable-version = strenv(ref) |
  .kai.commit-hash = strenv(commit_hash) |
  .kai.go-module-version = strenv(go_module_version) |
  .kai.chart-version = strenv(chart_version)' "${script_dir}/kai-version.yaml"

(cd "${repo_root}/runai-cluster/cluster/runai-chart" && helm dependency update)
echo "Updated KAI to ${ref} (${commit_hash})"
