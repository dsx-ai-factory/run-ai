include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=karta-migrate
##@ Basic Commands

# envtest runs the migration's CRD-schema integration test
# (karta_envtest_test.go) against a real apiserver. Same pattern as
# anyworkload-controller's Makefile: the docker hop spins up setup-envtest,
# installs KUBEBUILDER_ASSETS, then runs `go test ./...` (so non-envtest
# tests still run too).
test: envtest-docker-go ## Run unit-tests and env-tests.
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
