#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Renovate post-upgrade hook.
#
# A prettier bump changes how prettier formats existing, untouched source.
# 3.4.2 -> 3.9.6 collapses short union types onto one line, which reformats 33
# files under control-plane/frontend. `npm run format` runs prettier --check,
# so the bump fails its own lint job:
#
#   [warn] Code style issues found in 33 files. Run Prettier with --write to fix.
#
# Reformat with the version the branch is moving to, so the PR carries the churn
# its own upgrade causes.
#
# Renovate updates npm lockfiles with `npm install --package-lock-only`, which
# never populates node_modules, so ./node_modules/.bin/prettier does not exist
# here and `npm run format:write` cannot resolve its binary. Fetch the pinned
# version through npx instead.
#
# Run from the package directory, exactly as the format script does. Prettier
# resolves .prettierignore relative to the working directory, so invoking it
# from the repo root silently ignores the file and rewrites the JSON assets and
# READMEs that `npm run format` is configured to leave alone.
set -euo pipefail

dir="${1:?usage: $0 <package-file-dir>}"

if [ ! -f "$dir/package.json" ]; then
  echo "renovate-prettier: no package.json in $dir" >&2
  exit 1
fi

cd "$dir"

# Only the version actually declared -- resolving "latest" would reformat
# against a prettier the branch is not adopting.
version=$(node -p "
  const p = require('./package.json');
  ((p.devDependencies || {}).prettier || (p.dependencies || {}).prettier || '');
")

if [ -z "$version" ]; then
  echo "renovate-prettier: $dir does not depend on prettier, nothing to do"
  exit 0
fi

if [ ! -d src ]; then
  echo "renovate-prettier: $dir has no src directory, nothing to do"
  exit 0
fi

args=(--write)
if [ -f .prettierrc.json ]; then
  args+=(--config .prettierrc.json)
fi

echo "renovate-prettier: formatting $dir/src with prettier@$version"
npx --yes "prettier@$version" "${args[@]}" ./src
