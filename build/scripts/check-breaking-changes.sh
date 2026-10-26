#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# check-breaking-changes.sh
# Compares the current package against its latest prod tag using apidiff.
# Exit 0: no breaking changes — the bump targets select a patch bump
# Exit 1: breaking changes found — the bump targets select a minor bump
#
# Usage: check-breaking-changes.sh <TAG_FORMAT> <PKG_DIR> [LATEST_TAG]
#   TAG_FORMAT  e.g. runai-common-packages/runaihttp
#   PKG_DIR     absolute path to the package directory ($(CURDIR))
#   LATEST_TAG  optional — skip the lookup when the caller already resolved it; looked up here when omitted.

set -euo pipefail

export TAG_FORMAT="${1:?TAG_FORMAT required (e.g. runai-common-packages/runaihttp)}"
PKG_DIR="${2:?PKG_DIR required (absolute path)}"
LATEST_TAG="${3:-}"

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=./versioning-helpers.sh
. "${SCRIPT_DIR}/versioning-helpers.sh"

# Generated mockery packages are test scaffolding rather than a contract for
# consumers. Regenerating them (e.g. the mockery v2 -> v3 migration) rewrites
# every exported mock symbol, which would otherwise dominate the report.
IGNORED_CHANGES_REGEX='^- (\./([^ ]*/)?mocks\.|package [^ ]*/mocks:)'

# ── helpers ──────────────────────────────────────────────────────────────────

die() { echo -e "${RED}Error: $*${NC}" >&2; exit 2; }

# Drop changes for ignored packages. grep exits 1 when nothing is left, which is
# a valid empty result rather than a failure.
filter_ignored() {
    grep -v -E "$IGNORED_CHANGES_REGEX" || true
}

# Extract one section of an apidiff report. Report.Text writes the incompatible
# changes under an "Incompatible changes:" header, then the compatible ones
# under "Compatible changes:".
report_section() {
    local section="$1"
    awk -v want="$section" '
        /^(Incompatible|Compatible) changes:$/ { current = $1; next }
        current == want && NF
    '
}

ensure_apidiff() {
    local install=false
    local current_go_series
    local apidiff_go_series=""

    current_go_series=$(go env GOVERSION | cut -d. -f1,2)
    if command -v apidiff &>/dev/null; then
        apidiff_go_series=$(go version -m "$(command -v apidiff)" 2>/dev/null \
            | awk 'NR == 1 { print $NF }' \
            | cut -d. -f1,2)
    fi

    # apidiff uses go/types and must be built with the package's Go toolchain
    # series; an older binary cannot read export data from newer Go versions.
    if ! command -v apidiff &>/dev/null || [ "$apidiff_go_series" != "$current_go_series" ]; then
        install=true
    fi

    if [ "$install" = "true" ]; then
        echo "Installing apidiff for ${current_go_series} ..."
        GOTOOLCHAIN="$(go env GOVERSION)" \
            go install golang.org/x/exp/cmd/apidiff@latest \
            || die "Failed to install apidiff. Make sure Go is installed and GOPATH/bin is on PATH."
    fi
}

# Pure-bash relative path (avoids GNU realpath vs macOS realpath differences)
rel_path() {
    local target="$1" base="$2"
    echo "${target#"${base}"/}"
}

# ── main ─────────────────────────────────────────────────────────────────────

cd "$PKG_DIR" || die "Package directory not found: ${PKG_DIR}"
ensure_apidiff

REPO_ROOT=$(git -C "$PKG_DIR" rev-parse --show-toplevel)
PKG_REL=$(rel_path "$PKG_DIR" "$REPO_ROOT")

if [ -z "$LATEST_TAG" ]; then
    versioning_fetch_tags
    LATEST_TAG=$(cd "$REPO_ROOT" && versioning_latest_prod_tag)
fi

if [ -z "$LATEST_TAG" ]; then
    echo -e "${YELLOW}No previous prod release found for ${TAG_FORMAT}.${NC}"
    exit 0
fi

echo "Comparing current code against: ${LATEST_TAG}"

# Extract only this package from the old tag. Avoid checking out the entire
# monorepo: unrelated paths must not affect a package compatibility check.
OLD_TREE_DIR=$(mktemp -d)
cleanup() {
    rm -rf "$OLD_TREE_DIR"
}
trap cleanup EXIT

if ! git -C "$REPO_ROOT" cat-file -e "${LATEST_TAG}:${PKG_REL}" 2>/dev/null; then
    echo -e "${YELLOW}Package '${PKG_REL}' not found in ${LATEST_TAG} — treating as new package.${NC}"
    exit 0
fi

git -C "$REPO_ROOT" archive "$LATEST_TAG" "$PKG_REL" \
    | tar -x -C "$OLD_TREE_DIR" \
    || die "Failed to extract ${PKG_REL} from ${LATEST_TAG}."

OLD_PKG_DIR="${OLD_TREE_DIR}/${PKG_REL}"

# ── module API diff ───────────────────────────────────────────────────────────

# apidiff accepts import paths or export-data files, not source directories.
# Generate module export data from each package tree independently, then compare
# the snapshots. Module mode covers all public sub-packages and detects package
# deletion while filtering Go internal packages by default.
OLD_API="${OLD_TREE_DIR}/old.api"
NEW_API="${OLD_TREE_DIR}/new.api"
APIDIFF_STDERR="${OLD_TREE_DIR}/apidiff.stderr"
API_GOCACHE="${OLD_TREE_DIR}/go-build-cache"

OLD_MODULE_PATH=$(cd "$OLD_PKG_DIR" && GOWORK=off go list -m -f '{{.Path}}') \
    || die "Failed to read the old module path from ${LATEST_TAG}."
NEW_MODULE_PATH=$(cd "$PKG_DIR" && GOWORK=off go list -m -f '{{.Path}}') \
    || die "Failed to read the current module path from ${PKG_REL}."

rm -f "${OLD_PKG_DIR}/go.sum"
(cd "$OLD_PKG_DIR" && GOWORK=off GOFLAGS=-mod=mod GONOSUMDB="${GONOSUMDB:+${GONOSUMDB},}github.com/dsx-ai-factory/run-ai" GOCACHE="$API_GOCACHE" apidiff -m -w "$OLD_API" "$OLD_MODULE_PATH" 2>"$APIDIFF_STDERR") \
    || die "Failed to export the API from ${LATEST_TAG}: $(cat "$APIDIFF_STDERR")"
(cd "$PKG_DIR" && GOWORK=off GOCACHE="$API_GOCACHE" apidiff -m -w "$NEW_API" "$NEW_MODULE_PATH" 2>"$APIDIFF_STDERR") \
    || die "Failed to export the current API: $(cat "$APIDIFF_STDERR")"

result=$(apidiff -m "$OLD_API" "$NEW_API" 2>"$APIDIFF_STDERR") \
    || die "apidiff failed: $(cat "$APIDIFF_STDERR")"

BREAKING_CHANGES=$(printf '%s\n' "$result" | report_section Incompatible | filter_ignored)
COMPAT_CHANGES=$(printf '%s\n' "$result" | report_section Compatible | filter_ignored)

BREAKING=false
[ -n "$BREAKING_CHANGES" ] && BREAKING=true

# ── report ───────────────────────────────────────────────────────────────────

echo ""

if [ -n "$COMPAT_CHANGES" ]; then
    echo -e "${BLUE}Backward-compatible changes:${NC}"
    echo -e "$COMPAT_CHANGES"
fi

if [ "$BREAKING" = "true" ]; then
    echo -e "${BLUE}Breaking (incompatible) changes:${NC}"
    echo -e "$BREAKING_CHANGES"
    echo -e "${YELLOW}These changes break the public API of ${TAG_FORMAT}.${NC}"
    exit 1
else
    [ -z "$COMPAT_CHANGES" ] && echo -e "${GREEN}No API changes detected vs ${LATEST_TAG}.${NC}"
    [ -n "$COMPAT_CHANGES" ] && echo -e "${GREEN}Backward-compatible API changes are treated as a patch bump.${NC}"
    exit 0
fi
