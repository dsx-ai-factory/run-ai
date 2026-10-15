# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

SERVICE_NAME=go-operator
# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= latest

DEBUG=0
ifeq ($(DEBUG), 1)
DOCKER_BUILD_ADDITIONAL_ARGS=-f ./Dockerfile.debug
GOBUILD_ADDITIONAL_FLAGS=-gcflags="all=-N -l"
endif

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# run.ai/operator-bundle:$VERSION and run.ai/operator-catalog:$VERSION.
REGISTRY ?= nvcr.io/nvidia/runai
IMAGE_TAG_BASE ?= ${REGISTRY}/go-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=runai-operator crd:allowDangerousTypes=true,generateEmbeddedObjectMeta=true webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: generate-mocks-local controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	@echo "**********************************************************"
	@echo "*  NOTE!                                                 *"
	@echo "*  This generates *go-operator* related code only.    	 *"
	@echo "*  If you want to generate code that relates to our SDK, *"
	@echo "*  you should run \`make generate\` in the SDK directory   *"
	@echo "*  instead.                                              *"
	@echo "**********************************************************"
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: generate-client
generate-client: manifests ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	@echo "**********************************************************"
	@echo "*  NOTE!                                                 *"
	@echo "*  This generates *go-operator* related manifests only.	*"
	@echo "*  If you want to generate code that relates to our SDK,	*"
	@echo "*  you should run \`make manifests\` in the SDK directory  *"
	@echo "*  instead.                                              *"
	@echo "**********************************************************"
	hack/update-client.sh

.PHONY: check-updated-crds
check-updated-crds:
	hack/check-updated-crds.sh

.PHONY: test
test: validate test-go test-asset-admission-docker

# These standard Go envtests need real API-server binaries. Run them explicitly
# so neither a missing KUBEBUILDER_ASSETS nor Ginkgo spec discovery skips them.
.PHONY: test-asset-admission-docker test-asset-admission
test-asset-admission-docker: git-config-command gocache
	${DOCKER_GO_COMMAND_LINUX} make test-asset-admission

test-asset-admission:
	@assets="$${KUBEBUILDER_ASSETS:-}"; \
	if [ -z "$$assets" ]; then \
		assets_dir="$$(mktemp -d)"; \
		trap 'chmod -R u+w "$$assets_dir" && rm -rf "$$assets_dir"' EXIT; \
		assets="$$(go run sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_VERSION) --bin-dir "$$assets_dir" use 1.33.0 -p path)"; \
	fi; \
	test -x "$$assets/kube-apiserver" && test -x "$$assets/etcd"; \
	KUBEBUILDER_ASSETS="$$assets" go test ./internal/run/operands/agent/resources \
		-run '^TestAssetIdentityAdmissionPolicies$$' -count=1 -timeout=3m -v

.PHONY: test-unit  
test-unit: test-go ## Run unit tests only

.PHONY: test-envtest
test-envtest: manifests generate envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path --arch amd64)" \
		$(GINKGO_COMMAND) run -v -r --skip-package=./tests --ldflags=${LDFLAGS} -cover --output-dir ./testbin/ ./...
	egrep -v 'tests/|apis/' < ./testbin/coverprofile.out > cover.out

.PHONY: lint
lint: lint-go ## Run linter.

.PHONY: validate
validate: purify-go validate-mockery

##@ Build
.PHONY: build
build: build-go

.PHONY: build-darwin-arm
build-darwin-arm: vet fmt manifests generate check-updated-crds ## Build manager binary.
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/manager \
		$(GOBUILD_ADDITIONAL_FLAGS)\
		-ldflags=${LDFLAGS} \
		cmd/main.go

.PHONY: run
run: generate ## Run a controller from your host.
	go run -ldflags=${LDFLAGS} ./cmd/main.go


ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.3)

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-install-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell pwd)
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
