# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=metrics-service

MOCKERY := ${BIN_DIR}/mockery

OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

VERSIONED_CLIENT=true

GITHUB_REPO_NAME=run-ai
SHARED_PACKAGES_PATH=pkg
OPENAPI_CLIENT_NAME_GO=client
SHARED_PACKAGES_PATH_ABS=control-plane/metrics-service/pkg

test: test-prepare test-ginkgo integration-tests ## Run unit and integration tests
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate: openapi-validate ## Run validate.
.PHONY: validate

purify: purify-go ## Purify the code.
.PHONY: purify

build: openapi-build build-go-parallel generate-mocks-client ## Build the service.
.PHONY: build

build-go-parallel:
	$(MAKE) -j2 build-go-amd build-go-arm
.PHONY: build-go-parallel

##@ OpenAPI

openapi-lint: openapi-lint-generic ## Lint the swagger bundle.
.PHONY: openapi-lint

build-integration-tests:
	# Two independent Go builds; both images embed their output, so the compose
	# build below has to follow them.
	$(MAKE) -j2 _build-promtool-agent build-integration-tests-generic
	# Every service sits behind a profile, so an unfiltered `build` selects
	# nothing and the images get built serially by `up` instead.
	AGENT_PORT=8090 docker compose --profile "*" -f ./integration-tests/docker-compose.yaml build

#@ Integration tests
integration-tests: build-integration-tests
	AGENT_PORT=8090 docker compose --profile "*" -f ./integration-tests/docker-compose.yaml down --volumes --remove-orphans
	AGENT_PORT=8090 docker compose --profile "*" -f ./integration-tests/docker-compose.yaml up --remove-orphans --force-recreate --abort-on-container-exit

# used for launching docker compose without the integration tests docker, in order to run them from IDE
integration-tests-debug: build-integration-tests
	AGENT_PORT=8090 docker compose --profile run-setup -f ./integration-tests/docker-compose.yaml up --remove-orphans --force-recreate --abort-on-container-exit

_build-promtool-agent:
	go fmt ./integration-tests/prometheus-with-agent/...
	${DOCKER_GO_BASE_COMMAND} -e GOOS=linux -e GOARCH=amd64 ${DOCKERHUB_PROXY}library/golang:${GO_VERSION} go build -o ./integration-tests/prometheus-with-agent/bin/promtool-agent -buildvcs=false ./integration-tests/prometheus-with-agent/main/...
	chmod +x ./integration-tests/prometheus-with-agent/bin/promtool-agent

# Pulling the compose images is network-bound and the build is CPU-bound, so
# they run together rather than back to back.
test-prepare:
	$(MAKE) -j2 _compose-pull build-for-test
.PHONY: test-prepare

# Purely an optimization — `up` pulls anything missing anyway, so a failure here
# (an old compose without --ignore-buildable, a flaky registry) must not fail the build.
_compose-pull:
	-AGENT_PORT=8090 docker compose --profile "*" -f ./integration-tests/docker-compose.yaml pull --ignore-buildable
.PHONY: _compose-pull

# Generates the same code as `build` but compiles a single binary: compose
# builds the image for the host platform, so only that architecture is ever
# copied in, and the second binary `build` produces is dead weight here.
ifeq ($(ARCH),arm64)
BUILD_GO_FOR_TEST=build-go-arm
else
BUILD_GO_FOR_TEST=build-go-amd
endif

# openapi-build regenerates internal/generated and pkg/client/generated, which
# the other two read, and its own prerequisites consume the bundled spec without
# declaring it — so it runs first and stays serial. The two that follow write to
# disjoint trees.
build-for-test:
	$(MAKE) -j1 openapi-build
	$(MAKE) -j2 $(BUILD_GO_FOR_TEST) generate-mocks-client
.PHONY: build-for-test

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: openapi-generate-code build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy
