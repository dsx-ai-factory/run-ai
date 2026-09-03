include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=org-units-migrator

##@ Basic Commands

GINKGO_COMMAND=go install github.com/onsi/ginkgo/v2/ginkgo@v2.19.0 && go get github.com/onsi/gomega/...@v1.42.0 && \
	go mod tidy && go run github.com/onsi/ginkgo/v2/ginkgo -r --procs=1  --compilers=1 --randomize-all --randomize-suites --fail-on-pending  --keep-going --timeout=5m --trace

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
