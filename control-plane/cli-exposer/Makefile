# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=cli-exposer
PACKAGE=github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer

GIT_TAG=${VERSION}

ifeq (${GIT_TAG},)
    GIT_TAG=0.0.1-local
endif

GO_LDFLAGS='-X ${PACKAGE}/pkg/version.controlPlaneVersion=${GIT_TAG}'

##@ Basic Commands 

test: test-ginkgo ## Run unit-tests.
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate: openapi-validate validate-mockery ## Run validate.
.PHONY: validate

purify: purify-go ## Purify the code.
.PHONY: purify

build: openapi-build build-cli build-go ## Build the service.
.PHONY: build

##@ OpenAPI

openapi-lint: openapi-lint-generic ## Lint the swagger bundle.
.PHONY: openapi-lint

build-cli:
	@echo "Building CLI files"
	@if [ -d "./bin" ]; then rm -rfv bin; fi
	@echo "Directory exists"
	@make -C ../runai-cli dist
	@mkdir -p ./bin
	@cp -r -v ../runai-cli/bin/dist ./bin/dist
	@cp -r -v ../runai-cli/docs ./bin/docs
	@cp -r -v ../runai-cli/version ./bin/version

.PHONY: build-cli

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: openapi-generate-code build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy

coverage:
	@go test -v ./... -coverprofile cover.out
	@cat cover.out | grep -v "cli-exposer/internal/generated" | grep -v "/cmd/main.go" > filtered_cover.out
	@go tool cover -func=filtered_cover.out -o=coverage.out
	@cat coverage.out

run-local: build-cli
	@echo "Building CLI files"
	@go mod vendor
	@go build -ldflags '-X ${PACKAGE}/pkg/version.versionFilePath=bin/version \
		-X ${PACKAGE}/pkg/document.staticFilePath=internal/static \
		-X ${PACKAGE}/pkg/installer.binariesPath=bin/dist \
		-X ${PACKAGE}/pkg/installer.installerBasePath=installer' \
		cmd/main.go
	@./main
	rm -f ./main
