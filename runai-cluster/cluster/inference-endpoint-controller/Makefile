# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}
SERVICE_NAME=inference-endpoint-controller

LOCALBIN ?= $(shell pwd)/bin
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_GEN_VERSION ?= v0.21.0
ENVTEST_VERSION = v0.24.2-0.20260925035230-4d452bc0517c
ENVTEST_K8S_VERSION = 1.36
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_VERSION ?= v5.8.1
KIND_CLUSTER ?= inference-endpoint-controller-test-e2e
IMG ?= inference-endpoint-controller:test
TEST_ARCH ?= $(shell go env GOARCH)

.PHONY: build test test-dev lint validate format manifests generate copy-crds controller-gen setup-envtest kustomize run docker-build build-installer test-e2e

build: build-go

test: git-config-command gocache
	${DOCKER_GO_BASE_COMMAND} ${DOCKERHUB_PROXY}library/golang:${GO_VERSION} make test-dev

test-dev: setup-envtest
	KUBEBUILDER_ASSETS="$$($(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN)/k8s -p path)" go test ./... -coverprofile cover.out -timeout 10m

lint: lint-go

format: format-go

validate: generate manifests copy-crds
	git diff --exit-code -- api config ../runai-chart/charts/crds/crds/inference.platform.run.ai_inferenceendpoints.yaml ../runai-crd-upgrader/crds/inference.platform.run.ai_inferenceendpoints.yaml ../runai-chart/templates/clusterroles/inference-endpoint-controller.yaml

generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd paths="./..." output:crd:artifacts:config=config/crd/bases

copy-crds: manifests
	cp config/crd/bases/inference.platform.run.ai_inferenceendpoints.yaml ../runai-chart/charts/crds/crds/
	cp config/crd/bases/inference.platform.run.ai_inferenceendpoints.yaml ../runai-crd-upgrader/crds/
	./hack/generate-helm-rbac.sh

controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN):
	mkdir -p $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

setup-envtest: envtest
	$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN)/k8s -p path

kustomize: $(KUSTOMIZE)
$(KUSTOMIZE):
	mkdir -p $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

run:
	go run ./cmd/main.go

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(TEST_ARCH) go build -o bin/$(SERVICE_NAME)-$(TEST_ARCH) ./cmd/main.go
	docker build --platform linux/$(TEST_ARCH) --build-arg TARGETARCH=$(TEST_ARCH) -t $(IMG) .

build-installer: kustomize
	mkdir -p dist
	$(KUSTOMIZE) build config/default > dist/install.yaml

.PHONY: install uninstall deploy undeploy
install: kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found -f -

deploy: kustomize
	@tmp=$$(mktemp -d); trap 'rm -rf "$$tmp"' EXIT; \
	cp -R config "$$tmp/config"; \
	(cd "$$tmp/config/manager" && $(KUSTOMIZE) edit set image controller=$(IMG)); \
	$(KUSTOMIZE) build "$$tmp/config/default" | kubectl apply -f -

undeploy: kustomize
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found -f -

test-e2e:
	KIND_CLUSTER=$(KIND_CLUSTER) ./hack/test-e2e.sh
