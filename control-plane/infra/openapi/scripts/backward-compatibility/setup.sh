#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0




if [[ -z "$BASE_BRANCH_NAME" ]]; then
    echo "Base branch name is not set, skipping backwards compatibility check"
    exit 0
fi

if [[ -z "$SOURCE_FILE_PATH" ]]; then
    echo "Source file path is required"
    exit 1
fi

if [[ -z "$SEVERITY_LEVELS_FILE" ]]; then
    echo "Severity levels file is required"
    exit 1
fi


if [[ -z "$GITHUB_TOKEN" ]]; then
    echo "GITHUB_TOKEN is required"
    exit 1
fi

if [[ -z "$GIT_REPO_PATH" ]]; then
    echo "GIT_REPO_PATH is required"
    exit 1
fi

if [[ -z "$ALLOF_MERGE_SCRIPT_PATH" ]]; then
    echo "Allof merge script path is required"
    exit 1
fi


echo "Installing Git..."
apk add git
echo "Git: $(git --version)"

echo "Installing OpenSSH client..."
apk add openssh-client


echo "Installing Node.js and npm..." 
apk add nodejs npm
echo "Node.js: $(node --version)"
echo "NPM: $(npm --version)"

original_dir=$(pwd)
script_dir=$(dirname "$ALLOF_MERGE_SCRIPT_PATH")
echo "Installing js-yaml and allof-merge in $script_dir"
cd $script_dir
npm install js-yaml
npm install allof-merge

cd $original_dir



go install github.com/oasdiff/oasdiff@v1.11.11 || (echo "Failed to install oasdiff" && exit 1)

git config --global --add safe.directory $GIT_REPO_PATH
git config --global url."https://$GITHUB_TOKEN@github.com/".insteadOf "git@github.com:"
git config --global url."https://$GITHUB_TOKEN@github.com/".insteadOf "https://github.com/"

