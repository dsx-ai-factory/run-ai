# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

GINKGO_PROCS?=4
GINKGO_COMPILERS?=4
GINKGO_TIMEOUT?=9m

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

# Change for every new service
SERVICE_NAME=cluster-service

VERSIONED_CLIENT=true

GITHUB_REPO_NAME=run-ai
SHARED_PACKAGES_PATH=pkg
OPENAPI_CLIENT_NAME_GO=client
SHARED_PACKAGES_PATH_ABS=control-plane/cluster-service/pkg

GINKGO_COMMAND+=--skip-package=integration-tests

##@ Basic Commands

test: build test-ginkgo docker-compose-integration-tests-up ## Run unit tests + integration tests.
.PHONY: test

unit-test: test-ginkgo ## Run unit tests only.
.PHONY: unit-test

integ-tests: build docker-compose-integration-tests-up ## Run integration tests.
.PHONY: integ-tests

build-integration-tests: build build-integration-tests-generic
	docker compose --profile "*" -f ./integration-tests/docker-compose.yaml build

docker-compose-integration-tests-up: build-integration-tests
	docker compose --profile "*" -f ./integration-tests/docker-compose.yaml up -V --remove-orphans --force-recreate --abort-on-container-exit || ${FAILURE_MESSAGE_HANDLER}
.PHONY: docker-compose-integration-tests-up

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

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: openapi-generate-code build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy
