include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_COMMON}

# Change for every new service
SERVICE_NAME=clients
### Provide the following variables to the make command:
CLIENT=
PACKAGE_NAME=
###

CWD=$(shell pwd)

GENERATED_PATH=$(CWD)/$(CLIENT)/generated
API_FILE_BASE_PATH=$(CWD)/$(CLIENT)/api

##@ Basic Commands

build: ## Build the service.
	# no-op
.PHONY: build

lint: ## Run linter.
	# no-op
.PHONY: lint

test:  ## Run unit-tests.
	# no-op
.PHONY: test

validate: ## Validate that generated files are up to date.
	# no-op
.PHONY: validate

generate-control-plane-client:
	make generate-client CLIENT=control-plane PACKAGE_NAME=cpclient
.PHONY: generate-cp-client

generate-client:
	rm -rf $(GENERATED_PATH)
	GENERATED_PATH=$(GENERATED_PATH) \
	API_FILE_BASE_PATH=$(API_FILE_BASE_PATH) \
	PACKAGE_NAME=$(PACKAGE_NAME) \
	./scripts/generate-openapi-client.sh
.PHONY: generate-client
