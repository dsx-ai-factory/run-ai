#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


set -euo pipefail

# KRM owns the KAI chart pin, so the scheduler version is no longer ours to
# choose: it is whatever the KRM release bundles. This repo still compiles
# against KAI's Go module, though, so the two must move together or our types
# drift ahead of the scheduler that is actually deployed. Resolving the Go
# version from the KRM release is what keeps them in lockstep.
krm_version="${1:?usage: $0 <KRM chart version, for example v0.1.0>}"
krm_chart="oci://ghcr.io/kai-scheduler/kai-resource-management/kai-resource-management"
module="github.com/kai-scheduler/KAI-scheduler"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/../.." && pwd)"
chart_dir="${repo_root}/runai-cluster/cluster/runai-chart"

krm_metadata="$(helm show chart "${krm_chart}" --version "${krm_version}")"
kai_version="$(yq '.dependencies[] | select(.name == "kai-scheduler").version' - <<<"${krm_metadata}")"

if [[ -z "${kai_version}" || "${kai_version}" == "null" ]]; then
  echo "KRM ${krm_version} declares no kai-scheduler dependency" >&2
  exit 1
fi

tmp_mod_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_mod_dir}"' EXIT
commit_hash="$(
  cd "${tmp_mod_dir}"
  go mod init kai-version-resolver >/dev/null 2>&1
  go list -m -f '{{.Origin.Hash}}' "${module}@${kai_version}"
)"

export krm_version kai_version commit_hash
echo "KRM ${krm_version} bundles KAI ${kai_version} (${commit_hash})"

while IFS= read -r gomod; do
  module_dir="$(dirname "${gomod}")"
  echo "Updating ${gomod}"
  (cd "${module_dir}" && go get "${module}@${kai_version}" && go mod tidy)
done < <(grep -rl --include=go.mod "${module}" "${repo_root}")

# yq silently matches nothing if the dependency name is wrong, so check first.
if ! yq -e '.dependencies[] | select(.name == "kai-resource-management")' \
  "${chart_dir}/Chart.yaml" >/dev/null; then
  echo "runai-chart declares no kai-resource-management dependency" >&2
  exit 1
fi
yq -i '(.dependencies[] | select(.name == "kai-resource-management").version) = strenv(krm_version)' \
  "${chart_dir}/Chart.yaml"

yq -i '.krm.chart-version = strenv(krm_version) |
  .kai.human-readable-version = strenv(kai_version) |
  .kai.commit-hash = strenv(commit_hash) |
  .kai.go-module-version = strenv(kai_version)' "${script_dir}/kai-version.yaml"

(cd "${chart_dir}" && helm dependency update)
echo "Updated KRM to ${krm_version} (KAI ${kai_version})"
