# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

DOCKER_WORK_DIR?=/local/runai-installer/installer
GOLANG_LINTER_CONFIG_PATH?=.golangci.yaml

include ../../build/makefile/index.mk

##@ Basic Commands

SERVICE_NAME=runai-installer
BIN_REL_PATH?=bin
# Windows needs the .exe suffix for the shell to treat the file as a program.
BIN_EXT=$(if $(filter windows,$(OS)),.exe,)
OUTPUT_PATH?=$(BIN_REL_PATH)/$(OS)/$(ARCH)/$(SERVICE_NAME)$(BIN_EXT)
export GOPRIVATE=github.com/run-ai/*

# Build tags. The released binary is built with none of them, so it exposes
# only `install` and `airgap`. INTERNAL_TAG additionally registers `upgrade`
# and `uninstall`.
BUILD_TAGS?=
INTERNAL_TAG=runai_internal
INTERNAL_BIN_REL_PATH?=bin-internal

GO_TEST_FLAGS=$(if $(BUILD_TAGS),-tags "$(BUILD_TAGS)",)
test: test-go ## Run unit tests.
.PHONY: test

test-internal: git-config-command
	BUILD_TAGS="$(INTERNAL_TAG)" $(MAKE) test
.PHONY: test-internal

GO_BUILD_TAGS=$(BUILD_TAGS)
lint: lint-go ## Run linter.
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
	OS=windows ARCH=amd64 make build-for-platform
	OS=windows ARCH=arm64 make build-for-platform
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
