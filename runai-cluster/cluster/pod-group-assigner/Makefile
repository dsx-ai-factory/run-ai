include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=pod-group-assigner

# without "-buildvcs=false" the CI would fail with "error obtaining VCS status", see https://github.com/pre-commit/pre-commit/issues/2722
# pkg/tests is included: it now runs against a fake client (no envtest/kind), so
# it is a plain go test package like the rest.
GO_TEST_PACKAGES=`go list -buildvcs=false ./...`
GO_TEST_FLAGS=-v

##@ Basic Commands

unittest: test-go ## Run unit-tests.
.PHONY: unittest

test: test-go ## Run unit-tests (fake-client based; no kind/envtest required).
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate:
  #no-op
.PHONY: validate

format: format-go ## Format the code.
.PHONY: format

build: build-go ## Build the service.
.PHONY: build

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy
