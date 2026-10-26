# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=audit-service
REPO_DIR=$(shell pwd)
BIN_DIR=${REPO_DIR}/bin

OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

VERSIONED_CLIENT=true

GITHUB_REPO_NAME=run-ai
SHARED_PACKAGES_PATH=pkg
OPENAPI_CLIENT_NAME_GO=client
SHARED_PACKAGES_PATH_ABS=control-plane/audit-service/pkg

##@ Basic Commands
test: build  test-go integration-tests## Run unit and integration tests
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate: openapi-validate validate-mockery ## Run validate.
.PHONY: validate

purify: purify-go ## Purify the code.
.PHONY: purify

build: openapi-build build-go ## Build the service.
.PHONY: build

##@ OpenAPI

openapi-lint: openapi-lint-generic ## Lint the swagger bundle.
.PHONY: openapi-lint

build-integration-tests: build-integration-tests-generic
	# build dockers
	docker compose -f ./integration-tests/docker-compose.yaml build

#@ Integration tests
integration-tests: build-integration-tests
	docker compose --profile "*" -f ./integration-tests/docker-compose.yaml down --volumes --remove-orphans
	docker compose --profile "*" -f ./integration-tests/docker-compose.yaml up --remove-orphans --force-recreate --abort-on-container-exit

# used for launching docker compose without the integration tests docker, in order to run them from IDE
integration-tests-debug: build-integration-tests
	docker compose --profile run-setup -f ./integration-tests/docker-compose.yaml up --remove-orphans --force-recreate --abort-on-container-exit
