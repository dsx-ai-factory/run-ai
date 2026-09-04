DOCKER_REPO_BASE?=nvcr.io/nvidia/runai
WORKSPACE_HOST_PATH?=${PWD}/..
DOCKER_WORK_DIR?=/local/runai-mcp-server
GOLANG_LINTER_CONFIG_PATH?=../control-plane/infra/lint/.golangci.yaml

include ../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=runai-mcp-server
PACKAGE=github.com/dsx-ai-factory/run-ai/runai-mcp-server

# Override golang.mk defaults: our entrypoint lives at cmd/runai-mcp/main.go,
# not the default cmd/main.go.
BUILD_IN_PATH=cmd/runai-mcp/main.go

# Stamp the binary with the git tag (or local fallback) so the MCP server can
# report a real version to clients via mcp.Implementation.Version.
GIT_TAG=${VERSION}
ifeq (${GIT_TAG},)
    GIT_TAG=0.0.0-local
endif
GO_LDFLAGS='-X ${PACKAGE}/internal/server.serverVersion=${GIT_TAG}'

RUNAI_BASE_URL?=
RUNAI_CLIENT_ID?=
RUNAI_CLIENT_SECRET?=

##@ Basic Commands

test: test-go ## Run unit tests.
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

purify: purify-go ## go fmt + go vet + go mod tidy.
.PHONY: purify

build: build-go ## Build the service (amd64 + arm64).
.PHONY: build

##@ MCP Inspector

inspect: ## Launch the MCP inspector against a local `go run` of the server.
	@$(foreach v,RUNAI_BASE_URL RUNAI_CLIENT_ID RUNAI_CLIENT_SECRET,$(if $(strip $($(v))),,$(error $(v) is required but not set)))
	RUNAI_BASE_URL=${RUNAI_BASE_URL} RUNAI_CLIENT_ID=${RUNAI_CLIENT_ID} RUNAI_CLIENT_SECRET=${RUNAI_CLIENT_SECRET} \
	npx @modelcontextprotocol/inspector go run ./cmd/runai-mcp serve
.PHONY: inspect
