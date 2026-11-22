# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../../build/makefile/runai-cluster.mk
include ../../../../build/makefile/index.mk
WORKSPACE_HOST_PATH=${PWD}/../../../..
GOLANG_LINTER_CONFIG_PATH=../../../common/build/lint/.golangci.yaml
DOCKER_WORK_DIR=/local/runai-cluster/cluster/common/runaipublic

# Change for every new service
SERVICE_NAME=cluster-common-runaipublic
##@ Basic Commands

build: ## Build the service.
	# no-op
.PHONY: build

lint: lint-go ## Run linter.
.PHONY: lint

test: test-go ## Run unit-tests.
.PHONY: test

validate: ## Validate that generated files are up to date.
	# no-op
.PHONY: validate
