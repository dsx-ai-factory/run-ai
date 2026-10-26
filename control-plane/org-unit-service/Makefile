# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

# Change for every new service
SERVICE_NAME=org-unit-service
OPENAPI_GENERATOR_SERVER_CODE_PATH=generated
SHARED_PACKAGES_PATH=pkg
OPENAPI_CLIENT_NAME_GO=client
SHARED_PACKAGES_PATH_ABS=control-plane/org-unit-service/pkg


##@ Basic Commands

GINKGO_COMMAND=go install github.com/onsi/ginkgo/v2/ginkgo@v2.19.0 && go get github.com/onsi/gomega/...@v1.41.0 && \
	go mod tidy && go run github.com/onsi/ginkgo/v2/ginkgo -r --procs=1  --compilers=1 --randomize-all --randomize-suites --fail-on-pending  --keep-going --timeout=5m --trace

test: test-ginkgo ## Run unit-tests.
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
