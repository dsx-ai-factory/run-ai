#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Renovate post-upgrade hook.
#
# A folder often holds several Go modules with a local replace between them --
# a service and its pkg/client, a package and its integration-tests. Renovate
# only updates the module that had a dependency change, so its sibling's go.mod
# goes stale and every command in that directory fails with
# "go: updates to go.mod needed".
#
# Tidy every module in the group, and regenerate mocks where the module has a
# mockery config: interfaces such as jetstream.JetStream gain methods between
# releases, and a checked-in mock that no longer satisfies the interface breaks
# compilation of the whole test package.
set -euo pipefail

root="${1:?usage: $0 <group-root-dir>}"

if [ ! -d "$root" ]; then
  echo "renovate-tidy-group: no such directory: $root" >&2
  exit 1
fi

# Materialise the list first: piping find straight into the loop hides a find
# failure, and an empty result would otherwise report success having done
# nothing -- the failure mode this hook exists to prevent.
modules_file=$(mktemp)
trap 'rm -f "$modules_file"' EXIT
if ! find "$root" -name go.mod -not -path '*/vendor/*' -print0 >"$modules_file"; then
  echo "renovate-tidy-group: failed to enumerate modules under $root" >&2
  exit 1
fi
mapfile -d '' -t gomods <"$modules_file"

if [ "${#gomods[@]}" -eq 0 ]; then
  echo "renovate-tidy-group: no go.mod found under $root" >&2
  exit 1
fi

for gomod in "${gomods[@]}"; do
  dir=$(dirname "$gomod")

  echo "renovate-tidy-group: tidying $dir"
  go -C "$dir" mod tidy

  # Go through the module's own make target rather than a globally installed
  # mockery. Modules pin MOCKERY_VERSION themselves (v2.53.3 today) while
  # build/makefile/golang.mk pins v3, and v3 emits a different file format --
  # using the wrong one would rewrite every mock instead of adding the missing
  # methods. Skip modules that have a config but no target for it.
  if [ -f "$dir/.mockery.yaml" ] || [ -f "$dir/.mockery.yml" ]; then
    for target in mockery generate_mocks; do
      if [ -f "$dir/Makefile" ] && grep -qE "^${target}:" "$dir/Makefile"; then
        echo "renovate-tidy-group: regenerating mocks in $dir (make $target)"
        # Non-fatal. Modules are processed one at a time, so aborting here would
        # skip `go mod tidy` for every module after this one -- reintroducing the
        # stale-sibling failure this hook exists to prevent. Mock generation also
        # reaches the public module proxy to install mockery, which tidy does not,
        # so it has failure modes of its own. A stale mock instead surfaces in CI
        # as a readable "does not implement" compile error.
        if ! make -C "$dir" "$target"; then
          echo "renovate-tidy-group: mock regeneration failed in $dir; continuing" >&2
        fi
        break
      fi
    done
  fi
done

echo "renovate-tidy-group: processed ${#gomods[@]} module(s) under $root"
