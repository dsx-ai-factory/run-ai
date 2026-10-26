# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=goofys
BUILD_IN_PATH=./cmd
##@ Basic Commands

test: GO_TEST_PACKAGES=./cmd
test: test-go
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate:
  #no-op
.PHONY: validate

build: build-go ## Build the service.
.PHONY: build
