include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

# Change for every new service
SERVICE_NAME=workloads

##@ Basic Commands 

test: test-ginkgo ## Run unit-tests.
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

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: openapi-generate-code build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy

#########################################################################################
#####################  coverage  ########################################################
#########################################################################################

GO_ACC_IGNORE=pkg/clients,pkg/kube_client
coverage: test-go-acc
