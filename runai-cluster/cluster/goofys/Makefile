include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=goofys
##@ Basic Commands

test:
  #no-op
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate:
  #no-op
.PHONY: validate

build: build-go ## Build the service.
.PHONY: build
