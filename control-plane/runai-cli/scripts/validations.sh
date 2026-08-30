#!/bin/bash

CODE_DIRS="cmd/ internal/ pkg/"

# Define colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

logInfo() {
  echo "${GREEN}==> ${NC}$1"
}

logError() {
  echo "${RED}==> Error: ${NC}$1" >&2
}

validate_goimports() {
  make linter-fixes || {
    logError "Failed to run 'make linter-fixes'."
    exit 1
  }

  echo "Checking imports differences"
  pwd

  if [ -n "$(git status --porcelain $(echo ${CODE_DIRS}))" ]; then
    echo "There are changes in the repository."
    git --no-pager diff $(echo ${CODE_DIRS})
    logError "Please run 'make linter-fixes' and commit the changes."
    exit 1
  else
    echo
    logInfo "goimports validation passed, No changes found."
  fi
}

CLI_TEMP_DOCS_DIR="temp-docs"
CLI_DOCS_DIR="docs"
validate_docs() {
    echo "Checking for documentation differences"
    mkdir -p ${CLI_TEMP_DOCS_DIR}
    make build-docs-markdown CLI_DOCS_DIR=temp-docs
    echo
    diff -r ${CLI_TEMP_DOCS_DIR} ${CLI_DOCS_DIR} || {
      sleep 1
      logError "Documentation is not up to date. Please run 'make build-docs-markdown' and commit the changes."
      echo
      rm -rf temp-docs
      exit 1
    }
    rm -rf temp-docs
    echo
    logInfo "documents validation passed, No changes found."
}

CLI_TEMP_MOCKS_DIR="temp-mocks"
CLI_MOCKS_DIR="internal/autogenerate/mocks"
validate_mocks() {
    echo "Checking for mock differences"
    mkdir -p ${CLI_TEMP_MOCKS_DIR}
    make generate-mockgen MOCKS_DIR=${CLI_TEMP_MOCKS_DIR}
    echo
    diff -r ${CLI_TEMP_MOCKS_DIR} ${CLI_MOCKS_DIR} || {
      sleep 1
      logError "Mocks are not up to date. Please run 'make generate-mockgen' and commit the changes."
      echo
      rm -rf ${CLI_TEMP_MOCKS_DIR}
      exit 1
    }
    rm -rf ${CLI_TEMP_MOCKS_DIR}
    echo
    logInfo "Mocks validation passed, No changes found."
}

validate_version() {
  echo "Checking for version differences"
  HEAD_REF="origin/master"
  # Differentiate when tests are being run locally or by GitHub Actions and inside a PR
  if [ "${GITHUB_ACTIONS}" = "true" ]; then
    if [ -n "${GITHUB_BASE_REF}" ]; then
      HEAD_REF="origin/${GITHUB_BASE_REF}"
      echo "Running on GitHub Actions, inside a PR. Fetching base branch changes.."
      git fetch --no-tags --depth=1 origin "+refs/heads/${GITHUB_BASE_REF}"
    else
      logInfo "Running on GitHub Actions, but not in a PR. skipping version check.."
      return
    fi
  else
    echo "Running locally"
  fi

  echo "HEAD_REF: $HEAD_REF"
  git --no-pager diff --name-only "$HEAD_REF" -- $(echo ${CODE_DIRS}) || {
    logError "Failed to run 'git diff'."
    exit 1
  }

  if [ -n "$(git --no-pager diff --name-only "$HEAD_REF" -- $(echo ${CODE_DIRS}) )" ]; then
    echo "There are changes in the repository."

    if [ -n "$(git --no-pager diff --name-only "$HEAD_REF" -- version)" ]; then
      logInfo "Bump was made to version file."
      git --no-pager diff --name-only "$HEAD_REF" -- version
      return
    fi

    # there are changes in the code, but not in the version file
    logError "same version, please bump the version by run 'make bump-version BUMP_TYPE=<major|minor|patch>' and commit the changes."
    exit 1
  else
    echo
    logInfo "SDK version differences validation passed, No changes found."
  fi
}

validate_goimports
echo "---------------------------------"
validate_docs
echo "---------------------------------"
validate_mocks
echo "---------------------------------"
validate_version

logInfo "All validations passed."
exit 0
