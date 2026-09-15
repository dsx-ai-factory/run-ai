# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

SERVICE_NAME=runai-installer-clients

DOCKER_WORK_DIR=/local/runai-installer/clients
DOCKER_COMMAND=${DOCKER_COMMAND_WORK_DIR}

test: test-go
.PHONY: test

lint: lint-go
.PHONY: lint

validate:
	# no op
.PHONY: validate

build: build-go
.PHONY: build
