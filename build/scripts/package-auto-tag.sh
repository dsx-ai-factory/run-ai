#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Tags every shared package that has changed since its latest prod tag, through the
# same `make package-bump-version` a developer runs. Driven by
# .github/workflows/package-auto-tag.yml; runnable by hand from an up-to-date main.
#
# Env:
#   PACKAGES    comma-separated package directories to consider instead of all of them
#   PATCH_ONLY  passed through: never prompt, never tag a breaking change (see
#               versioning-helpers.sh). The workflow always sets it.
#   DRY_RUN     passed through: report the tags, create none.
#
# Exit status is non-zero when any package needs a person: a breaking change that
# PATCH_ONLY refused, or an error. Packages skipped because main moved during the run
# are left to the next run, which is already queued behind this one.

set -uo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT"

SUMMARY=${GITHUB_STEP_SUMMARY:-/dev/stdout}

# One level deep only: a pathspec `*` crosses slashes, and the nested modules below a
# package (queryapi/integration-tests, runaidb/integration-tests) are its test
# harnesses, not packages anyone imports.
all_packages() {
    git ls-files 'runai-common-packages/*/go.mod' 'runai-application-packages/*/go.mod' \
        | grep -E '^runai-(common|application)-packages/[^/]+/go\.mod$' \
        | xargs -r -n1 dirname | sort
}

latest_prod_tag() {
    git tag -l "$1/v*" | grep -v -- '-dev' | sort -V | tail -1
}

# A package is pending when something under it changed after its latest prod tag.
# Asked of the repository, not of the push that triggered the run: GitHub keeps only
# one queued run per concurrency group and drops the rest, so a push-based list would
# silently lose packages.
pending() {
    local pkg="$1" tag
    tag=$(latest_prod_tag "$pkg")
    [ -n "$tag" ] || return 1
    [ -n "$(git log --format=%H -1 "${tag}..HEAD" -- "$pkg")" ]
}

git fetch --tags --force origin >/dev/null 2>&1

if [ -n "${PACKAGES:-}" ]; then
    IFS=',' read -ra candidates <<<"$PACKAGES"
else
    mapfile -t candidates < <(all_packages)
fi

tagged=() refused=() moved=() failed=() untagged=() unchanged=()
for raw in "${candidates[@]}"; do
    pkg=$(echo "$raw" | xargs)
    pkg=${pkg%/}
    [ -n "$pkg" ] || continue
    if [ ! -f "$pkg/go.mod" ] || [ ! -f "$pkg/Makefile" ]; then
        failed+=("$pkg (not a package directory)")
        continue
    fi
    if [ -z "$(latest_prod_tag "$pkg")" ]; then
        # The first version of a package is a decision, not a patch.
        untagged+=("$pkg")
        continue
    fi
    if [ -z "${PACKAGES:-}" ] && ! pending "$pkg"; then
        unchanged+=("$pkg")
        continue
    fi

    echo "::group::${pkg}"
    output=$(make -C "$pkg" package-bump-version 2>&1)
    status=$?
    echo "$output"
    echo "::endgroup::"

    tag=$(echo "$output" | sed 's/\x1b\[[0-9;]*m//g' \
        | grep -oE '(Would publish tag|Published tag): \S+' | awk '{print $NF}' | tail -1)
    case "$status" in
        0) tagged+=("$pkg -> ${tag:-?}") ;;
        # versioning.mk's main-sync check: main moved under this run. The queued run
        # picks the package up from the new tip.
        *) if echo "$output" | grep -q "not up to date with origin/main\|must be on branch 'main'"; then
               moved+=("$pkg")
           elif echo "$output" | grep -q "PATCH_ONLY is set"; then
               refused+=("$pkg")
           else
               failed+=("$pkg (exit ${status})")
           fi ;;
    esac
done

{
    echo "## Package auto-tag${DRY_RUN:+ (dry run — nothing tagged)}"
    echo
    section() {
        local title="$1"; shift
        [ "$#" -gt 0 ] || return 0
        echo "**${title}**"
        printf -- '- %s\n' "$@"
        echo
    }
    section "Tagged" "${tagged[@]}"
    section "Breaking API change — not tagged, needs a person" "${refused[@]}"
    section "Failed" "${failed[@]}"
    section "Skipped: main moved during the run (next run takes them)" "${moved[@]}"
    section "Skipped: no prod tag yet (first version is a decision)" "${untagged[@]}"
    if [ "${#unchanged[@]}" -gt 0 ]; then
        echo "Unchanged since their latest tag: ${#unchanged[@]} package(s)."
    fi
} >>"$SUMMARY"

[ "${#refused[@]}" -eq 0 ] && [ "${#failed[@]}" -eq 0 ]
