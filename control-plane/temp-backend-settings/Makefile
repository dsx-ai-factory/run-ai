include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=temp-backend-settings

OPENAPI_GENERATOR_ADDITIONAL_PROPERTY_ADD_RESPONSE_HEADERS=true

OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

MOCKERY=$(BIN_DIR)/mockery
MOCKERY_VERSION=v3.2.5

GITHUB_REPO_NAME=run-ai
SHARED_PACKAGES_PATH=pkg
OPENAPI_CLIENT_NAME_GO=client
SHARED_PACKAGES_PATH_ABS=control-plane/temp-backend-settings/pkg

##@ Basic Commands

validate_mockery:
	@mkdir -p ${BIN_DIR}
	@if [ -e ${MOCKERY} ] && [ "$$(${MOCKERY} --version | awk '{print $$NF}')" != "${MOCKERY_VERSION}" ]; then \
		echo "existing mockery binary is not of required version ${MOCKERY_VERSION}. Reinstalling it"; \
		rm -f ${MOCKERY}; \
	fi
	@if [ ! -e ${MOCKERY} ]; then \
		GOTOOLCHAIN=go1.26.3 GOBIN=${BIN_DIR} go install github.com/vektra/mockery/v3@${MOCKERY_VERSION}; \
	fi

mockery: validate_mockery
	${MOCKERY}

test: build-go test-go ## Run unit-tests.
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
