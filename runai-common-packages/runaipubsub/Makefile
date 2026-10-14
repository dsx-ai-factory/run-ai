# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/index.mk

SERVICE_NAME=runaipubsub
TAG_FORMAT=runai-common-packages/$(SERVICE_NAME)

DOCKER_WORK_DIR=/local/runai-common-packages/${SERVICE_NAME}
DOCKER_COMMAND=${DOCKER_COMMAND_WORK_DIR}

test: test-go integration-tests
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate: validate-mockery
	# no op
.PHONY: validate

build:
.PHONY: build

# Local/dev only — not part of CI test/build; requires Docker and git access to private modules.
INTEG_TESTS_BIN=./integration-tests/bin/integ-tests

build-integration-tests: build-integration-tests-generic
	docker image rm nats-tests --force 2>/dev/null || true
	docker image rm kafka-tests --force 2>/dev/null || true
	docker compose -f ./integration-tests/docker-compose.yaml build
.PHONY: build-integration-tests

integration-tests: build-integration-tests
	docker compose -f ./integration-tests/docker-compose.yaml --profile redis-setup --profile kafka-setup --profile kafka-tests up --remove-orphans --force-recreate --abort-on-container-exit
	docker compose -f ./integration-tests/docker-compose.yaml down --remove-orphans --volumes
	docker compose -f ./integration-tests/docker-compose.yaml --profile redis-setup --profile nats-setup --profile nats-tests up --remove-orphans --force-recreate --abort-on-container-exit
	docker compose -f ./integration-tests/docker-compose.yaml down --remove-orphans --volumes
.PHONY: integration-tests
