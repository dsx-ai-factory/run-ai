include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

SERVICE_NAME=advice-restart-controller
##@ Basic Commands

test: envtest-docker-go ## Run unit-tests and env-tests.
.PHONY: test

test-dev: envtest-go
.PHONY: test-dev

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
