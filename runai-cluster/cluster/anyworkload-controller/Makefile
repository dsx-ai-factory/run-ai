include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=anyworkload-controller
GO_ENVTEST_INNER_COMMAND=go test ./tests/... --ginkgo.fail-fast -timeout 30m
# Exclude the envtest/integration suite (tests/) from the unit-test job.
GO_TEST_PACKAGES=`go list -buildvcs=false ./... | grep -v /tests$$`
GO_TEST_FLAGS=-v
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_GEN_VERSION ?= v0.19.0

##@ Basic Commands

unittest: test-go ## Run unit-tests.
.PHONY: unittest

test: envtest-docker-go ## Run unit-tests and env-tests.
.PHONY: test

test-dev: envtest-go
.PHONY: test-dev

# adds to generic envtest-go dependency on generate-crds-for-envtest which is necessary to this service
envtest-go: generate-crds-for-envtest

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

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

# Generate CRDs third party for importing to envtest. 
# It generates the CRDs into a folder which is gitignored. The envtest will use this folder to import the CRDs.
.PHONY: generate-crds-for-envtest
generate-crds-for-envtest: controller-gen
	$(CONTROLLER_GEN) crd paths=github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/... \
    output:crd:artifacts:config=tests/testdata/generated/crds
	$(CONTROLLER_GEN) crd paths=sigs.k8s.io/jobset/api/jobset/... \
    output:crd:artifacts:config=tests/testdata/generated/crds
	$(CONTROLLER_GEN) crd paths=github.com/run-ai/karta/pkg/api/runai/v1alpha1/... \
    output:crd:artifacts:config=tests/testdata/generated/crds
