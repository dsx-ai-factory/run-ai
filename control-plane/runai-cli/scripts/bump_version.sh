#!/bin/bash

VERSION_FILE="version"

current_dir=$(pwd)

check_run_folder() {
  if [[ $(basename "${current_dir}") != "runai-cli" ]]; then
    echo "$current_dir"
    echo "Please run this script from the runai-cli directory"
    exit 1
  fi
}

check_os() {
    OS_TYPE="Unknown operating system"
    # Check if the operating system is Linux
    if [ -f /etc/os-release ]; then
        OS_TYPE="linux"
    # Check if the operating system is macOS
    elif [ -f /System/Library/CoreServices/SystemVersion.plist ]; then
        if [[ $(uname) == "Darwin" ]]; then
          OS_TYPE="darwin"
        fi
    fi
    echo "$OS_TYPE"
}

check_run_folder

# Get the current version
# shellcheck source=./version
source "./${VERSION_FILE}"

## Bump the version
major="${RUNAI_CLI_MAJOR}"
minor="${RUNAI_CLI_MINOR}"
patch="${RUNAI_CLI_PATCH}"

echo "Current version: $major.$minor.$patch"

# Increment the appropriate part of the version number
case "$1" in
    major)
        ((major++))
        minor=0
        patch=0
        ;;
    minor)
        ((minor++))
        patch=0
        ;;
    patch)
        ((patch++))
        ;;
    *)
        echo "Usage: $0 {major|minor|patch}"
        exit 1
esac

os=$(check_os)

# Update the version in the environment file
case ${os} in
  linux)
    sed -i "s/^RUNAI_CLI_MAJOR=.*/RUNAI_CLI_MAJOR=$major/" "$VERSION_FILE"
    sed -i "s/^RUNAI_CLI_MINOR=.*/RUNAI_CLI_MINOR=$minor/" "$VERSION_FILE"
    sed -i "s/^RUNAI_CLI_PATCH=.*/RUNAI_CLI_PATCH=$patch/" "$VERSION_FILE"
    ;;
  darwin)
    sed -i '' "s/^RUNAI_CLI_MAJOR=.*/RUNAI_CLI_MAJOR=$major/" "$VERSION_FILE"
    sed -i '' "s/^RUNAI_CLI_MINOR=.*/RUNAI_CLI_MINOR=$minor/" "$VERSION_FILE"
    sed -i '' "s/^RUNAI_CLI_PATCH=.*/RUNAI_CLI_PATCH=$patch/" "$VERSION_FILE"
    ;;
  *)
    echo "Unknown operating system"
    exit 1
esac

new_version="$major.$minor.$patch"
echo "Version bumped to $new_version"
