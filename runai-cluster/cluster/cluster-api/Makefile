# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=cluster-api
OPENAPI_CLIENT_SUB_FOLDER=generated

##@ Basic Commands

test: test-go ## Run unit-tests.
.PHONY: test

test-dev:
	go test -v ./... -coverprofile cover.out
.PHONY: test-dev

lint: lint-go ## Run linter.
.PHONY: lint

# validate fails CI if the committed OpenAPI codegen is out of sync with
# the source swagger. Re-runs generate-openapi and diffs against HEAD.
# Note: locally this overwrites internal/generated and swagger/_build if
# they're stale — CI runs from a clean checkout so the diff is meaningful.
validate: validate-openapi-uptodate
.PHONY: validate

validate-openapi-uptodate:
	@echo "Validating generated OpenAPI is in sync with source swagger..."
	@$(MAKE) generate-openapi
	@# git status --porcelain (not git diff --quiet) so we catch both modified
	@# tracked files AND newly-generated untracked files (e.g. a new model file
	@# from a new endpoint that someone forgot to git add).
	@status=$$(git status --porcelain -- internal/generated swagger/_build); \
	if [ -n "$$status" ]; then \
		echo ""; \
		echo "ERROR: Generated OpenAPI code is out of sync with source swagger."; \
		echo "Run 'make generate-openapi' and commit the result."; \
		echo ""; \
		echo "$$status"; \
		exit 1; \
	fi
	@echo "OpenAPI generated code is in sync"
.PHONY: validate-openapi-uptodate

format: format-go

# Here we can add the automatic solution for some to the linter issues.
.PHONY: linter-fixes
linter-fixes:
	@echo "Fixing import order in linter"
ifeq (, $(shell which goimports))
	go install golang.org/x/tools/cmd/goimports@latest
	ls $$(go env GOPATH)/bin/goimports || (echo "goimports is not found" && exit 1)
else
	@echo "goimports is already installed"
endif
	$$(go env GOPATH)/bin/goimports -local "github.com/GoogleContainerTools/skaffold" -w .

build: build-go ## Build the service.
.PHONY: build

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy

.PHONY: install-requirements
install-requirements:
	@echo "Installing Go packages..."
	@for pkg in $(PACKAGES); do \
		echo "Installing $$pkg..."; \
		go install $$pkg; \
	done
	@echo "All packages installed"

.PHONY: generate-mockgen
generate-mockgen: install-requirements
	@echo "Generating mocks..."
	mockgen -source=internal/portforward/handler.go -destination=internal/autogenerate/mocks/portforward/handler.go -package=mocks
	mockgen -source=internal/websocket/upgrader.go -destination=internal/autogenerate/mocks/websocket/upgrader.go -package=mocks
	mockgen -source=internal/websocket/service/pod.go -destination=internal/autogenerate/mocks/websocket/service/pod.go -package=mocks
	mockgen -source=internal/exec/handler.go -destination=internal/autogenerate/mocks/exec/handler.go -package=mocks
	mockgen -source=internal/exec/service.go -destination=internal/autogenerate/mocks/exec/service.go -package=mocks
	mockgen -source=internal/attach/handler.go -destination=internal/autogenerate/mocks/attach/handler.go -package=mocks
	mockgen -source=internal/attach/service.go -destination=internal/autogenerate/mocks/attach/service.go -package=mocks
	mockgen -source=internal/logs/handler.go -destination=internal/autogenerate/mocks/logs/handler.go -package=mocks
	mockgen -source=internal/logs/service.go -destination=internal/autogenerate/mocks/logs/service.go -package=mocks
	mockgen -source=internal/authz/authorizers/authorizer.go -destination=internal/autogenerate/mocks/authz/authorizers/authorizer.go -package=authz
	mockgen -source=internal/workload/middleware.go -destination=internal/autogenerate/mocks/workload/locator.go -package=workload
	sh ${BUILD_MAKEFILE_DIR}scripts/spdx-header-go.sh internal/autogenerate/mocks

generate-openapi:
	@bash scripts/generate_openapi.sh
