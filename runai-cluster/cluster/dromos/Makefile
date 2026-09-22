# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Library module - no binary, no container image.
# Consumers (e.g. anyworkload-controller) link dromos at compile time.
SERVICE_NAME=dromos
GO_ENVTEST_INNER_COMMAND=go test ./tests/... --ginkgo.fail-fast -timeout 30m
# Exclude the envtest/integration suite (tests/) from the unit-test job.
GO_TEST_PACKAGES=`go list -buildvcs=false ./... | grep -v /tests$$`
GO_TEST_FLAGS=-v

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_GEN_VERSION ?= v0.19.0

# WorkloadRunner CRD module - dromos owns the CRD.
API_MODULE := api/v1alpha1
CRDS_DIR   := $(shell pwd)/config/crd/bases

##@ Basic Commands

unittest: test-go ## Run unit-tests (excludes the envtest integration suite under tests/).
.PHONY: unittest

test: envtest-docker-go ## Run unit-tests + envtest integration suite (in Docker).
.PHONY: test

test-dev: envtest-go ## Run envtest integration suite locally (no Docker).
.PHONY: test-dev

# Generates third-party CRDs needed by envtest. The dependency makes them available
# before envtest-go runs.
envtest-go: generate-crds-for-envtest

lint: ## No-op - matches workload-common library convention. Lint runs against
	## consumers (e.g. anyworkload-controller) which transitively cover dromos.
.PHONY: lint

format: format-go ## Format the code.
.PHONY: format

validate: ## No-op - dromos has no container image to validate.
.PHONY: validate

build: ## No-op - dromos is a library with no binary entry point.
	@echo "dromos is a library; nothing to build standalone. Consumers (e.g. anyworkload-controller) compile dromos transitively."
.PHONY: build

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

# Generate third-party CRDs needed by envtest (jobset, NVIDIA DRA, karta).
# Outputs to tests/testdata/generated/crds (gitignored).
# jobset needs generateEmbeddedObjectMeta=true so the apiserver keeps pod-template
# metadata.labels (where the mnnvl plugin stamps dromos.run.ai/mnnvl-workload-name).
.PHONY: generate-crds-for-envtest
generate-crds-for-envtest: controller-gen
	$(CONTROLLER_GEN) crd paths=github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/... \
    output:crd:artifacts:config=tests/testdata/generated/crds
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true paths=sigs.k8s.io/jobset/api/jobset/... \
    output:crd:artifacts:config=tests/testdata/generated/crds
	$(CONTROLLER_GEN) crd paths=github.com/run-ai/karta/pkg/api/runai/v1alpha1/... \
    output:crd:artifacts:config=tests/testdata/generated/crds

##@ CRD Build (mirrors karta's two-module Makefile pattern)

.PHONY: manifests
manifests: controller-gen ## Generate WorkloadRunner CRD manifest from the api sub-module.
	cd $(API_MODULE) && $(CONTROLLER_GEN) crd:allowDangerousTypes=true,generateEmbeddedObjectMeta=true paths="./..." output:crd:artifacts:config=$(CRDS_DIR)

.PHONY: generate
generate: controller-gen ## Regenerate zz_generated.deepcopy.go in the api sub-module.
	cd $(API_MODULE) && $(CONTROLLER_GEN) object paths="./..."
	sh ${BUILD_MAKEFILE_DIR}scripts/spdx-header-go.sh $(API_MODULE)

.PHONY: copy-crds
copy-crds: manifests ## Copy WorkloadRunner CRD to runai-chart and runai-crd-upgrader.
	cp $(CRDS_DIR)/run.ai_workloadrunners.yaml ../runai-chart/charts/crds/crds/run.ai_workloadrunner.yaml
	cp $(CRDS_DIR)/run.ai_workloadrunners.yaml ../runai-crd-upgrader/crds/run.ai_workloadrunner.yaml

.PHONY: install-crd
install-crd: manifests ## Install dromos CRDs into the cluster.
	kubectl apply --server-side -f $(CRDS_DIR)

.PHONY: uninstall-crd
uninstall-crd: ## Uninstall dromos CRDs from the cluster.
	kubectl delete -f $(CRDS_DIR) --ignore-not-found

.PHONY: validate-generate
validate-generate: generate manifests copy-crds ## CI guard: fails if generated files are stale.
	@git diff --exit-code -- $(API_MODULE) $(CRDS_DIR) ../runai-chart/charts/crds/crds/run.ai_workloadrunner.yaml ../runai-crd-upgrader/crds/run.ai_workloadrunner.yaml \
	  || (echo "::error::generated files out of date - run 'make generate manifests copy-crds' and commit"; exit 1)

##@ Multi-module helpers (root + api sub-module)

.PHONY: tidy
tidy: ## go mod tidy in both modules.
	go mod tidy
	cd $(API_MODULE) && go mod tidy

.PHONY: api-fmt
api-fmt: ## go fmt the api sub-module.
	cd $(API_MODULE) && go fmt ./...

.PHONY: api-vet
api-vet: ## go vet the api sub-module.
	cd $(API_MODULE) && go vet ./...

.PHONY: api-test
api-test: ## go test the api sub-module.
	cd $(API_MODULE) && go test ./...
