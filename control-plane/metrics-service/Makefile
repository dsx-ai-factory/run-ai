# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=metrics-service

OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

VERSIONED_CLIENT=true

GITHUB_REPO_NAME=run-ai
SHARED_PACKAGES_PATH=pkg
OPENAPI_CLIENT_NAME_GO=client
SHARED_PACKAGES_PATH_ABS=control-plane/metrics-service/pkg

test: build  test-ginkgo integration-tests ## Run unit and integration tests
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate: openapi-validate validate-mockery ## Run validate.
.PHONY: validate

purify: purify-go ## Purify the code.
.PHONY: purify

build: openapi-build build-go generate-mocks-client ## Build the service.
.PHONY: build

##@ OpenAPI

openapi-lint: openapi-lint-generic ## Lint the swagger bundle.
.PHONY: openapi-lint

build-integration-tests: build-integration-tests-generic
	# build dockers
	AGENT_PORT=8090 docker compose -f ./integration-tests/docker-compose.yaml build

#@ Integration tests
integration-tests: _build-promtool-agent build-integration-tests
	AGENT_PORT=8090 docker compose --profile "*" -f ./integration-tests/docker-compose.yaml down --volumes --remove-orphans
	AGENT_PORT=8090 docker compose --profile "*" -f ./integration-tests/docker-compose.yaml up --remove-orphans --force-recreate --abort-on-container-exit

# used for launching docker compose without the integration tests docker, in order to run them from IDE
integration-tests-debug: _build-promtool-agent build-integration-tests
	AGENT_PORT=8090 docker compose --profile run-setup -f ./integration-tests/docker-compose.yaml up --remove-orphans --force-recreate --abort-on-container-exit

_build-promtool-agent:
	go fmt ./integration-tests/prometheus-with-agent/...
	${DOCKER_GO_BASE_COMMAND} -e GOOS=linux -e GOARCH=amd64 ${DOCKERHUB_PROXY}library/golang:${GO_VERSION} go build -o ./integration-tests/prometheus-with-agent/bin/promtool-agent -buildvcs=false ./integration-tests/prometheus-with-agent/main/...
	chmod +x ./integration-tests/prometheus-with-agent/bin/promtool-agent

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: openapi-generate-code build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy
