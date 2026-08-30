include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=inference-workload-controller
#
# Exclude the envtest/integration suite (pkg/tests) from the unit-test job.
GO_TEST_PACKAGES=`go list -buildvcs=false ./... | grep -v pkg/tests`
GO_TEST_FLAGS=-v
GO_LIST_ENABLED=0

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

#########################################################################################
#####################  coverage  ########################################################
#########################################################################################

coverage: 
	go install github.com/ory/go-acc@v0.2.8 
	go-acc --ignore "clients/assets" --ignore cmd -o /tmp/cover.out ./... 
	grep -v zz_ < /tmp/cover.out | grep -v "/generated/pkg" > cover.out
.PHONE: coverage
