include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

SERVICE_NAME=workloads-and-policies

BUILD_IN_PATH=./...

##@ Basic Commands 

test: test-ginkgo ## Run unit-tests.
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate: ## Override the imported target to do nothing (silently)
	@echo "Validation step skipped."
.PHONY: validate

purify: purify-go ## Purify the code.
.PHONY: purify

build: ## no need to build we want this CI to just run tests
.PHONY: build

#########################################################################################
#####################  coverage  ########################################################
#########################################################################################

GO_ACC_IGNORE=pkg/clients,pkg/kube_client
coverage: test-go-acc
