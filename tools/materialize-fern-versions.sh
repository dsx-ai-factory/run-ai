#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
#
# Materialize the docs of every superseded-but-supported version from its
# release branch, so a single Fern build on main serves all of them.
#
# main's fern/products/*.yml is the live (newest) version. When a version is
# superseded, fern/docs.yml repoints it at
#
#     versions/v<X.Y>/fern/products/<product>.yml
#
# and this script fills fern/versions/v<X.Y>/ from the tip of release-<X.Y>,
# mirroring the repo layout (docs/, fern/products/, local-dev.md,
# CONTRIBUTING.md). Because the layout is mirrored, the `../../docs/...` paths
# in the release branch's product files resolve unchanged: no path rewriting.
#
# The versions to build are read from fern/docs.yml itself, so registering a
# version is the docs.yml edit and nothing else. See fern/VERSIONING.md.
#
# fern/versions/ is generated and gitignored. Nothing is committed back: every
# build reads the current branch tip, so a docs fix merged to release-<X.Y>
# reaches the site on the next publish with no sync PR in between.
#
# Usage:
#   tools/materialize-fern-versions.sh [--fetch] [--remote NAME] [FERN_DIR]
#
#   --fetch        fetch each release branch from the remote first (CI, or to
#                  pick up branch updates locally). Shallow and blob-less, so
#                  it stays cheap on this repo's large history.
#   --remote NAME  remote to read release branches from (default: origin)
#   FERN_DIR       Fern workspace (default: ./fern)

set -euo pipefail

fetch=false
remote=origin
fern_dir=fern

while [ $# -gt 0 ]; do
  case "$1" in
    --fetch) fetch=true ;;
    --remote) remote="$2"; shift ;;
    -h|--help) sed -n '2,/^$/s/^# \{0,1\}//p' "$0"; exit 0 ;;
    -*) echo "error: unknown option: $1" >&2; exit 2 ;;
    *) fern_dir="$1" ;;
  esac
  shift
done

docs_yml="$fern_dir/docs.yml"
if [ ! -f "$docs_yml" ]; then
  echo "error: no such file: $docs_yml" >&2
  exit 2
fi

# Paths in the release branch that the product files reference. Anything else
# a release product file points at is reported by tools/check-fern-nav.py.
sources=(/docs/ /fern/products/ /local-dev.md /CONTRIBUTING.md)

versions=$(grep -v '^[[:space:]]*#' "$docs_yml" \
  | grep -oE 'path:[[:space:]]*versions/v[0-9]+\.[0-9]+/' \
  | grep -oE '[0-9]+\.[0-9]+' | sort -uV || true)

if [ -z "$versions" ]; then
  echo "no versioned entries in $docs_yml; nothing to materialize"
  exit 0
fi

worktree=""
cleanup() {
  if [ -n "$worktree" ]; then
    git worktree remove --force "$worktree" >/dev/null 2>&1 || rm -rf "$worktree"
    git worktree prune
  fi
}
trap cleanup EXIT

for version in $versions; do
  branch="release-$version"
  dest="$fern_dir/versions/v$version"

  if $fetch; then
    git fetch --quiet --no-tags --depth=1 --filter=blob:none "$remote" \
      "+refs/heads/$branch:refs/remotes/$remote/$branch"
  fi

  # Prefer the remote-tracking ref: it is what was actually merged, whereas a
  # local branch of the same name may carry unpushed or stale work.
  sha=""
  for ref in "refs/remotes/$remote/$branch" "refs/heads/$branch"; do
    if sha=$(git rev-parse --verify --quiet "$ref^{commit}"); then
      break
    fi
  done
  if [ -z "$sha" ]; then
    echo "error: v$version is registered in $docs_yml but $branch was not found." >&2
    echo "       Re-run with --fetch, or check that the branch exists on $remote." >&2
    exit 1
  fi

  # A sparse, detached worktree rather than `git archive`: in a blob-less
  # clone, checkout fetches the missing blobs in one batch, while archive
  # would fetch them one at a time. Sparse settings are passed with -c and the
  # pattern file is per-worktree, so the caller's repo config is untouched.
  worktree=$(mktemp -d "${TMPDIR:-/tmp}/fern-v$version.XXXXXX")
  git worktree add --quiet --detach --no-checkout "$worktree" "$sha"
  sparse_file=$(git -C "$worktree" rev-parse --path-format=absolute --git-path info/sparse-checkout)
  mkdir -p "$(dirname "$sparse_file")"
  printf '%s\n' "${sources[@]}" > "$sparse_file"
  git -C "$worktree" -c core.sparseCheckout=true -c core.sparseCheckoutCone=false \
    checkout --quiet --force --detach "$sha"

  if [ ! -d "$worktree/docs" ] || [ ! -d "$worktree/fern/products" ]; then
    echo "error: $branch ($sha) has no docs/ or fern/products/." >&2
    exit 1
  fi

  rm -rf "$dest"
  mkdir -p "$dest/fern"
  cp -R "$worktree/docs" "$dest/docs"
  cp -R "$worktree/fern/products" "$dest/fern/products"
  for file in local-dev.md CONTRIBUTING.md; do
    if [ -f "$worktree/$file" ]; then
      cp "$worktree/$file" "$dest/$file"
    fi
  done
  echo "$sha" > "$dest/.source-sha"

  cleanup
  worktree=""
  echo "v$version: materialized $branch @ ${sha:0:12} -> $dest"
done
