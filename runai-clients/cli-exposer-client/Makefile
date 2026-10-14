# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

SERVICE_NAME=cli-exposer-client

DOCKER_WORK_DIR=/local/runai-clients/${SERVICE_NAME}
DOCKER_COMMAND=${DOCKER_COMMAND_WORK_DIR}

test: test-go
.PHONY: test

lint: 
	# no op
.PHONY: lint

validate: validate-mockery
	# no op
.PHONY: validate

build: 
	# no op
.PHONY: build
