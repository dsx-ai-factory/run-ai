# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

#  Change for every new service
SERVICE_NAME=policy-service

OPENAPI_GENERATOR_SERVER_CODE_PATH=generated

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk
OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

##@ Basic Commands
test: test-ginkgo
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

copy-to-backend-js:
	cp swagger/_build/swagger.yaml ../backend/js/backend/swagger/policy.yaml
.PHONY: copy-to-backend-jso

openapi-lint: openapi-lint-generic ## Lint the swagger bundle.
.PHONY: openapi-lint

#########################################################################################
#####################  coverage  ########################################################
#########################################################################################
GO_ACC_IGNORE=clients/assets
coverage: test-go-acc

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: openapi-generate-code build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy
