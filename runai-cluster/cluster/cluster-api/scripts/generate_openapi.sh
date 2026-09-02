#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


ORIGINAL_DIR=$(pwd)
CP_DIR="../../../control-plane"
TEMP_DIR="${CP_DIR}/cluster-api"

check_run_folder() {
  if [[ $(basename "${ORIGINAL_DIR}") != "cluster-api" ]]; then
    echo "$ORIGINAL_DIR"
    echo "Please run this script from the cluster-api root directory"
    exit 1
  fi
}

cleanup() {
	  cd ${ORIGINAL_DIR}
	  rm -rf ${TEMP_DIR}
	  exit $1
}

check_run_folder

mkdir ${TEMP_DIR}
cp -r "swagger" "${TEMP_DIR}/" || cleanup 1
cp "Makefile" "${TEMP_DIR}/" || cleanup 1

cd ${TEMP_DIR}

# Portable in-place edit: GNU sed (-i'' = filename) and BSD sed (-i '' = no-suffix backup)
# disagree on -i syntax. Use a temp file instead so this works on both Linux CI and macOS.
#
# We're now running from control-plane/cluster-api/ (two levels deep under the
# repo root), so the shared build/makefile/ at the root is reachable via ../../.
# Rewrites:
#   1. The RC `runai-cluster.mk` include -> `control-plane.mk` so CP defaults
#      apply (workdir = /local/control-plane/${SERVICE_NAME}).
#   2. The RC index.mk include -> the new shared build/makefile/index.mk.
#   3. Drops `DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}` since
#      DOCKER_WORK_DIR_CLUSTER is RC-only and undefined under CP layout.
sed -e 's|^include.*runai-cluster\.mk.*|include ../../build/makefile/control-plane.mk|' \
    -e 's|^include.*index\.mk.*|include ../../build/makefile/index.mk|' \
    -e '/^DOCKER_WORK_DIR=/d' \
    "Makefile" > "Makefile.tmp" && mv "Makefile.tmp" "Makefile" || cleanup 1

make openapi-build || cleanup 1

if [ $? -eq 0 ]; then
    cp -rf "internal/generated" "${ORIGINAL_DIR}/internal/"
    cp -rf "swagger/_build" "${ORIGINAL_DIR}/swagger/"
fi

cleanup 0
