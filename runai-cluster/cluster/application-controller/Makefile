# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=application-controller
##@ Basic Commands

test: test-go test-application-cleanup-docker ## Run unit and API-server cleanup tests.
.PHONY: test

# Run cleanup admission against a real API server even when the unit-test
# container has no envtest assets, so CI cannot silently skip these regressions.
.PHONY: test-application-cleanup-docker test-application-cleanup
test-application-cleanup-docker: git-config-command gocache
	${DOCKER_GO_COMMAND_LINUX} make test-application-cleanup

test-application-cleanup:
	@set -eu; \
	assets="$${KUBEBUILDER_ASSETS:-}"; \
	if [ -z "$$assets" ]; then \
		assets_dir="$$(mktemp -d)"; \
		trap 'chmod -R u+w "$$assets_dir"; rm -rf "$$assets_dir"' EXIT; \
		assets="$$(go run sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_VERSION) --bin-dir "$$assets_dir" use 1.33.0 -p path)"; \
	fi; \
	test -x "$$assets/kube-apiserver"; \
	test -x "$$assets/etcd"; \
	KUBEBUILDER_ASSETS="$$assets" go test ./pkg/webhook -count=1 -timeout=3m -v

test-dev:
	go test -v ./... -coverprofile cover.out
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
local-deploy: build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy

coverage:
	go install github.com/ory/go-acc@v0.2.8 && go-acc -o /tmp/coverage.out ./... && grep -v zz_ < /tmp/coverage.out  > cover.out  
.PHONY: coverage
