#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


# This script generates go client code from an openapi spec.
# Requires API_FILE_BASE_PATH to be set to the directory containing the openapi spec.

set -o errexit
set -o nounset
set -o pipefail

if [ -z "${API_FILE_BASE_PATH:-}" ]; then
  echo "API_FILE_BASE_PATH must be set to the directory containing the openapi spec"
  exit 1
fi

API_FILE_PATH="${API_FILE_PATH:-${API_FILE_BASE_PATH}/api.yaml}"
IGNORE_FILE_PATH="${IGNORE_FILE_PATH:-${API_FILE_BASE_PATH}/.openapi-generator-ignore}"
GENERATED_PATH="${GENERATED_PATH:-${API_FILE_BASE_PATH}/generated}"

# If the generated directory exists, remove it.
if [ -d "${GENERATED_PATH}" ]; then
  rm -rf "${GENERATED_PATH}"
fi

docker run --rm \
    -v ${API_FILE_PATH}:/local/api.yaml \
    -v ${IGNORE_FILE_PATH}:/local/.openapi-generator-ignore \
    -v ${GENERATED_PATH}:/local/generated openapitools/openapi-generator-cli:v6.0.1\
    generate \
    --ignore-file-override /local/.openapi-generator-ignore \
    -p enumClassPrefix=true \
    -i /local/api.yaml \
    -o /local/generated \
    -g go \
    --additional-properties=packageName=$PACKAGE_NAME,generateInterfaces=true,outputAsLibrary=true,structPrefix=true

# For the CI runner, we make sure the generated code owner is the current user (not root)
if [ -n "${GITHUB_ACTIONS:-}" ]; then
    sudo chown -R $(id -u):$(id -g) $GENERATED_PATH
fi