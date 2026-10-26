# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

REPO_DIR=$(shell pwd)
# Change for every new service
SERVICE_NAME=notifications-proxy

DOCKER_COMPOSE_FILE=./integration_tests/docker-compose.yaml
# The compose file builds the slack mock on the Go this module's go.mod asks for.
export GO_ALPINE_VERSION

OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0
OPENAPI_GENERATOR_ADDITIONAL_PROPERTY_ADD_RESPONSE_HEADERS=true

VERSIONED_CLIENT=true

GITHUB_REPO_NAME=run-ai
SHARED_PACKAGES_PATH=pkg
OPENAPI_CLIENT_NAME_GO=client
SHARED_PACKAGES_PATH_ABS=control-plane/notifications-proxy/pkg

generate-integration-test-smtp-openapi-client:
	${DOCKER_OPENAPI_CLIENT_GO_GENERATOR_COMMAND} \
    	-i ./integration_tests/mailhog/api/openapi.yaml \
		-o ./integration_tests/mailhog/mailhogclient  \
		--additional-properties packageName=mailhogclient \
		--global-property modelDocs=false,apiDocs=false  \
		--git-repo-id mailhog --git-user-id smtp \
		-g go

	cd ${REPO_DIR}/integration_tests/mailhog/mailhogclient && \
		rm -rf ./git_push.sh  ./test ./go.mod ./go.sum  ./.travis.yml ./.openapi-generator* ./.gitignore && \
		go mod tidy && \
		go fmt ./...
	sh ${BUILD_MAKEFILE_DIR}scripts/spdx-header-openapi.sh ./integration_tests/mailhog/mailhogclient

test: build test-go integration-tests ## Run unit and integration tests
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

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: openapi-generate-code build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy

#@ Integration tests
# Build the sibling notifications-service binary from monorepo source so the compose
# stack builds its image locally instead of pulling a published one from the registry.
build-notifications-service-bin: git-config-command
	env GOOS=linux GOARCH=amd64 GOCACHE=/tmp/.cache CGO_ENABLED=0 GO111MODULE=on \
	   go build -C ../notifications-service -o bin/notifications-service-amd64 ./cmd/main.go
	env GOOS=linux GOARCH=arm64 GOCACHE=/tmp/.cache CGO_ENABLED=0 GO111MODULE=on \
	   go build -C ../notifications-service -o bin/notifications-service-arm64 ./cmd/main.go
.PHONY: build-notifications-service-bin

INTEG_TESTS_PACKAGE=./integration_tests/tests/

build-integration-tests: build-notifications-service-bin build-integration-tests-generic
	docker image rm integration_tests-notifications-proxy --force
	docker image rm integration_tests-notification-proxy-integration-test --force
	docker compose -f $(DOCKER_COMPOSE_FILE) build

integration-tests: build-integration-tests
	docker compose -f $(DOCKER_COMPOSE_FILE) --profile "*" down --volumes --remove-orphans
	docker compose -f $(DOCKER_COMPOSE_FILE) --profile "*" up --remove-orphans --force-recreate --abort-on-container-exit

# used for launching docker compose without the integration tests docker, in order to run them from IDE
integration-tests-debug: build-integration-tests
	docker compose --profile run-setup --profile smtp --profile slack --profile redis -f $(DOCKER_COMPOSE_FILE) up --remove-orphans --force-recreate --abort-on-container-exit
