##@ Basic Commands

SERVICE_NAME=provisioner-service
BIN_REL_PATH?=bin
export GOPRIVATE=github.com/run-ai/*

GOLANGCI_LINT_VERSION=v2.5.0
OPENAPI_GENERATOR_VERSION_SERVER=v7.12.0
REDOCLY_CLI_VERSION=1.12.0
DOCKERHUB_PROXY?=

DOCKER_TAG?=0.0.0-dev
DOCKER_REPO_BASE?=nvcr.io/nvidia/runai
DOCKER_REPO_FULL=${DOCKER_REPO_BASE}/runai-provisioner
DOCKER_IMAGE_NAME=${DOCKER_REPO_FULL}:${DOCKER_TAG}

UNAME_P := $(shell uname -p)
ifeq ($(UNAME_P),arm)
ARCH?=arm64
else
ARCH?=amd64
endif

git-config-command:
ifdef GITHUB_TOKEN
	@echo Creating gitconfig with github token
	git config --file $(HOME)/.gitconfig url.https://$(GITHUB_TOKEN)@github.com/.insteadOf https://github.com/
endif
.PHONY: git-config-command

test: git-config-command
	go test $$(go list ./... | grep -v /e2e) -short
.PHONY: test

lint: git-config-command
	docker run --rm \
		-u $(shell id -u):$(shell id -g) \
		-v "$(abspath $(CURDIR)/..):/workspace" \
		-v "$(HOME)/.gitconfig:/.gitconfig:ro" \
		-w /workspace/provisioner-service \
		-e HOME=/tmp \
		-e GOPRIVATE=$(GOPRIVATE) \
		-e GOFLAGS=-buildvcs=false \
		-e GIT_CONFIG_GLOBAL=/.gitconfig \
		$(DOCKERHUB_PROXY)golangci/golangci-lint:$(GOLANGCI_LINT_VERSION)-alpine \
		golangci-lint run -c .golangci.yaml
.PHONY: lint

validate: openapi-lint
.PHONY: validate

purify: git-config-command
	go fmt ./...
	go vet ./...
	go mod tidy
.PHONY: purify

build: git-config-command
	ARCH=amd64 $(MAKE) build-for-platform
	ARCH=arm64 $(MAKE) build-for-platform
.PHONY: build

build-for-platform: git-config-command
	GOOS=linux GOARCH=$(ARCH) CGO_ENABLED=0 GOFIPS140=v1.0.0 go build -o $(BIN_REL_PATH)/$(SERVICE_NAME)-$(ARCH) ./cmd
.PHONY: build-for-platform

##@ OpenAPI

openapi-generate-server-code:
	@ echo "Removing old generated code..."
	rm -rf internal/generated
	@ echo "Generating OpenAPI server code from source..."
	docker run --rm -u $(shell id -u):$(shell id -g) -v "$(CURDIR):/local" \
		$(DOCKERHUB_PROXY)openapitools/openapi-generator-cli:$(OPENAPI_GENERATOR_VERSION_SERVER) generate \
		-i /local/swagger/source/openapi.yaml \
		-g go-server \
		-o /local/internal/generated \
		--additional-properties outputAsLibrary=true \
		--additional-properties sourceFolder=pkg/api \
		--additional-properties packageName=api \
		--additional-properties router=chi \
		--additional-properties hideGenerationTimestamp=true \
		--additional-properties enumClassPrefix=true \
		--openapi-normalizer REMOVE_X_INTERNAL=true
	go fmt ./internal/generated/...
.PHONY: openapi-generate-server-code

openapi-lint:
	docker run --rm -v "$(CURDIR):/local" \
		$(DOCKERHUB_PROXY)redocly/cli:$(REDOCLY_CLI_VERSION) \
		lint /local/swagger/_build/swagger.yaml
.PHONY: openapi-lint
