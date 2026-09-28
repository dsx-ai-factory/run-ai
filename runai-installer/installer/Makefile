# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

##@ Basic Commands

SERVICE_NAME=runai-installer
BIN_REL_PATH?=bin
OUTPUT_PATH?=$(BIN_REL_PATH)/$(OS)/$(ARCH)/$(SERVICE_NAME)
export GOPRIVATE=github.com/run-ai/*

# Build tags. The released binary is built with none of them, so it exposes
# only `install` and `airgap`. INTERNAL_TAG additionally registers `upgrade`
# and `uninstall`.
BUILD_TAGS?=
INTERNAL_TAG=runai_internal
INTERNAL_BIN_REL_PATH?=bin-internal

# Derive the Go runtime version from go.mod so the linter is always built
# with a toolchain new enough for whatever go.mod currently requires.
GO_VERSION_BASE?=$(shell awk '$$1=="go"{print $$2;exit}' go.mod)
GOLANGCI_LINT_VERSION=v2.9.0
DOCKERHUB_PROXY?=

# Ensure ~/.gitconfig exists before any container bind-mounts it. Docker creates a
# root-owned DIRECTORY at the mount source when the path is missing, which then breaks
# every git command for the user (see DOCKER_GO_BASE_COMMAND in golang.mk).
git-config-command:
	@touch $(HOME)/.gitconfig
ifdef GITHUB_TOKEN
	@echo Creating gitconfig with github token
	git config --file $(HOME)/.gitconfig url.https://$(GITHUB_TOKEN)@github.com/.insteadOf https://github.com/
endif
.PHONY: git-config-command

test: git-config-command
	go test -tags "$(BUILD_TAGS)" ./...
.PHONY: test

test-internal: git-config-command
	BUILD_TAGS="$(INTERNAL_TAG)" $(MAKE) test
.PHONY: test-internal

lint: git-config-command
	docker run --rm \
		-u $(shell id -u):$(shell id -g) \
		-v "$(abspath $(CURDIR)/..):/workspace" \
		-v "$(HOME)/.gitconfig:/.gitconfig:ro" \
		-w /workspace/installer \
		-e HOME=/tmp \
		-e GOPRIVATE=$(GOPRIVATE) \
		-e GOFLAGS=-buildvcs=false \
		-e GIT_CONFIG_GLOBAL=/.gitconfig \
		--entrypoint /bin/sh \
		$(DOCKERHUB_PROXY)library/golang:$(GO_VERSION_BASE) \
		-c 'go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) && golangci-lint run -c .golangci.yaml $(if $(BUILD_TAGS),--build-tags "$(BUILD_TAGS)",)'
.PHONY: lint

lint-internal: git-config-command
	BUILD_TAGS="$(INTERNAL_TAG)" $(MAKE) lint
.PHONY: lint-internal

# Keeps the internal variant compiling and its command surface tested, so that
# a change to the released commands cannot silently break the internal build.
validate: git-config-command
	go build -tags "$(INTERNAL_TAG)" ./...
	go test -tags "$(INTERNAL_TAG)" ./internal/cli/...
.PHONY: validate

purify: git-config-command
	go vet -tags "$(BUILD_TAGS)" ./...
.PHONY: purify

build: git-config-command
	OS=linux ARCH=amd64 make build-for-platform
	OS=linux ARCH=arm64 make build-for-platform
	OS=darwin ARCH=amd64 make build-for-platform
	OS=darwin ARCH=arm64 make build-for-platform
.PHONY: build

# Same binaries plus the legacy commands, written to $(INTERNAL_BIN_REL_PATH)
# so they can never be picked up as the released artifact.
build-internal: git-config-command
	BUILD_TAGS="$(INTERNAL_TAG)" BIN_REL_PATH="$(INTERNAL_BIN_REL_PATH)" $(MAKE) build
.PHONY: build-internal

build-for-platform: git-config-command
	mkdir -p $(dir $(OUTPUT_PATH))
	GOOS=$(OS) GOARCH=$(ARCH) GOFIPS140=v1.0.0 go build -tags "$(BUILD_TAGS)" -ldflags="$(LDFLAGS)" -o $(OUTPUT_PATH) ./cmd
.PHONY: build-for-platform

test-uninstall-integration-existing: git-config-command
	go test -v -tags integration -timeout 2m \
		./internal/uninstall/deps_integration_tests/
.PHONY: test-uninstall-integration-existing

test-uninstall-integration: git-config-command
	@CLUSTER_NAME="uninstall-int-$$(uuidgen | tr '[:upper:]' '[:lower:]' | tr -d '-' | cut -c1-6)"; \
	KUBECONFIG_FILE="/tmp/$${CLUSTER_NAME}.kubeconfig"; \
	echo "--- Creating kind cluster $${CLUSTER_NAME}"; \
	kind create cluster --name "$${CLUSTER_NAME}" --kubeconfig "$${KUBECONFIG_FILE}"; \
	KUBECONFIG="$${KUBECONFIG_FILE}" \
		go test -v -tags integration -timeout 2m \
		./internal/uninstall/deps_integration_tests/; \
	EXIT=$$?; \
	echo "--- Deleting kind cluster $${CLUSTER_NAME}"; \
	kind delete cluster --name "$${CLUSTER_NAME}" || true; \
	rm -f "$${KUBECONFIG_FILE}"; \
	exit $$EXIT
.PHONY: test-uninstall-integration
