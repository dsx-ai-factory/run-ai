#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
SUPPORTED_DIRS_WITH_INTEGRATION_TESTS=$(cat runai-cluster/cluster/integration-tests/supported-dirs.txt)
INFRA_DIRS=$(cat runai-cluster/cluster/integration-tests/infra-dirs.txt)
if [[ "$RUN_ALL_TESTS" == "true" ]]; then
  echo "RUN_ALL_TESTS is set to true, running all integration tests."
  CHANGED_DIRS=$SUPPORTED_DIRS_WITH_INTEGRATION_TESTS
  echo CHANGED_DIRS=$CHANGED_DIRS >> $GITHUB_OUTPUT
  exit 0
fi

git fetch origin main
changed_files=$(git --no-pager diff --name-only origin/main..HEAD "$CLUSTER_PATH")

# First, we need to extract all the directories that have changed from the files
all_changed_dirs=""
for file in $changed_files; do
  dir=$(echo "$file" | sed "s|$CLUSTER_PATH/\([^/]*\)/.*|\1|")
  # If integration-tests dir have changed, we need to run all the integration tests
  if [[ "$dir" == "integration-tests" ]]; then
      echo "integration-tests dir have changed, running all integration tests."
      CHANGED_DIRS=$SUPPORTED_DIRS_WITH_INTEGRATION_TESTS
      echo CHANGED_DIRS=$CHANGED_DIRS >> $GITHUB_OUTPUT
      exit 0
  fi

  dir_already_in_list=false
  for changed_dir in $all_changed_dirs; do
    if [[ "$dir" == "$changed_dir" ]]; then
      dir_already_in_list=true
      break
    fi
  done
  if $dir_already_in_list; then
    continue
  fi
  if [[ -z "$all_changed_dirs" ]]; then
    all_changed_dirs="$dir"
  else
    all_changed_dirs="$all_changed_dirs $dir"
  fi
done

# If an infra dir has changed, we need to run all the integration tests
infra_dirs=""
for dir in $all_changed_dirs; do
  for infra_dir in $infra_dirs; do
    if [[ "$dir" == "$infra_dir" ]]; then
      if [[ -z "$infra_dirs" ]]; then
        infra_dirs="$dir"
      else
        infra_dirs="$infra_dirs $dir"
      fi
      break
    fi
  done
done

if [[ -n "$infra_dirs" ]]; then
  echo "Found directories that if changed, we need to run all the integration tests: $infra_dirs"
  CHANGED_DIRS=$SUPPORTED_DIRS_WITH_INTEGRATION_TESTS
  echo CHANGED_DIRS=$CHANGED_DIRS >> $GITHUB_OUTPUT
  exit 0
fi

# Now, extract only the directories that have support in integration tests
supported_changed_dirs=""
for dir in $all_changed_dirs; do
  for supported_dir in $SUPPORTED_DIRS_WITH_INTEGRATION_TESTS; do
    if [[ "$dir" == "$supported_dir" ]]; then
      if [[ -z "$supported_changed_dirs" ]]; then
        supported_changed_dirs="$dir"
      else
        supported_changed_dirs="$supported_changed_dirs $dir"
      fi
      break
    fi
  done
done

if [[ -z "$supported_changed_dirs" ]]; then
  echo "::notice:: No supported directories have changed."
  echo CHANGED_DIRS=$supported_changed_dirs >> $GITHUB_OUTPUT
  exit 0
fi
CHANGED_DIRS=$supported_changed_dirs
echo "Found supported directories with changes: $CHANGED_DIRS"
echo CHANGED_DIRS=$CHANGED_DIRS >> $GITHUB_OUTPUT