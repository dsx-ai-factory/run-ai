
# Change for every new service
SERVICE_NAME=assets-service
OPENAPI_GENERATOR_SERVER_CODE_PATH=generated

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk
OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.12.0

GINKGO_COMMAND=go run github.com/onsi/ginkgo/v2/ginkgo -r --procs=1 --compilers=4 --randomize-all --randomize-suites --keep-going --timeout=15m --trace --skip-package=third_party

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

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: openapi-generate-code build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: openapi-generate-code build-go docker-build-generic docker-push-generic
.PHONY: local-deploy

#########################################################################################
#####################  coverage  ########################################################
#########################################################################################

GO_ACC_IGNORE=pkg/clients,pkg/kube_client,cmd
coverage: test-go-acc
