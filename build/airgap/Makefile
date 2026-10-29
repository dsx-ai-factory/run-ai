# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

GO ?= go
VERSION ?= 0.0.0
REPOSITORY ?= nvcr.io/nvidia/runai
CONTROL_PLANE_VERSION ?= $(VERSION)
CLUSTER_VERSION ?= $(VERSION)
CONTROL_PLANE_REPOSITORY ?= $(REPOSITORY)
CLUSTER_REPOSITORY ?= $(REPOSITORY)
MAX_WORKERS ?= 10
REPOSITORY_ROOT := ../..
IMAGE_LOCK_OUTPUT_DIR ?= $(REPOSITORY_ROOT)/dist/image-locks
ARTIFACT_LOCK_OUTPUT_DIR ?= $(REPOSITORY_ROOT)/dist/artifact-locks
ARTIFACT_LOCK_IMAGE_LOCKS ?=
NGC_HELM_BASE_URL ?= https://helm.ngc.nvidia.com/nvidia/runai-oss/charts
ARTIFACT_LOCK_CHART_URLS ?= \
	$(NGC_HELM_BASE_URL)/control-plane-$(VERSION).tgz \
	$(NGC_HELM_BASE_URL)/runai-cluster-$(VERSION).tgz
ARTIFACT_LOCK_CHART_VERSIONS ?= $(VERSION) $(VERSION)
ARTIFACT_LOCK_CLI_ROOTS ?=
ARTIFACT_LOCK_CLI_VERSIONS ?=
ARTIFACT_LOCK_CLI_SOURCES ?=
ARTIFACT_LOCK_CLI_USERNAMES ?=
ARTIFACT_LOCK_CLI_PASSWORD_ENVS ?=
ARTIFACT_LOCK_CLI_TOKEN_ENVS ?=
ARTIFACT_LOCK_HELM_USERNAME ?= $$oauthtoken
ARTIFACT_LOCK_HELM_PASSWORD_ENV ?= NGC_API_KEY
NGC_RESOURCE_BASE_URL ?= https://api.ngc.nvidia.com/v2/org/nvidia/team/runai/resources

.PHONY: build
build:
	$(GO) build ./cmd/...

.PHONY: image-lock
image-lock: image-lock-control-plane image-lock-cluster

.PHONY: image-lock-control-plane
image-lock-control-plane: image-lock-control-plane-amd64 image-lock-control-plane-arm64

.PHONY: image-lock-cluster
image-lock-cluster: image-lock-cluster-amd64 image-lock-cluster-arm64

.PHONY: image-lock-amd64
image-lock-amd64: image-lock-control-plane-amd64 image-lock-cluster-amd64

.PHONY: image-lock-arm64
image-lock-arm64: image-lock-control-plane-arm64 image-lock-cluster-arm64

.PHONY: image-lock-control-plane-amd64
image-lock-control-plane-amd64:
	$(GO) run ./cmd/image-lock \
		--name runai-control-plane \
		--version "$(CONTROL_PLANE_VERSION)" \
		--registry "$(CONTROL_PLANE_REPOSITORY)" \
		--platform linux/amd64 \
		--max-workers "$(MAX_WORKERS)" \
		--root "$(REPOSITORY_ROOT)/control-plane" \
		--chart "$(REPOSITORY_ROOT)/control-plane/control-plane-chart" \
		--output "$(IMAGE_LOCK_OUTPUT_DIR)/runai-control-plane-{version}-linux-amd64.yaml"

.PHONY: image-lock-control-plane-arm64
image-lock-control-plane-arm64:
	$(GO) run ./cmd/image-lock \
		--name runai-control-plane \
		--version "$(CONTROL_PLANE_VERSION)" \
		--registry "$(CONTROL_PLANE_REPOSITORY)" \
		--platform linux/arm64 \
		--max-workers "$(MAX_WORKERS)" \
		--root "$(REPOSITORY_ROOT)/control-plane" \
		--chart "$(REPOSITORY_ROOT)/control-plane/control-plane-chart" \
		--output "$(IMAGE_LOCK_OUTPUT_DIR)/runai-control-plane-{version}-linux-arm64.yaml"

.PHONY: image-lock-cluster-amd64
image-lock-cluster-amd64:
	$(GO) run ./cmd/image-lock \
		--name runai-cluster \
		--version "$(CLUSTER_VERSION)" \
		--registry "$(CLUSTER_REPOSITORY)" \
		--platform linux/amd64 \
		--max-workers "$(MAX_WORKERS)" \
		--root "$(REPOSITORY_ROOT)/runai-cluster" \
		--chart "$(REPOSITORY_ROOT)/runai-cluster/cluster/runai-chart" \
		--output "$(IMAGE_LOCK_OUTPUT_DIR)/runai-cluster-{version}-linux-amd64.yaml"

.PHONY: image-lock-cluster-arm64
image-lock-cluster-arm64:
	$(GO) run ./cmd/image-lock \
		--name runai-cluster \
		--version "$(CLUSTER_VERSION)" \
		--registry "$(CLUSTER_REPOSITORY)" \
		--platform linux/arm64 \
		--max-workers "$(MAX_WORKERS)" \
		--root "$(REPOSITORY_ROOT)/runai-cluster" \
		--chart "$(REPOSITORY_ROOT)/runai-cluster/cluster/runai-chart" \
		--output "$(IMAGE_LOCK_OUTPUT_DIR)/runai-cluster-{version}-linux-arm64.yaml"

.PHONY: artifact-lock
artifact-lock:
	test -n "$(strip $(ARTIFACT_LOCK_IMAGE_LOCKS))"
	test -n "$${$(ARTIFACT_LOCK_HELM_PASSWORD_ENV)}"
	$(GO) run ./cmd/artifact-lock \
		--version "$(VERSION)" \
		$(foreach value,$(ARTIFACT_LOCK_IMAGE_LOCKS),--image-lock "$(value)") \
		$(foreach value,$(ARTIFACT_LOCK_CHART_URLS),--chart-url "$(value)") \
		$(foreach value,$(ARTIFACT_LOCK_CHART_VERSIONS),--chart-version "$(value)") \
		$(foreach value,$(ARTIFACT_LOCK_CLI_ROOTS),--cli-root "$(value)") \
		$(foreach value,$(ARTIFACT_LOCK_CLI_VERSIONS),--cli-version "$(value)") \
		$(foreach value,$(ARTIFACT_LOCK_CLI_SOURCES),--cli-source "$(value)") \
		$(foreach value,$(ARTIFACT_LOCK_CLI_USERNAMES),--cli-username "$(value)") \
		$(foreach value,$(ARTIFACT_LOCK_CLI_PASSWORD_ENVS),--cli-password-env "$(value)") \
		$(foreach value,$(ARTIFACT_LOCK_CLI_TOKEN_ENVS),--cli-token-env "$(value)") \
		--helm-username '$(ARTIFACT_LOCK_HELM_USERNAME)' \
		--helm-password-env "$(ARTIFACT_LOCK_HELM_PASSWORD_ENV)" \
		--ngc-resource-base-url "$(NGC_RESOURCE_BASE_URL)" \
		--output-dir "$(ARTIFACT_LOCK_OUTPUT_DIR)"

.PHONY: test
test:
	$(GO) test ./...

.PHONY: lint
lint:
	$(GO) vet ./...
	test -z "$$(gofmt -l .)"

.PHONY: validate
validate:
	$(GO) mod tidy -diff
