# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/index.mk

SERVICE_NAME=cache
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

# Integration tests
INTEG_TESTS_BIN=./integration-tests/bin/integ-tests

build-integration-tests: build-integration-tests-generic
	docker image rm cache-tests-redis --force
	docker image rm cache-tests-nats --force
	docker compose -f ./integration-tests/docker-compose.yaml build 
		
integration-tests: build-integration-tests
	docker compose -f ./integration-tests/docker-compose.yaml --profile inmemory-test up --remove-orphans --force-recreate --abort-on-container-exit
	docker compose -f ./integration-tests/docker-compose.yaml --profile redis-setup --profile redis-test up --remove-orphans --force-recreate --abort-on-container-exit
	docker compose -f ./integration-tests/docker-compose.yaml --profile nats-setup --profile nats-test up --remove-orphans --force-recreate --abort-on-container-exit
	docker compose -f ./integration-tests/docker-compose.yaml --profile keydb-setup --profile keydb-test up --remove-orphans --force-recreate --abort-on-container-exit
	docker compose -f ./integration-tests/docker-compose.yaml --profile valkey-setup --profile valkey-test up --remove-orphans --force-recreate --abort-on-container-exit

integration-tests-debug:
	docker compose -f ./integration-tests/docker-compose.yaml --profile redis-setup  --profile valkey-setup  --profile nats-setup --profile keydb-setup up --remove-orphans --force-recreate --abort-on-container-exit
