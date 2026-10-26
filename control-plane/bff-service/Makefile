# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=bff-service

# Customize code generation properties for this specific service
OPENAPI_GENERATOR_ADDITIONAL_PROPERTY_ADD_RESPONSE_HEADERS=true

OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

VERSIONED_CLIENT=true

GITHUB_REPO_NAME=run-ai
SHARED_PACKAGES_PATH=pkg
OPENAPI_CLIENT_NAME_GO=client
SHARED_PACKAGES_PATH_ABS=control-plane/bff-service/pkg

##@ Basic Commands 
test: build-go  test-go docker-compose-integration-tests-up ## Run unit-tests + integration tests
.PHONY: test

unit-test: build  test-go ## Run unit-tests.
.PHONY: unit-test

# Run integration tests locally with the localdebug profile
run-local: build-go
	cd integration-tests && docker-compose --profile localdebug up -V --remove-orphans --force-recreate --abort-on-container-exit
.PHONY: run-local

INTEG_TESTS_BIN=${BUILD_DIR}/integ-tests-runner

build-integration-tests: build-integration-tests-generic
	# build dockers
	docker compose -f ./integration-tests/docker-compose.yaml build

docker-compose-integration-tests-up: build-integration-tests
	docker compose --profile "*" -f ./integration-tests/docker-compose.yaml up -V --remove-orphans --force-recreate --abort-on-container-exit || ${FAILURE_MESSAGE_HANDLER}
.PHONY: docker-compose-integration-tests-up

docker-compose-integration-tests-debug: build-integration-tests
	docker compose --profile "*" -f ./integration-tests/docker-compose.yaml up -V --remove-orphans --force-recreate --abort-on-container-exit
.PHONY: docker-compose-integration-tests-debug

coverage:
	go tool cover -html=c.out
.PHONY: coverage

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
