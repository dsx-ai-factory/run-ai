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

fail() {
  echo "$*" >&2
  exit 1
}

if [[ ! "${krm_version}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  fail "KRM version must look like v0.1.0, got ${krm_version}"
fi
for tool in go helm yq; do
  command -v "${tool}" >/dev/null || fail "missing required tool: ${tool}"
done
# The Python yq wrapper installs under the same name but rejects the syntax below.
yq --version 2>/dev/null | grep -q mikefarah || fail "yq must be mikefarah/yq v4 (brew install yq)"

# Nothing is edited until both artifacts resolve, so an unpublished release
# fails cleanly instead of leaving half the modules bumped.
krm_metadata="$(helm show chart "${krm_chart}" --version "${krm_version}")" ||
  fail "KRM chart ${krm_version} is not published to ${krm_chart} yet"
kai_version="$(yq '.dependencies[] | select(.name == "kai-scheduler").version' - <<<"${krm_metadata}")"

if [[ -z "${kai_version}" || "${kai_version}" == "null" ]]; then
  fail "KRM ${krm_version} declares no kai-scheduler dependency"
fi

tmp_mod_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_mod_dir}"' EXIT
commit_hash="$(
  cd "${tmp_mod_dir}"
  go mod init kai-version-resolver >/dev/null 2>&1
  go list -m -f '{{.Origin.Hash}}' "${module}@${kai_version}"
)" || fail "KAI Go module ${kai_version} is not resolvable yet"

echo "KRM ${krm_version} bundles KAI ${kai_version} (${commit_hash})"

# Prints the absolute directory of every local replace target in a go.mod.
local_replace_targets() {
  local dir rel
  dir="$(dirname "$1")"
  sed -nE 's#.*=>[[:space:]]+(\.\.?/[^[:space:]]+).*#\1#p' "$1" | while IFS= read -r rel; do
    (cd "${dir}/${rel}" 2>/dev/null && pwd) || true
  done
}

direct="$(grep -rl --include=go.mod "${module}" "${repo_root}" | xargs -n1 dirname | sort -u)"

# A module that reaches KAI only through a local replace never names it, yet
# inherits the bumped requirements and fails "go: updates to go.mod needed"
# unless it is tidied too. Repeat until no new module joins, since these chain.
dependents=""
while :; do
  added=false
  while IFS= read -r gomod; do
    gomod="${repo_root}/${gomod#./}"
    dir="$(dirname "${gomod}")"
    grep -qxF "${dir}" <<<"${direct}"$'\n'"${dependents}" && continue
    while IFS= read -r target; do
      if grep -qxF "${target}" <<<"${direct}"$'\n'"${dependents}"; then
        dependents+="${dir}"$'\n'
        added=true
        break
      fi
    done < <(local_replace_targets "${gomod}")
  # Relative paths, so a hidden directory above the checkout doesn't exclude everything.
  done < <(cd "${repo_root}" && find . -name go.mod -not -path '*/.*' -not -path '*/vendor/*')
  [[ "${added}" == true ]] || break
done

while IFS= read -r module_dir; do
  echo "Updating ${module_dir}"
  (cd "${module_dir}" && go get "${module}@${kai_version}" && go mod tidy)
done <<<"${direct}"

while IFS= read -r module_dir; do
  [[ -n "${module_dir}" ]] || continue
  echo "Tidying ${module_dir} (reaches KAI through a local replace)"
  (cd "${module_dir}" && go mod tidy)
done <<<"${dependents}"

# A tidy can leave an earlier module stale when a later one moves a shared
# requirement, so confirm the end state rather than trusting the order above.
untidy=""
while IFS= read -r module_dir; do
  [[ -n "${module_dir}" ]] || continue
  (cd "${module_dir}" && go mod tidy -diff >/dev/null 2>&1) || untidy+="  ${module_dir}"$'\n'
done <<<"${direct}"$'\n'"${dependents}"
[[ -z "${untidy}" ]] || fail "go.mod is not tidy after the update in:"$'\n'"${untidy}"

echo "Bumped the KAI Go module to ${kai_version} (bundled by KRM ${krm_version})"
