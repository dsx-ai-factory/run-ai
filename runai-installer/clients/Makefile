# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Run:ai API Clients Generation Makefile
#
# This Makefile generates Go API clients from 3 Run:ai service OpenAPI specifications.
# Specs must be BUNDLED (all $refs resolved) before use.
#
# To get bundled specs, either:
# 1. Copy from runai repo: control-plane/infra/openapi/specs/{service-name}.yaml
# 2. Or bundle locally in runai repo: make openapi-bundle (in each service dir)
#
# x-internal paths are filtered out using --openapi-normalizer REMOVE_X_INTERNAL=true

SHELL := /bin/bash

# Versions (matching runai infra)
REDOCLY_CLI_VERSION=1.12.0
OPENAPI_GENERATOR_VERSION=v7.7.0

# Docker commands
DOCKER_COMMAND=docker run --rm -u $(shell id -u):$(shell id -g) -v "${PWD}:/local" -w /local

# Golang version for formatting
GOLANG_VERSION=1.25.5-bookworm

# Spec files (must be bundled/resolved specs)
SPEC_DIR=specs
SPEC_AUTH=$(SPEC_DIR)/authorization.yaml
SPEC_CLUSTER=$(SPEC_DIR)/cluster-service.yaml
SPEC_ORGUNIT=$(SPEC_DIR)/org-unit-service.yaml

# APIs to generate from each spec (colon-separated list)
APIS_AUTH=AccessRules:Roles:Permissions:PermissionSets
APIS_CLUSTER=Clusters:NodePools:Nodes
APIS_ORGUNIT=Departments:Projects

# OpenAPI Generator base command
DOCKER_OPENAPI_COMMAND=$(DOCKER_COMMAND) -e JAVA_OPTS="-Dlog.level=error" openapitools/openapi-generator-cli:$(OPENAPI_GENERATOR_VERSION)

# Common generator options
GENERATOR_COMMON_OPTS=--additional-properties packageVersion=1.0.0 \
	--additional-properties generateInterfaces=true \
	--additional-properties enumClassPrefix=true \
	--additional-properties hideGenerationTimestamp=true \
	--additional-properties structPrefix=true \
	--openapi-normalizer REMOVE_X_INTERNAL=true \
	-g go

##@ Help

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Generation

.PHONY: generate
generate: check-specs generate-auth generate-cluster generate-orgunit format-code spdx-header ## Generate all clients (filters x-internal)
	@echo ""
	@echo "=========================================="
	@echo "Run:ai API client generation complete!"
	@echo "=========================================="
	@echo "Generated packages:"
	@echo "  - runai/authorization/"
	@echo "  - runai/cluster/"
	@echo "  - runai/orgunit/"

.PHONY: generate-auth
generate-auth: ## Generate authorization service client
	@echo ""
	@echo "=== Generating Authorization client (APIs: $(APIS_AUTH)) ==="
	@rm -rf runai/authorization
	@mkdir -p runai/authorization
	@$(DOCKER_OPENAPI_COMMAND) generate \
		-i /local/$(SPEC_AUTH) \
		-o /local/runai/authorization \
		--additional-properties packageName=authorization \
		$(GENERATOR_COMMON_OPTS) \
		--global-property "apis=$(APIS_AUTH)" \
		--global-property models,supportingFiles \
		--global-property apiTests=false,modelTests=false
	@rm -rf runai/authorization/test runai/authorization/.openapi-generator runai/authorization/api
	@rm -f runai/authorization/go.mod runai/authorization/go.sum
	@echo "Authorization client generated"

.PHONY: generate-cluster
generate-cluster: ## Generate cluster service client
	@echo ""
	@echo "=== Generating Cluster client (APIs: $(APIS_CLUSTER)) ==="
	@rm -rf runai/cluster
	@mkdir -p runai/cluster
	@$(DOCKER_OPENAPI_COMMAND) generate \
		-i /local/$(SPEC_CLUSTER) \
		-o /local/runai/cluster \
		--additional-properties packageName=cluster \
		$(GENERATOR_COMMON_OPTS) \
		--global-property "apis=$(APIS_CLUSTER)" \
		--global-property models,supportingFiles \
		--global-property apiTests=false,modelTests=false
	@rm -rf runai/cluster/test runai/cluster/.openapi-generator runai/cluster/api
	@rm -f runai/cluster/go.mod runai/cluster/go.sum
	@echo "Cluster client generated"

.PHONY: generate-orgunit
generate-orgunit: ## Generate org-unit service client
	@echo ""
	@echo "=== Generating OrgUnit client (APIs: $(APIS_ORGUNIT)) ==="
	@rm -rf runai/orgunit
	@mkdir -p runai/orgunit
	@$(DOCKER_OPENAPI_COMMAND) generate \
		-i /local/$(SPEC_ORGUNIT) \
		-o /local/runai/orgunit \
		--additional-properties packageName=orgunit \
		$(GENERATOR_COMMON_OPTS) \
		--global-property "apis=$(APIS_ORGUNIT)" \
		--global-property models,supportingFiles \
		--global-property apiTests=false,modelTests=false
	@rm -rf runai/orgunit/test runai/orgunit/.openapi-generator runai/orgunit/api
	@rm -f runai/orgunit/go.mod runai/orgunit/go.sum
	@echo "OrgUnit client generated"

.PHONY: spdx-header
spdx-header: ## Inject NVIDIA SPDX headers into the generated clients (openapi-generator emits none)
	@echo ""
	@echo "Stamping SPDX headers on generated clients..."
	@sh hack/spdx-header.sh runai

.PHONY: format-code
format-code: ## Format generated Go code
	@echo ""
	@echo "Formatting generated code..."
	@docker run --rm -w /local -v "${PWD}:/local:z" -u $(shell id -u):$(shell id -g) \
		-e 'GOPRIVATE=github.com/run-ai/*' \
		-e CGO_ENABLED=0 \
		-e GO111MODULE=on \
		-e HOME=/tmp \
		-e GOCACHE=/tmp/.cache/go-build \
		-e GOPATH=/tmp/go \
		golang:$(GOLANG_VERSION) sh -c "cd /local && go fmt ./runai/..." 2>/dev/null || true
	@echo "Running go mod tidy..."
	@docker run --rm -w /local -v "${PWD}:/local:z" -u $(shell id -u):$(shell id -g) \
		-e 'GOPRIVATE=github.com/run-ai/*' \
		-e CGO_ENABLED=0 \
		-e GO111MODULE=on \
		-e HOME=/tmp \
		-e GOCACHE=/tmp/.cache/go-build \
		-e GOPATH=/tmp/go \
		golang:$(GOLANG_VERSION) sh -c "cd /local && go mod tidy" 2>/dev/null || true

.PHONY: check-specs
check-specs: ## Verify all spec files exist
	@echo "Checking spec files..."
	@for spec in $(SPEC_AUTH) $(SPEC_CLUSTER) $(SPEC_ORGUNIT); do \
		if [ ! -s $$spec ]; then \
			echo "Error: $$spec not found or empty."; \
			echo ""; \
			echo "Specs must be BUNDLED (all \$$refs resolved)."; \
			echo "Get bundled specs from runai repo:"; \
			echo "  control-plane/infra/openapi/specs/{service-name}.yaml"; \
			exit 1; \
		fi; \
		echo "  Found: $$spec"; \
	done
	@echo "All spec files found"

##@ Bundling (requires runai repo structure)

.PHONY: bundle-specs
bundle-specs: ## Bundle specs using redocly (requires source files)
	@echo "Bundling specs with redocly..."
	@mkdir -p $(SPEC_DIR)
	@if [ -f swagger/source/openapi.yaml ]; then \
		$(DOCKER_COMMAND) redocly/cli:$(REDOCLY_CLI_VERSION) bundle swagger/source/openapi.yaml -o $(SPEC_DIR)/bundled.yaml; \
	else \
		echo "No swagger/source/openapi.yaml found. This target requires runai repo structure."; \
		echo "Copy bundled specs from: control-plane/infra/openapi/specs/"; \
	fi

##@ Utilities

.PHONY: clean
clean: ## Clean all generated files
	@echo "Cleaning generated files..."
	@rm -rf runai/authorization runai/cluster runai/orgunit
	@echo "Clean complete"

.PHONY: show-apis
show-apis: ## Show which APIs will be generated from each spec
	@echo "APIs to generate:"
	@echo ""
	@echo "Authorization ($(SPEC_AUTH)) -> runai/authorization/"
	@echo "$(APIS_AUTH)" | tr ':' '\n' | sed 's/^/  - /'
	@echo ""
	@echo "Cluster ($(SPEC_CLUSTER)) -> runai/cluster/"
	@echo "$(APIS_CLUSTER)" | tr ':' '\n' | sed 's/^/  - /'
	@echo ""
	@echo "OrgUnit ($(SPEC_ORGUNIT)) -> runai/orgunit/"
	@echo "$(APIS_ORGUNIT)" | tr ':' '\n' | sed 's/^/  - /'

.PHONY: validate-specs
validate-specs: check-specs ## Validate OpenAPI specs
	@echo "Validating specs..."
	@for spec in $(SPEC_AUTH) $(SPEC_CLUSTER) $(SPEC_ORGUNIT); do \
		echo "Validating $$spec..."; \
		$(DOCKER_COMMAND) redocly/cli:$(REDOCLY_CLI_VERSION) lint $$spec || true; \
	done
