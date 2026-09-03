#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

#
# Shared checker for forbidden `replace` directives in go.mod files.
#
# The naive way to write these checks is a grep for the protected module path
# followed by `=>`, i.e. matching only the LEFT side of the directive. That is
# trivially bypassed, because Go does not care what the left side is called:
#
#   replace github.com/dsx-ai-factory/run-ai/runai-clients/some-client => ../../control-plane/foo/pkg/client
#   replace example.com/anything => github.com/dsx-ai-factory/run-ai/runai-common-packages/cache v1.2.3
#
# Both pull the protected module into the build while the left side matches
# nothing. So we parse each directive and check BOTH sides:
#
#   * left side  — the module being replaced.
#   * right side — resolved to a module path. For a filesystem target we read the
#     module line of the go.mod it points at, so the check follows what the
#     replace actually resolves to rather than how the path is spelled.
#
# Usage:
#   validate-replaces.sh <label> <target-ere> <allow-ere> <local-only|any> <go.mod>...
#
#     label       human-readable name used in the error output.
#     target-ere  ERE matching the protected module paths. Matched as a substring
#                 (deliberately): write it without the `github.com/dsx-ai-factory/run-ai/`
#                 prefix so it also matches the repo-relative directory paths used
#                 when a replacement target has no readable go.mod.
#     allow-ere   ERE of exempt module paths, matched as a WHOLE module path (the
#                 script anchors it), so an entry does not accidentally exempt
#                 <entry>-fork or any path that merely contains it. Spell entries
#                 fully qualified, e.g. github\.com/dsx-ai-factory/run-ai/<module>. Pass ""
#                 for none.
#     local-only  only treat a LEFT-side match as an error when the target is a
#                 filesystem path (a version-pin replace is fine). `any` errors
#                 on every replace of a protected module.
#
# Prints every offending directive and exits 1 if there are any.

set -euo pipefail

if [ "$#" -lt 5 ]; then
  echo "usage: $(basename "$0") <label> <target-ere> <allow-ere> <local-only|any> <go.mod>..." >&2
  exit 2
fi

LABEL="$1"
TARGET_RE="$2"
ALLOW_RE="$3"
MODE="$4"
shift 4

# Protected-module detection is a substring match: the target may be seen either
# as a module path or as a directory path.
matches_target() {
  [ -n "$1" ] && printf '%s' "$1" | grep -qE "$TARGET_RE"
}

# Exemptions are matched against the WHOLE module path. An unanchored match would
# make every allow-list entry a substring rule, exempting e.g.
# control-plane/authorization/pkg/client-fork, or any module path that happens to
# contain an allowed value.
is_allowed() {
  [ -n "$ALLOW_RE" ] && [ -n "$1" ] && printf '%s' "$1" | grep -qE "^($ALLOW_RE)\$"
}

# Parse every replace directive (both the single-line and the `replace ( ... )`
# block form) out of a go.mod, emitting `<line>\t<left module>\t<right target>`.
# The right target is the raw token as written — a module path or a filesystem
# path; resolution happens below, where the go.mod's own directory is known.
parse_replaces() {
  awk '
    { line = $0; sub(/\/\/.*/, "", line); gsub(/\r/, "", line) }
    inblock && line ~ /^[[:space:]]*\)/ { inblock = 0; next }
    !inblock {
      if (line ~ /^[[:space:]]*replace[[:space:]]*\([[:space:]]*$/) { inblock = 1; next }
      if (line !~ /^[[:space:]]*replace[[:space:]]/) next
      sub(/^[[:space:]]*replace[[:space:]]+/, "", line)
    }
    {
      if (index(line, "=>") == 0) next
      left = substr(line, 1, index(line, "=>") - 1)
      right = substr(line, index(line, "=>") + 2)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", left)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", right)
      split(left, l, /[[:space:]]+/)
      split(right, r, /[[:space:]]+/)
      if (l[1] == "" || r[1] == "") next
      print NR "\t" l[1] "\t" r[1]
    }
  ' "$1"
}

# Read the module path out of a go.mod. Go accepts the block form
#
#   module (
#       example.com/foo
#   )
#
# as well as `module example.com/foo`, so taking the second field of the `module`
# line yields "(" for a block-form file — which matches no protected pattern and
# would let a local replace pointing at it through. Paths may also be quoted.
module_path() {
  awk '
    { line = $0; sub(/\/\/.*/, "", line); gsub(/\r/, "", line); $0 = line }
    inblock {
      if ($0 ~ /^[[:space:]]*\)/) { exit }
      if (NF == 0) next
      print $1; exit
    }
    $1 == "module" {
      if (NF == 1 || $2 == "(") {
        if (NF >= 3 && $3 != ")") { print $3; exit }
        inblock = 1; next
      }
      print $2; exit
    }
  ' "$1" | tr -d '"'
}

# Turn the right-hand target into the module path it actually resolves to. For a
# filesystem path that means reading the module line of the go.mod living there;
# if that go.mod is unreachable we fall back to the canonical (or raw) path, which
# still matches the target ERE for in-repo directories.
resolve_target() {
  local base_dir="$1" target="$2" dir

  case "$target" in
    ./* | ../* | /* | . | ..) ;;
    *)
      printf '%s' "$target"
      return
      ;;
  esac

  case "$target" in
    /*) dir="$target" ;;
    *) dir="$base_dir/$target" ;;
  esac

  if [ -f "$dir/go.mod" ]; then
    local module
    module="$(module_path "$dir/go.mod")"
    if [ -n "$module" ]; then
      printf '%s' "$module"
      return
    fi
  fi

  if [ -d "$dir" ]; then
    (cd "$dir" && pwd -P)
    return
  fi

  printf '%s' "$target"
}

offending=""

for gomod in "$@"; do
  [ -f "$gomod" ] || continue
  base_dir="$(dirname "$gomod")"

  while IFS=$'\t' read -r lineno left right; do
    [ -n "${left:-}" ] || continue

    case "$right" in
      ./* | ../* | /* | . | ..) is_local=1 ;;
      *) is_local=0 ;;
    esac

    resolved="$(resolve_target "$base_dir" "$right")"

    reason=""
    if matches_target "$left" && ! is_allowed "$left" &&
      { [ "$MODE" = "any" ] || [ "$is_local" -eq 1 ]; }; then
      reason="replaces a protected module"
    elif [ "$resolved" != "$left" ] && matches_target "$resolved" && ! is_allowed "$resolved"; then
      reason="resolves to protected module '$resolved'"
    fi

    if [ -n "$reason" ]; then
      offending="${offending}  ${gomod}:${lineno}: ${left} => ${right}"$'\n'
      offending="${offending}    ^ ${reason}"$'\n'
    fi
  done < <(parse_replaces "$gomod")
done

if [ -n "$offending" ]; then
  printf 'Error: %s\n' "$LABEL" >&2
  printf '%s' "$offending" >&2
  exit 1
fi

exit 0
