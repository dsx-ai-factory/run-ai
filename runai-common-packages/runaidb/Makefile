# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/index.mk


SERVICE_NAME=runaidb
TAG_FORMAT=runai-common-packages/$(SERVICE_NAME)

DOCKER_WORK_DIR=/local/runai-common-packages/${SERVICE_NAME}
DOCKER_COMMAND=${DOCKER_COMMAND_WORK_DIR}

test: test-go integration-tests
.PHONY: test

integration-tests:
	cd integration-tests && go test -v ./...
.PHONY: integration-tests

lint: lint-go ## Run linter.
.PHONY: lint

validate: validate-mockery
	# no op
.PHONY: validate

build:
	# no op
.PHONY: build

purify:
	go fmt ./...
	go vet ./...
	go mod tidy
.PHONY: purify
