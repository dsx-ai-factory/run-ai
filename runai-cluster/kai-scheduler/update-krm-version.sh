#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


set -euo pipefail

# KRM owns the KAI chart pin, so the scheduler version is no longer ours to
# choose: it is whatever the KRM release bundles. This repo still compiles
# against KAI's Go module, though, so the two must move together or our types
# drift ahead of the scheduler that is actually deployed. Resolving the Go
# version from the KRM release is what keeps them in lockstep.
#
# The chart itself is not touched: runai-chart declares KRM as a version range,
# so there is no pin to move. The argument names the KRM release whose KAI the
# Go modules should target, which is normally the newest one the range covers.
krm_version="${1:?usage: $0 <KRM release whose KAI to target, for example v0.1.0>}"
krm_chart="oci://ghcr.io/kai-scheduler/kai-resource-management/kai-resource-management"
module="github.com/kai-scheduler/KAI-scheduler"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/../.." && pwd)"

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

echo "KRM ${krm_version} bundles KAI ${kai_version} (${commit_hash})"

while IFS= read -r gomod; do
  module_dir="$(dirname "${gomod}")"
  echo "Updating ${gomod}"
  (cd "${module_dir}" && go get "${module}@${kai_version}" && go mod tidy)
done < <(grep -rl --include=go.mod "${module}" "${repo_root}")

echo "Bumped the KAI Go module to ${kai_version} (bundled by KRM ${krm_version})"
