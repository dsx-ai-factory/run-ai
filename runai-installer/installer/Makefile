##@ Basic Commands

SERVICE_NAME=runai-installer
BIN_REL_PATH?=bin
OUTPUT_PATH?=$(BIN_REL_PATH)/$(OS)/$(ARCH)/$(SERVICE_NAME)
export GOPRIVATE=github.com/run-ai/*

GOLANGCI_LINT_VERSION=v2.9.0
DOCKERHUB_PROXY?=

git-config-command:
ifdef GITHUB_TOKEN
	@echo Creating gitconfig with github token
	git config --file $(HOME)/.gitconfig url.https://$(GITHUB_TOKEN)@github.com/.insteadOf https://github.com/
endif
.PHONY: git-config-command

test: git-config-command
	go test ./...
.PHONY: test

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
		$(DOCKERHUB_PROXY)golangci/golangci-lint:$(GOLANGCI_LINT_VERSION)-alpine \
		golangci-lint run -c .golangci.yaml
.PHONY: lint

validate:
	# no-op
.PHONY: validate

purify: git-config-command
	go vet ./...
.PHONY: purify

build: git-config-command
	OS=linux ARCH=amd64 make build-for-platform
	OS=linux ARCH=arm64 make build-for-platform
	OS=darwin ARCH=amd64 make build-for-platform
	OS=darwin ARCH=arm64 make build-for-platform
.PHONY: build

build-for-platform: git-config-command
	mkdir -p $(dir $(OUTPUT_PATH))
	GOOS=$(OS) GOARCH=$(ARCH) GOFIPS140=v1.0.0 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_PATH) ./cmd
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
