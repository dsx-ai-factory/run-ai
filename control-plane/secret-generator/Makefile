include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=secret-generator

##@ Basic Commands

test: test-ginkgo ## Run unit-tests.
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate: ## Run validate.
	# no-op
.PHONY: validate

purify: purify-go ## Purify the code.
.PHONY: purify

build: build-go ## Build the service.
.PHONY: build

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy
