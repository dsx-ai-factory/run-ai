
# Change for every new service
SERVICE_NAME=k8s-objects-tracker

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk
OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

##@ Basic Commands 

test: test-ginkgo
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate: openapi-validate ## Run validate.
.PHONY: validate

purify: purify-go ## Purify the code.
.PHONY: purify

build: openapi-build build-go ## Build the service.
.PHONY: build

##@ OpenAPI

openapi-lint: openapi-lint-generic ## Lint the swagger bundle.
.PHONY: openapi-lint

#########################################################################################
#####################  coverage  ########################################################
#########################################################################################

GO_ACC_IGNORE=pkg/clients
coverage: test-go-acc
