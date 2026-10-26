#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0



# Function to determine base branch
determine_base_branch() {
    # Check if the base branch is 'main', if so, find the latest version branch
    if [[ "$BASE_BRANCH_NAME" == "main" ]]; then
        echo "Base branch is 'main', looking for the latest version branch."
        
        # In PR context, we need to fetch version branches first
        echo "Fetching version branches from remote..."
        
        # Try to fetch all refs first (more comprehensive)
        if ! git fetch origin '+refs/heads/v*:refs/remotes/origin/v*'; then
            echo "Failed to fetch version branches with pattern, trying to fetch all branches..."
            # Fallback: fetch all remote references
            git fetch origin || echo "Failed to fetch from origin, continuing with available branches..."
        fi
        
        # List all branches that match the v<semver> pattern and sort them

        latest_version_branch=$(git branch -r \
            | grep -o 'origin/v[0-9]\+\.[0-9]\+' \
            | sort -V \
            | tail -n 1)
        
        if [[ -n "$latest_version_branch" ]]; then
            BASE_BRANCH_NAME=${latest_version_branch#"origin/"}
            echo "Latest version branch found: $latest_version_branch"
            echo "Using base branch: $BASE_BRANCH_NAME"
        else
            echo "No version branches found after fetching."
            echo "Available branches matching version pattern:"
            git branch -r | grep -E 'origin/v[0-9]+\.[0-9]+' || echo "No version branches available"
            echo "Using main as fallback"
            BASE_BRANCH_NAME="main"
        fi
    else
        # If base is not 'main', we just use the base branch
        echo "Using base branch: $BASE_BRANCH_NAME"
    fi
}


if [[ -z "$BASE_BRANCH_NAME" ]]; then
    echo "Base branch name is not set, skipping backwards compatibility check"
    exit 0
fi

determine_base_branch



echo "Fetching base branch: $BASE_BRANCH_NAME"
if ! git fetch origin "$BASE_BRANCH_NAME"; then
    echo "Failed to fetch branch: $BASE_BRANCH_NAME"
    exit 1
fi

mkdir -p /tmp/script_files/

echo "Checking out base version of source file..."
if ! git show "origin/$BASE_BRANCH_NAME:$SOURCE_FILE_PATH" > "/tmp/script_files/base_source_file.yaml"; then
    echo "Failed to checkout source file from base branch: $BASE_BRANCH_NAME"
    echo "File path: $SOURCE_FILE_PATH"
    exit 1
fi

echo "Base source file saved to: /tmp/script_files/base_source_file.yaml"


node $ALLOF_MERGE_SCRIPT_PATH /tmp/script_files/base_source_file.yaml /tmp/script_files/base_source_file_flattened.yaml
node $ALLOF_MERGE_SCRIPT_PATH $SOURCE_FILE_PATH /tmp/script_files/source_file_flattened.yaml


oasdiff breaking /tmp/script_files/base_source_file_flattened.yaml /tmp/script_files/source_file_flattened.yaml --fail-on ERR --severity-levels $SEVERITY_LEVELS_FILE --err-ignore $ERR_IGNORE_FILE

