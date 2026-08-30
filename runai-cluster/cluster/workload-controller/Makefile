include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=workload-controller
CRDS_DIR=config/crd

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_GEN_VERSION ?= v0.16.3

##@ Basic Commands

test: envtest-docker-go ## Run unit-tests.
.PHONY: test

test-dev: envtest-go
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
local-deploy: export DOCKER_REPO_BASE=nvcr.io/nvidia/runai
local-deploy: build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy

install: $(CRDS_DIR)/run.ai_*workloads.yaml $(CRDS_DIR)/run.ai_*policies.yaml
	for crd in $^; do \
		kubectl apply -f $$crd; \
	done

uninstall: $(CRDS_DIR)/run.ai_*workloads.yaml $(CRDS_DIR)/run.ai_*policies.yaml
	for crd in $^; do \
		kubectl delete -f $$crd; \
	done

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object paths="./..."
	$(CONTROLLER_GEN) crd paths=./...

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)
