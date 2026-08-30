# To use the make file please run 'make help' to see the available commands.
include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

PACKAGE=github.com/run-ai/runai/control-plane/runai-cli
CURRENT_DIR=$(shell pwd)
RUNAI_CLI_NAME=runai
OS_ARCH?=linux-amd64
REPO_DIR=runai-cli
SERVICE_NAME=runai-cli

RUNAI_CLI_BIN_PATH=${CURRENT_DIR}/bin
RUNAI_CLI_DOCS_PATH=${CURRENT_DIR}/docs
CLI_DOCS_DIR=./docs
BIN_DIR=${CURRENT_DIR}/bin
DIST_DIR=${BIN_DIR}/dist
# Use .env file if exists - for local development
-include version

# Directory for generated mocks (can be overridden, e.g., MOCKS_DIR=temp-mocks)
MOCKS_DIR ?= internal/autogenerate/mocks
# Path to mockgen binary
MOCKGEN_BIN := $(shell go env GOPATH)/bin/mockgen

BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=${VERSION}
CLI_VERSION=${RUNAI_CLI_MAJOR}.${RUNAI_CLI_MINOR}.${RUNAI_CLI_PATCH}

ifeq (${GIT_TAG},)
    GIT_TAG=0.0.1-local
    CLI_VERSION=${RUNAI_CLI_MAJOR}.${RUNAI_CLI_MINOR}.${RUNAI_CLI_PATCH}-local
endif

override LDFLAGS += \
  -X ${PACKAGE}/pkg/version.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}/pkg/version.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}/pkg/version.controlPlane=${GIT_TAG} \
  -X ${PACKAGE}/pkg/version.version=${CLI_VERSION} \
  -extldflags "-static" \
  -s \
  -w

GENERAL_BUILD_OPTIONS=CGO_ENABLED=0 GO111MODULE=on
CPU_AVAILABLE="0"

# Generate code once before builds
.PHONY: generate
generate:
	go generate

# Build the project
ifeq ($(OS),Windows_NT)
else
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
$(info "Building on Linux")
CPU_AVAILABLE:=$(shell nproc)
else ifeq ($(UNAME_S),Darwin)
$(info "Building on Darwin")
CPU_AVAILABLE:=$(shell sysctl -n hw.physicalcpu)
else
$(error "The OS is not supported")
endif
endif

.PHONY: cli-linux-amd64
cli-linux-amd64:
	${GENERAL_BUILD_OPTIONS} GOOS=linux GOARCH=amd64 GOFIPS140=v1.0.0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/linux/amd64/${RUNAI_CLI_NAME}

.PHONY: cli-linux-arm64
cli-linux-arm64:
	${GENERAL_BUILD_OPTIONS} GOOS=linux GOARCH=arm64 GOFIPS140=v1.0.0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/linux/arm64/${RUNAI_CLI_NAME}

.PHONY: cli-darwin-amd64
cli-darwin-amd64:
	${GENERAL_BUILD_OPTIONS} GOOS=darwin GOARCH=amd64 GOFIPS140=v1.0.0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/darwin/amd64/${RUNAI_CLI_NAME}

.PHONY: cli-darwin-arm64
cli-darwin-arm64:
	${GENERAL_BUILD_OPTIONS} GOOS=darwin GOARCH=arm64 GOFIPS140=v1.0.0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/darwin/arm64/${RUNAI_CLI_NAME}

.PHONY: cli-windows-arm64
cli-windows-arm64:
	${GENERAL_BUILD_OPTIONS} GOARCH=arm64 GOOS=windows GOFIPS140=v1.0.0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/windows/arm64/${RUNAI_CLI_NAME}

.PHONY: cli-windows-amd64
cli-windows-amd64:
	${GENERAL_BUILD_OPTIONS} GOARCH=amd64 GOOS=windows GOFIPS140=v1.0.0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/windows/amd64/${RUNAI_CLI_NAME}

.PHONY: test-local
test-local: git-config-command
	GO111MODULE=on ${GENERAL_BUILD_OPTIONS} go test ./... -v

.PHONY: test
test: test-go

.PHONY: build-docs-markdown
build-docs-markdown: git-config-command
	go build -o ${BIN_DIR}/docgen github.com/run-ai/runai/control-plane/runai-cli/cmd/docgen
	${BIN_DIR}/docgen markdown -o $(CLI_DOCS_DIR)

.PHONY: lint
lint: lint-go ## Lint the code.

##@ OpenAPI

openapi-lint: openapi-lint-generic ## Lint the swagger bundle.
.PHONY: openapi-lint

.PHONY: clean
clean:
	rm -rfv ${BIN_DIR}/

.PHONY: build-all-platforms
build-all-platforms: cli-linux-amd64 cli-linux-arm64 cli-darwin-arm64 cli-darwin-amd64 cli-windows-amd64 cli-windows-arm64

.PHONY: build-run
build-run:
	@mkdir -p ${DIST_DIR}/linux/amd64 ${DIST_DIR}/linux/arm64 ${DIST_DIR}/darwin/amd64 ${DIST_DIR}/darwin/arm64 ${DIST_DIR}/windows/amd64 ${DIST_DIR}/windows/arm64
	@echo "available CPUs: $(CPU_AVAILABLE)"
	$(MAKE) -j$(CPU_AVAILABLE) build-all-platforms

.PHONY: dist
dist: clean git-config-command generate build-run ## Build all platform binaries into bin/dist (used by cli-exposer and releases).

# No standalone CLI artifact is built or published. cli-exposer is the
# single builder via 'make dist'; CI only tests/lints runai-cli. Run 'make dist'
# for local platform binaries.
.PHONY: build
build: ## No-op; run 'make dist' to produce binaries.
	@echo "runai-cli 'build' is a no-op; run 'make dist' for platform binaries."

.PHONY: lint
lint: lint-go

.PHONY: validate
validate: validate-script

# Define the packages you want to install
PACKAGES := \
    github.com/spf13/cobra-cli@latest \
    github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 \
    github.com/golang/mock/mockgen@v1.6.0 \
    golang.org/x/tools/cmd/goimports@latest

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
	# folder=mocks/controlplane pkg=cpmocks
	$(MOCKGEN_BIN) -source=internal/controlplane/node/service.go -destination=$(MOCKS_DIR)/controlplane/nodeClient.go -package=cpmocks -mock_names=Client=MockNodeClient
	$(MOCKGEN_BIN) -source=internal/controlplane/nodepool/service_nodepool_list.go -destination=$(MOCKS_DIR)/controlplane/nodepoolClient.go -package=cpmocks -mock_names=Client=MockNodepoolClient
	$(MOCKGEN_BIN) -source=internal/controlplane/project/list/service_project_list.go -destination=$(MOCKS_DIR)/controlplane/projectLister.go -package=cpmocks -mock_names=Lister=MockProjectLister
	$(MOCKGEN_BIN) -source=internal/controlplane/project/create/service_project_create.go -destination=$(MOCKS_DIR)/controlplane/projectCreator.go -package=cpmocks -mock_names=Creator=MockProjectCreator
	$(MOCKGEN_BIN) -source=internal/controlplane/cluster/service/service.go -destination=$(MOCKS_DIR)/controlplane/clusterClient.go -package=cpmocks
	$(MOCKGEN_BIN) -source=internal/controlplane/pvc/service_pvc_list.go -destination=$(MOCKS_DIR)/controlplane/pvcListClient.go -package=cpmocks -mock_names=Lister=MockPvcLister
	$(MOCKGEN_BIN) -source=internal/controlplane/department/service/service.go -destination=$(MOCKS_DIR)/controlplane/departmentClient.go -package=cpmocks
	$(MOCKGEN_BIN) -source=internal/controlplane/environment/list/service.go -destination=$(MOCKS_DIR)/controlplane/environmentListerClient.go -package=cpmocks -mock_names=Lister=MockEnvironmentLister
	$(MOCKGEN_BIN) -source=internal/controlplane/environment/describe/service_list.go -destination=$(MOCKS_DIR)/controlplane/environmentDescribeLister.go -package=cpmocks -mock_names=Lister=MockEnvironmentDescribeLister
	$(MOCKGEN_BIN) -source=internal/controlplane/compute/list/service.go -destination=$(MOCKS_DIR)/controlplane/computeListerClient.go -package=cpmocks -mock_names=Lister=MockComputeLister
	$(MOCKGEN_BIN) -source=internal/controlplane/compute/describe/service_list.go -destination=$(MOCKS_DIR)/controlplane/computeDescribeLister.go -package=cpmocks -mock_names=Lister=MockComputeDescribeLister
	$(MOCKGEN_BIN) -source=internal/controlplane/datasource/describe/service_list.go -destination=$(MOCKS_DIR)/controlplane/datasourceDescribeLister.go -package=cpmocks -mock_names=Lister=MockDatasourceDescribeLister
	$(MOCKGEN_BIN) -source=internal/controlplane/datasource/list/service.go -destination=$(MOCKS_DIR)/controlplane/datasourceListerClient.go -package=cpmocks -mock_names=Lister=MockDatasourceLister
	$(MOCKGEN_BIN) -source=internal/controlplane/credential/list/service.go -destination=$(MOCKS_DIR)/controlplane/credentialListerClient.go -package=cpmocks -mock_names=Lister=MockCredentialLister
	$(MOCKGEN_BIN) -source=internal/controlplane/mycredential/list/service.go -destination=$(MOCKS_DIR)/controlplane/myCredentialListerClient.go -package=cpmocks -mock_names=Lister=MockMyCredentialLister
	$(MOCKGEN_BIN) -source=internal/controlplane/datavolume/list/service.go -destination=$(MOCKS_DIR)/controlplane/datavolumeListerClient.go -package=cpmocks -mock_names=Lister=MockDatavolumeLister
	$(MOCKGEN_BIN) -source=internal/controlplane/workloadtype/list/service.go -destination=$(MOCKS_DIR)/controlplane/workloadTypeListerClient.go -package=cpmocks -mock_names=Lister=MockWorkloadTypeLister
	$(MOCKGEN_BIN) -source=internal/controlplane/application/list/service.go -destination=$(MOCKS_DIR)/controlplane/aiApplicationLister.go -package=cpmocks -mock_names=AIApplicationLister=MockAIApplicationLister,AIApplicationDeleter=MockAIApplicationDeleter,AIApplicationCreator=MockAIApplicationCreator

	# folder=mocks/controlplane/template pkg=cptemplatemocks
	$(MOCKGEN_BIN) -source=internal/controlplane/template/describe/service_template_lister.go -destination=$(MOCKS_DIR)/controlplane/template/templateDescribeLister.go -package=cptemplatemocks -mock_names=Lister=MockTemplateDescribeLister
	$(MOCKGEN_BIN) -source=internal/controlplane/template/list/service.go -destination=$(MOCKS_DIR)/controlplane/template/templateLister.go -package=cptemplatemocks -mock_names=TemplateLister=MockTemplateLister
	# folder=mocks/controlplane/workload pkg=cpworkloadmocks
	$(MOCKGEN_BIN) -source=internal/controlplane/workload/batch/service_workload_batch.go -destination=$(MOCKS_DIR)/controlplane/workload/workloadBatcher.go -package=cpworkloadmocks
	$(MOCKGEN_BIN) -source=internal/controlplane/workload/list/service_workload_pagination.go -destination=$(MOCKS_DIR)/controlplane/workload/workloadPaginator.go -package=cpworkloadmocks -mock_names=Lister=MockWorkloadPaginator
	$(MOCKGEN_BIN) -source=internal/controlplane/workload/workload_actions.go -destination=$(MOCKS_DIR)/controlplane/workload/workloadActions.go -package=cpworkloadmocks
	$(MOCKGEN_BIN) -source=internal/controlplane/workload/describe/service.go -destination=$(MOCKS_DIR)/controlplane/workload/describe.go -package=cpworkloadmocks
	$(MOCKGEN_BIN) -source=internal/controlplane/priorities/service_priorities_list.go -destination=$(MOCKS_DIR)/controlplane/priorities/servicePrioritiesList.go -package=cpworkloadmocks

	# folder=mocks/upgrade pkg=upgrademocks
	$(MOCKGEN_BIN) -source=internal/upgrade/download.go -destination=$(MOCKS_DIR)/upgrade/download.go -package=upgrademocks
	$(MOCKGEN_BIN) -source=internal/upgrade/flow.go -destination=$(MOCKS_DIR)/upgrade/flow.go -package=upgrademocks
	$(MOCKGEN_BIN) -source=internal/upgrade/factory.go -destination=$(MOCKS_DIR)/upgrade/factory.go -package=upgrademocks
	$(MOCKGEN_BIN) -source=internal/upgrade/cmd/update.go -destination=$(MOCKS_DIR)/upgrade/flow_check.go -package=upgrademocks UpdateFlowConditioner

	# folder=mocks/auth pkg=authmocks
	$(MOCKGEN_BIN) -source=internal/auth/exchange.go -destination=$(MOCKS_DIR)/auth/exchange.go -package=authmocks
	$(MOCKGEN_BIN) -source=internal/auth/idp.go -destination=$(MOCKS_DIR)/auth/idp.go -package=authmocks
	$(MOCKGEN_BIN) -source=internal/auth/browser.go -destination=$(MOCKS_DIR)/auth/browser.go -package=authmocks
	$(MOCKGEN_BIN) -source=internal/auth/auth.go -destination=$(MOCKS_DIR)/auth/auth_app.go -package=authmocks
	$(MOCKGEN_BIN) -source=internal/auth/token/revoke.go -destination=$(MOCKS_DIR)/auth/revoke.go -package=authmocks
	$(MOCKGEN_BIN) -source=internal/auth/token.go -destination=$(MOCKS_DIR)/auth/token_checker.go -package=authmocks TokenChecker
	$(MOCKGEN_BIN) -source=internal/auth/token/service.go -destination=$(MOCKS_DIR)/auth/service_manager.go -package=authmocks -mock_names=Refresher=MockRefresher,Checker=MockChecker,ReadWriter=MockReadWriter
	$(MOCKGEN_BIN) -source=internal/auth/interface.go -destination=$(MOCKS_DIR)/auth/token_service.go -package=authmocks -mock_names=TokenSerivce=MockTokenService

	# folder=mocks pkg=mocks
	$(MOCKGEN_BIN) -source=internal/websocket/handler.go -destination=$(MOCKS_DIR)/dialer.go -package=mocks
	$(MOCKGEN_BIN) -source=internal/network/conn.go -destination=$(MOCKS_DIR)/conn.go -package=mocks
	$(MOCKGEN_BIN) -source=internal/cluster/status/service.go -destination=$(MOCKS_DIR)/workloadStatus.go -package=mocks
	$(MOCKGEN_BIN) -source=internal/cmd/completion/interface.go -destination=$(MOCKS_DIR)/completion.go -package=mocks
	$(MOCKGEN_BIN) -source=internal/k8s/validation/cluster_validator.go -destination=$(MOCKS_DIR)/cluster_validator.go -package=mocks
	$(MOCKGEN_BIN) -source=internal/diagnostics/filter/namespaces_filter.go -destination=$(MOCKS_DIR)/diagnostics/filter/namespaces_filter.go -package=mocks
	$(MOCKGEN_BIN) -source=internal/diagnostics/collect_logs.go -destination=$(MOCKS_DIR)/diagnostics/collect_logs.go -package=mocks
	$(MOCKGEN_BIN) -source=internal/diagnostics/collector/collector.go -destination=$(MOCKS_DIR)/diagnostics/collector/collector.go -package=mocks

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

.PHONY: validate-script
validate-script:
	@echo "Validating..."
	@sh scripts/validations.sh

BUMP_TYPE=""
bump-version:
	@echo "Bumping version..."
	@if [ -z "$(BUMP_TYPE)" ]; then \
		echo "Error: please provide a version type: 'make bump-version BUMP_TYPE={major|minor|patch}'"; \
		exit 1; \
	fi
	@sh scripts/bump_version.sh $(BUMP_TYPE)

test-dev:
	go test -coverprofile ./cover.out -v ./...
	go tool cover -func=cover.out -o=coverage.out

coverage:
	@echo "Update coverage..."
	@sh scripts/coverage.sh
