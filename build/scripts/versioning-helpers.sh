#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Shell helpers sourced by build/makefile/versioning.mk recipes and by
# check-breaking-changes.sh. Required env: TAG_FORMAT. Optional: MAX_PATCH
# (defaults to 100).

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

VERSIONING_SCRIPTS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

versioning_max_patch() {
    echo "${MAX_PATCH:-100}"
}

# apidiff only catches signature-level breaks, not behavior changes — confirm with a human before tagging.
versioning_confirm() {
    local prompt="$1" reply
    read -r -p "${prompt} [y/N] " reply
    case "$reply" in
        y|Y) return 0 ;;
        *) return 1 ;;
    esac
}

versioning_refresh_tags() {
    local tag
    git tag -l "${TAG_FORMAT}/v*" | while IFS= read -r tag; do
        [ -n "$tag" ] || continue
        git tag -d "$tag" >/dev/null
    done
    versioning_fetch_tags
}

# Read-only fetch for callers (e.g. check-breaking-changes.sh) that must not
# delete the caller's local tags just to preview a diff.
versioning_fetch_tags() {
    git fetch --tags --force origin >/dev/null 2>&1
}

versioning_latest_prod_tag() {
    git tag -l "${TAG_FORMAT}/v*" | sed '/-dev/d' | sort -V | tail -1
}

versioning_latest_dev_tag() {
    git tag -l "${TAG_FORMAT}/v*-dev*" | sort -V | tail -1
}

versioning_tag_to_version() {
    local tag="$1"
    [ -n "$tag" ] || { echo ""; return 0; }
    echo "$tag" | sed -E "s#^${TAG_FORMAT}/v([0-9]+\\.[0-9]+\\.[0-9]+).*#\\1#"
}

versioning_version_gt() {
    local a="$1" b="$2"
    local a_major a_minor a_patch b_major b_minor b_patch
    IFS=. read -r a_major a_minor a_patch <<<"$a"
    IFS=. read -r b_major b_minor b_patch <<<"$b"
    a_major=${a_major:-0}; a_minor=${a_minor:-0}; a_patch=${a_patch:-0}
    b_major=${b_major:-0}; b_minor=${b_minor:-0}; b_patch=${b_patch:-0}
    if [ "$a_major" -gt "$b_major" ]; then return 0; fi
    if [ "$a_major" -lt "$b_major" ]; then return 1; fi
    if [ "$a_minor" -gt "$b_minor" ]; then return 0; fi
    if [ "$a_minor" -lt "$b_minor" ]; then return 1; fi
    [ "$a_patch" -gt "$b_patch" ]
}

versioning_max_version() {
    local a="${1:-0.0.0}" b="${2:-0.0.0}"
    if versioning_version_gt "$a" "$b"; then
        echo "$a"
    else
        echo "$b"
    fi
}

versioning_next_patch() {
    local version="$1"
    local major minor patch
    IFS=. read -r major minor patch <<<"$version"
    major=${major:-0}; minor=${minor:-0}; patch=${patch:-0}
    patch=$((patch + 1))
    if [ "$patch" -ge "$(versioning_max_patch)" ]; then
        patch=0
        minor=$((minor + 1))
    fi
    echo "${major}.${minor}.${patch}"
}

versioning_next_minor() {
    local version="$1"
    local major minor patch
    IFS=. read -r major minor patch <<<"$version"
    major=${major:-0}; minor=${minor:-0}
    echo "${major}.$((minor + 1)).0"
}

versioning_latest_dev_for_version() {
    local version="$1"
    git tag -l "${TAG_FORMAT}/v${version}-dev.*" | sort -V | tail -1
}

versioning_next_dev_tag() {
    local version="$1"
    local latest
    latest=$(versioning_latest_dev_for_version "$version")
    if [ -z "$latest" ]; then
        echo "${TAG_FORMAT}/v${version}-dev.0"
        return 0
    fi
    local dev_ver
    dev_ver=$(echo "$latest" | sed -E 's#.*-dev\.([0-9]+)$#\1#')
    echo "${TAG_FORMAT}/v${version}-dev.$((dev_ver + 1))"
}

# Single place that actually creates/pushes a tag — confirms first, then
# git tag + git push + echo. Used by both the prod and dev bump paths.
versioning_create_tag() {
    local new_tag="$1" tag_message="$2"
    if ! versioning_confirm "Create tag ${new_tag}?"; then
        echo "Aborted — no tag created."
        return 0
    fi
    git tag "$new_tag" -m "$tag_message"
    git push origin "$new_tag" >/dev/null 2>&1
    printf "%bPublished tag: %s%b\n" "$GREEN" "$new_tag" "$NC"
}

# Maps a check-breaking-changes.sh exit code to "patch"/"minor" on stdout (for
# `bump_kind=$(...)` capture) plus a human-readable reason on stderr. The
# script itself must be run separately, uncaptured — its own report is meant
# to print live, not get swallowed into this capture. Used by both the prod
# and dev bump paths.
versioning_bump_kind_from_result() {
    local result="$1"
    case "$result" in
        0) printf "%bNo breaking API changes: patch.%b\n" "$BLUE" "$NC" >&2; echo patch ;;
        1) printf "%bBreaking API changes detected: minor.%b\n" "$YELLOW" "$NC" >&2; echo minor ;;
        *) return "$result" ;;
    esac
}

# Next semver for a patch/minor bump off the given prod tag (falls back to
# 0.0.0 when there is none yet). Used by both the prod and dev bump paths.
versioning_next_target_version() {
    local bump_kind="$1" latest_prod_tag="$2"
    local prod_version
    prod_version=$(versioning_tag_to_version "$latest_prod_tag")
    [ -n "$prod_version" ] || prod_version="0.0.0"
    case "$bump_kind" in
        patch) versioning_next_patch "$prod_version" ;;
        minor) versioning_next_minor "$prod_version" ;;
        *) echo "Error: unknown bump kind '${bump_kind}' (expected patch or minor)" >&2; return 2 ;;
    esac
}

# Target version for a dev bump. Continues the dev sequence when a dev tag is
# already ahead of (or equal to) the prod-derived candidate; otherwise starts a
# fresh dev.0 at the candidate. When neither a prod nor a dev tag exists yet,
# the very first dev release is 0.0.0-dev.0 (no bump applied — there is nothing
# to bump from). Used only by the dev bump path.
versioning_next_dev_target_version() {
    local bump_kind="$1" latest_prod_tag="$2" latest_dev_tag="$3"
    if [ -z "$latest_prod_tag" ] && [ -z "$latest_dev_tag" ]; then
        echo "0.0.0"
        return 0
    fi
    local dev_version candidate_version
    dev_version=$(versioning_tag_to_version "$latest_dev_tag")
    candidate_version=$(versioning_next_target_version "$bump_kind" "$latest_prod_tag")
    versioning_max_version "$dev_version" "$candidate_version"
}

versioning_publish_prod_bump() {
    local bump_kind="$1" latest_prod_tag="$2"
    local target_version new_tag message
    target_version=$(versioning_next_target_version "$bump_kind" "$latest_prod_tag")
    case "$bump_kind" in
        patch) message="patch" ;;
        minor) message="minor — breaking changes" ;;
    esac
    new_tag="${TAG_FORMAT}/v${target_version}"
    versioning_create_tag "$new_tag" "Prod release ${new_tag} (${message})"
}

versioning_publish_dev_bump() {
    local bump_kind="$1" latest_prod_tag="$2"
    local latest_dev_tag target_version new_tag
    latest_dev_tag=$(versioning_latest_dev_tag)
    echo "Latest dev tag: ${latest_dev_tag}"
    target_version=$(versioning_next_dev_target_version "$bump_kind" "$latest_prod_tag" "$latest_dev_tag")
    new_tag=$(versioning_next_dev_tag "$target_version")
    versioning_create_tag "$new_tag" "Dev release ${new_tag}"
}

versioning_bump() {
    local channel="$1" pkg_dir="$2"
    local latest_prod_tag result bump_kind
    versioning_refresh_tags
    latest_prod_tag=$(versioning_latest_prod_tag)
    echo "Latest prod tag: ${latest_prod_tag}"
    result=0
    "${VERSIONING_SCRIPTS_DIR}/check-breaking-changes.sh" "$TAG_FORMAT" "$pkg_dir" "$latest_prod_tag" || result=$?
    bump_kind=$(versioning_bump_kind_from_result "$result")
    case "$channel" in
        prod) versioning_publish_prod_bump "$bump_kind" "$latest_prod_tag" ;;
        dev) versioning_publish_dev_bump "$bump_kind" "$latest_prod_tag" ;;
        *) echo "Error: unknown channel '${channel}' (expected prod or dev)" >&2; return 2 ;;
    esac
}
