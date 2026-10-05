# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Library module (no image). The Makefile exists for the shared targets,
# e.g. `make generate-mocks-local`, which regenerates the mocks configured in
# .mockery.yml.
SERVICE_NAME=workload-controller
