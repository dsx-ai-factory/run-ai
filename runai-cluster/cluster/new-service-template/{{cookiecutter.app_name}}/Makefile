include ../../common/build/makefile/index.mk

# Change for every new service
SERVICE_NAME={{cookiecutter.app_name}}
# cluster or engine
SUB_DIR={%- if cookiecutter.engine_instead_of_cluster -%}engine{% else %}cluster{% endif %}

##@ Basic Commands

test: test-go ## Run unit-tests.
.PHONY: test

test-dev:
	go test -v ./... -coverprofile cover.out
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
