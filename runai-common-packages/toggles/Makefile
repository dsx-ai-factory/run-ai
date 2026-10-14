# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/index.mk

SERVICE_NAME=toggles
TAG_FORMAT=runai-common-packages/$(SERVICE_NAME)

DOCKER_WORK_DIR=/local/runai-common-packages/${SERVICE_NAME}
DOCKER_COMMAND=${DOCKER_COMMAND_WORK_DIR}

all: purify test
.PHONY: all

test: test-go integration-tests
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate: validate-mockery
	# no op
.PHONY: validate

build:
.PHONY: build

# Integration tests
INTEG_TESTS_BIN=./integration-tests/bin/integ-tests

build-integration-tests: build-integration-tests-generic
	docker image rm toggles-test --force
	docker compose -f ./integration-tests/docker-compose.yaml build
.PHONY: build-integration-tests

integration-tests: build-integration-tests
	docker compose -f ./integration-tests/docker-compose.yaml --profile unleash-setup --profile unleash-test up --remove-orphans --force-recreate --abort-on-container-exit
.PHONY: integration-tests

integration-tests-debug:
	docker compose -f ./integration-tests/docker-compose.yaml --profile unleash-setup up --remove-orphans --force-recreate --abort-on-container-exit
.PHONY: integration-tests-debug

purify:
	go fmt ./...
	go vet ./...
	go mod tidy
.PHONY: purify
