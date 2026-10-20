# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Uses general.mk file from infra/makefile

## Paths
BUILD_DIR=bin
BIN_DIR=$(shell pwd)/${BUILD_DIR}
BUILD_OUT_PATH_AMD=${BUILD_DIR}/${SERVICE_NAME}-amd64
BUILD_OUT_PATH_ARM=${BUILD_DIR}/${SERVICE_NAME}-arm64
BUILD_IN_PATH?=cmd/main.go
GOLANG_LINTER_CONFIG_PATH?=../infra/lint/.golangci.yaml


## Versions
# Derive the Go runtime version from the service's go.mod (the `go X.Y.Z`
# directive). Services without a go.mod (e.g. redoc) fall back to
# GO_VERSION_DEFAULT; services that need to pin a different value (e.g.
# traefik, which has no service-level go.mod) can still override GO_VERSION
# in their own Makefile.
GO_VERSION_DEFAULT?=1.26.8

GO_VERSION_BASE?=$(or $(strip $(if $(wildcard go.mod),$(shell awk '$$1=="go"{print $$2;exit}' go.mod),)), $(GO_VERSION_DEFAULT))
# Use the default tagged version which should be ubuntu-based
GO_VERSION?=$(GO_VERSION_BASE)
GO_ALPINE_VERSION?=$(GO_VERSION_BASE)-alpine
MOCKERY_VERSION=v3.8.1
# Must be new enough to decode the Go release's export data: v2.9.0 fails
# every package on Go 1.27 with "export data version 4 ... (typecheck)".
# renovate: datasource=go depName=github.com/golangci/golangci-lint/v2
GOLANGCI_LINT_VERSION=v2.15.0
GOPRIVATE=github.com/run-ai/*

# Ginkgo runner settings. Override per service if needed.
GINKGO_TIMEOUT?=15m
GINKGO_PROCS?=1
GINKGO_COMPILERS?=1

ifneq ($(strip ${GITHUB_RUN_ID}),)
GINKGO_CONTAINER_NAME?=runai-${GITHUB_RUN_ID}-${SERVICE_NAME}-test
else
GINKGO_LOCAL_RUN_ID:=$(shell printf '%s' $$$$)
GINKGO_CONTAINER_NAME?=runai-local-${GINKGO_LOCAL_RUN_ID}-${SERVICE_NAME}-test
endif
GINKGO_DOCKER_RUN_ARGS=--name=${GINKGO_CONTAINER_NAME}

# Extra flags passed to `go test` (e.g. -short, -count=1, -tags=...).
GO_TEST_FLAGS?=-short

# Package selection for unit tests. Services that keep integration/envtest
# suites in a separate job override this to exclude them, e.g.
#   GO_TEST_PACKAGES=`go list ./... | grep -v pkg/tests`
GO_TEST_PACKAGES?=./...

# Unit-test command. Local default is plain `go test` so developers need no
# extra tooling and see normal output. CI sets GOTESTSUM_ENABLED=1 to wrap the
# same packages/flags in gotestsum and capture per-test results for flakiness
# detection. CI-only on purpose: --rerun-fails would mask real failures locally.
GOTESTSUM_ENABLED?=0
GOTESTSUM_VERSION?=v1.13.0
GOTESTSUM_FLAGS?=--format=pkgname --rerun-fails=1 --jsonfile=test-report.json
# Invoke gotestsum via `go run ...@version` (same pattern as ginkgo) so it needs
# no install step or custom image; it runs inside the stock golang container.
ifeq ($(GOTESTSUM_ENABLED),1)
GO_TEST_COMMAND?=go run gotest.tools/gotestsum@$(GOTESTSUM_VERSION) $(GOTESTSUM_FLAGS) --packages="$(GO_TEST_PACKAGES)" -- $(GO_TEST_FLAGS) -coverprofile ./cover.out
else
GO_TEST_COMMAND?=go test $(GO_TEST_PACKAGES) $(GO_TEST_FLAGS) -coverprofile ./cover.out
endif

# When set to 1, runs `go list -m all` before tests to surface broken
# `replace` directives. Off by default for control-plane; runai-cluster
# enables it.
GO_LIST_ENABLED?=0

INTEG_TESTS_PACKAGE?=./integration-tests/tests/
INTEG_TESTS_BIN?=${BUILD_DIR}/integ-tests
INTEG_TESTS_TAGS?=integration

GOCACHE_DOCKER_DIR=/tmp/.cache
GOCACHE_HOST_DIR=${HOME}/.cache/go-build-docker-gocache
GOPATH_HOST_DIR=${HOME}/.cache/go-build-docker-gopath
DOCKER_GO_CACHING_VOLUME_AND_ENV := -v ${GOPATH_HOST_DIR}:/go:z -v ${GOCACHE_HOST_DIR}:${GOCACHE_DOCKER_DIR}:z -e GOPATH=/go -e GOCACHE=${GOCACHE_DOCKER_DIR} -e GOLANGCI_LINT_CACHE=${GOCACHE_DOCKER_DIR}

### GO
DOCKER_GO_BASE_COMMAND=${DOCKER_COMMAND} -v ${HOME}/.gitconfig://.gitconfig -e GIT_CONFIG_GLOBAL=/.gitconfig -e GOPRIVATE=${GOPRIVATE} -e CGO_ENABLED=0 -e GO111MODULE=on ${DOCKER_GO_CACHING_VOLUME_AND_ENV}
DOCKER_GO_BASE_COMMAND_PRIVILEGED=${DOCKER_COMMAND_PRIVILEGED} -v ${HOME}/.gitconfig:/root/.gitconfig -e GOPRIVATE=${GOPRIVATE} --network=host -e CGO_ENABLED=0 -e GO111MODULE=on ${DOCKER_GO_CACHING_VOLUME_AND_ENV}
DOCKER_GO_COMMAND=${DOCKER_GO_BASE_COMMAND} -e GOOS=${OS} -e GOARCH=${ARCH} ${DOCKERHUB_PROXY}library/golang:${GO_VERSION}
DOCKER_GO_COMMAND_ARM=${DOCKER_GO_BASE_COMMAND} -e GOOS=linux -e GOARCH=arm64 ${DOCKERHUB_PROXY}library/golang:${GO_VERSION}
DOCKER_GO_COMMAND_PRIVILEGED=${DOCKER_GO_BASE_COMMAND_PRIVILEGED} -e GOOS=${OS} -e GOARCH=${ARCH} ${DOCKERHUB_PROXY}library/golang:${GO_VERSION}

DOCKER_GO_COMMAND_LINUX=${DOCKER_GO_BASE_COMMAND} -e GOOS=linux -e GOARCH=${ARCH} ${DOCKERHUB_PROXY}library/golang:${GO_VERSION}

# CGO-enabled command, for services that need to compile against C libraries.
DOCKER_CGO_BASE_COMMAND=${DOCKER_COMMAND} -v ${HOME}/.gitconfig://.gitconfig -e GOPRIVATE=${GOPRIVATE} ${DOCKER_GO_CACHING_VOLUME_AND_ENV}
DOCKER_CGO_COMMAND=${DOCKER_CGO_BASE_COMMAND} -e CGO_ENABLED=1 -e GOOS=${OS} -e GOARCH=${ARCH} ${DOCKERHUB_PROXY}library/golang:${GO_VERSION}

DOCKER_GINKGO_COMMAND=${DOCKER_GO_COMMAND} /bin/sh -c '${GINKGO_COMMAND}'
DOCKER_GINKGO_COMMAND_PRIVILEGED=${DOCKER_GO_COMMAND_PRIVILEGED} /bin/sh -c '${GINKGO_COMMAND}'

DOCKER_GO_LINTER_COMMAND=${DOCKER_GO_BASE_COMMAND} -e GOFLAGS="-buildvcs=false" --entrypoint /bin/sh ${DOCKERHUB_PROXY}library/golang:${GO_VERSION} -c 'go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} &&
DOCKER_GO_IMPORTS_COMMAND=${DOCKER_GO_BASE_COMMAND} ${DOCKERHUB_PROXY}cytopia/goimports:latest-0.3
DOCKER_MOCKERY_COMMAND=${DOCKER_GO_BASE_COMMAND} -e GO111MODULE=on --entrypoint /bin/sh ${DOCKERHUB_PROXY}library/golang:${GO_VERSION} -c 'go install github.com/vektra/mockery/v3@${MOCKERY_VERSION} &&

# CI-only per-test capture for ginkgo suites: emit a machine-readable JSON
# report when GOTESTSUM_ENABLED=1 (same CI knob as the go test path). Empty by
# default so local runs are unchanged.
ifeq ($(GOTESTSUM_ENABLED),1)
GINKGO_CAPTURE_FLAGS?=--json-report=test-report.json
else
GINKGO_CAPTURE_FLAGS?=
endif

GINKGO_COMMAND=go run github.com/onsi/ginkgo/v2/ginkgo -r --procs=${GINKGO_PROCS} --compilers=${GINKGO_COMPILERS} --randomize-all --randomize-suites --fail-on-pending --keep-going --timeout=${GINKGO_TIMEOUT} --trace ${GINKGO_CAPTURE_FLAGS}

GO_ACC_IGNORE?=""
GO_ACC_COMMAND=go install github.com/ory/go-acc@v0.2.8 && \
	go-acc --ignore="${GO_ACC_IGNORE}" -o /tmp/coverage.out ./... && \
	grep -v zz_ < /tmp/coverage.out | grep -v "/generated/pkg" > cover.out

GO_LDFLAGS=''

# Restrict golangci-lint to changes since the PR's base branch
# (--new-from-rev), so a PR is judged on the code it actually touched.
# Override to empty in a service Makefile to lint every file.
LINT_BASE_BRANCH ?= $(or $(GITHUB_BASE_REF),$(GITHUB_REF_NAME),main)
ifeq ($(GITHUB_REF_TYPE),tag)
  # GITHUB_REF_NAME is the tag itself here, so --new-from-rev would diff the
  # tagged commit against itself and report nothing. Lint the whole tree.
  LINT_REV ?= $(LINT_BASE_BRANCH)
  LINT_NEW_FROM_REV ?=
else
  LINT_REV ?= origin/$(LINT_BASE_BRANCH)
  LINT_NEW_FROM_REV ?= $(LINT_REV)
endif

################### OS DEPENDENCY ##########################
# Build the project

# Used to test on CI, linux refer to the CI runner. 
# Test containers uses container infra that is not supported in CI runner
ifeq ($(UNAME_S),Linux)
test-ginkgo: test-ginkgo-privileged
test-go-acc: 
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Skipping go acc tests in CI" ${BASE_CONSOLE}

else
test-ginkgo: test-ginkgo-local
test-go-acc: test-go-acc-local

endif
############################################################

### Targets

gocache:
	mkdir -p ${GOCACHE_HOST_DIR}
	mkdir -p ${GOPATH_HOST_DIR}
.PHONY: gocache

test-go: git-config-command gocache go-list
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running unit-tests for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_GO_COMMAND_LINUX} ${GO_TEST_COMMAND} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: test-go

go-list:
	@if [ "$(GO_LIST_ENABLED)" = "1" ]; then \
		${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running 'go list -m all' to validate go mod replace for ${SERVICE_NAME}" ${BASE_CONSOLE}; \
		${DOCKER_GO_COMMAND} go list -mod=mod -m all || ${FAILURE_MESSAGE_HANDLER}; \
		${SUCCESS_MESSAGE_HANDLER}; \
	else \
		echo "Skipping go list -m all (GO_LIST_ENABLED=$(GO_LIST_ENABLED))"; \
	fi
.PHONY: go-list

test-ginkgo-local: git-config-command gocache go-list
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running gikngo unit-tests for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${GINKGO_COMMAND} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: test-ginkgo-local

test-ginkgo-privileged: DOCKER_RUN_ARGS=${GINKGO_DOCKER_RUN_ARGS}
test-ginkgo-privileged: git-config-command gocache go-list
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running gikngo unit-tests for ${SERVICE_NAME}" ${BASE_CONSOLE}
	@ ${ECHO_COMMAND} ${YELLOW_CONSOLE} "${CONSOLE_PREFIX} Using privileged container, use with caution!" ${BASE_CONSOLE}
	@ remove_ginkgo_container() { \
		container_ids=$$(docker container ls --all --quiet --filter "name=^/${GINKGO_CONTAINER_NAME}$$") || return $$?; \
		[ -z "$$container_ids" ] || docker rm -f "${GINKGO_CONTAINER_NAME}" >/dev/null; \
	}; \
	teardown_ginkgo_run() { \
		trap - EXIT INT TERM; \
		remove_ginkgo_container; \
	}; \
	stop_ginkgo_client() { \
		if [ -n "$${docker_pid:-}" ]; then \
			kill -TERM "$$docker_pid" >/dev/null 2>&1 || true; \
			wait "$$docker_pid" >/dev/null 2>&1 || true; \
		fi; \
	}; \
	handle_signal() { \
		teardown_ginkgo_run; \
		stop_ginkgo_client; \
		exit "$$1"; \
	}; \
	remove_ginkgo_container; \
	cleanup_status=$$?; \
	if [ "$$cleanup_status" -ne 0 ]; then \
		${FAILURE_MESSAGE_HANDLER}; \
		exit 1; \
	fi; \
	trap 'teardown_ginkgo_run' EXIT; \
	trap 'handle_signal 130' INT; \
	trap 'handle_signal 143' TERM; \
	${DOCKER_GINKGO_COMMAND_PRIVILEGED} & \
	docker_pid=$$!; \
	wait "$$docker_pid"; \
	test_status=$$?; \
	teardown_ginkgo_run; \
	cleanup_status=$$?; \
	if [ "$$test_status" -ne 0 ] || [ "$$cleanup_status" -ne 0 ]; then \
		${FAILURE_MESSAGE_HANDLER}; \
		exit 1; \
	fi
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: test-ginkgo-privileged

test-go-acc-local: git-config-command gocache
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running go-acc coverage for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${GO_ACC_COMMAND} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: test-go-acc-local

# Regenerates the service mocks (.mockery.yml) and the client mocks
# (pkg/client/.mockery.yml), then fails if either left the working tree dirty.
# Each generator is a no-op when its .mockery.yml is absent, so services
# without a pkg/client need no extra target.
validate-mockery: git-config-command
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating mocks are up to date for ${SERVICE_NAME}" ${BASE_CONSOLE}
	@set -e; \
	mocks_state() { \
		git status --porcelain --untracked-files=all; \
		git diff HEAD; \
		git ls-files --others --exclude-standard | while IFS= read -r f; do git hash-object -- "$$f"; done; \
	}; \
	elapsed() { echo "$$(( $$(date +%s) - $$1 ))"; }; \
	started=$$(date +%s); \
	before=$$(mocks_state); \
	step=$$(date +%s); \
	$(MAKE) generate-mocks; \
	${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} generate-mocks took $$(elapsed $$step)s" ${BASE_CONSOLE}; \
	step=$$(date +%s); \
	$(MAKE) generate-mocks-client; \
	${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} generate-mocks-client took $$(elapsed $$step)s" ${BASE_CONSOLE}; \
	after=$$(mocks_state); \
	if [ "$$before" != "$$after" ]; then \
		${ECHO_COMMAND} ${RED_CONSOLE} "${CONSOLE_PREFIX} mockery produced uncommitted changes in ${SERVICE_NAME}. Run 'make generate-mocks generate-mocks-client' and commit them." ${BASE_CONSOLE}; \
		git --no-pager status --porcelain --untracked-files=all; \
		git --no-pager diff; \
		${FAILURE_MESSAGE_HANDLER}; \
	fi; \
	${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} validate-mockery took $$(elapsed $$started)s for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: validate-mockery

lint-go: git-config-command gocache
	@ echo $(GREEN_CONSOLE) "${CONSOLE_PREFIX} Running golangci linter for ${SERVICE_NAME}" $(BASE_CONSOLE); \
		if [ -n "$(LINT_NEW_FROM_REV)" ]; then git fetch origin $(LINT_BASE_BRANCH) --depth=1 2>/dev/null || true; fi; \
		$(DOCKER_GO_LINTER_COMMAND) \
		golangci-lint run -v -c $(GOLANG_LINTER_CONFIG_PATH) $(if $(LINT_NEW_FROM_REV),--new-from-rev=$(LINT_NEW_FROM_REV),)' || $(FAILURE_MESSAGE_HANDLER)
.PHONY: lint-go

.PHONY: build-go
build-go: build-go-amd build-go-arm

build-go-amd: git-config-command format-go gocache
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Building ${SERVICE_NAME}, GOOS: ${OS}, GOARCH ${ARCH}" ${BASE_CONSOLE}
	${DOCKER_GO_COMMAND} env GOFIPS140=v1.0.0 go build -ldflags ${GO_LDFLAGS} -o ${BUILD_OUT_PATH_AMD} ${BUILD_IN_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: build-go

build-go-arm: git-config-command format-go gocache
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Building ${SERVICE_NAME}, GOOS: linux, GOARCH arm64" ${BASE_CONSOLE}
	${DOCKER_GO_COMMAND_ARM} env GOFIPS140=v1.0.0 go build -ldflags ${GO_LDFLAGS} -o ${BUILD_OUT_PATH_ARM} ${BUILD_IN_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: build-go

format-go: git-config-command gocache
	${DOCKER_GO_COMMAND} \
	env GOFLAGS=-mod=mod go fmt ./...
.PHONY: format-go

purify-go: git-config-command gocache
	${DOCKER_GO_COMMAND} \
	go fmt ./... \
	&& go vet ./... \
	&& go mod tidy
.PHONY: purify-go

build-integration-tests-generic: git-config-command gocache
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Building integration-tests binary for ${SERVICE_NAME}" ${BASE_CONSOLE}
	@ mkdir -p $(dir ${INTEG_TESTS_BIN})
	${DOCKER_GO_COMMAND_LINUX} go test -c -tags ${INTEG_TESTS_TAGS} ${INTEG_TESTS_PACKAGE} -o ${INTEG_TESTS_BIN} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: build-integration-tests-generic

generate-mocks: git-config-command gocache
	@if [ ! -f .mockery.yml ] && [ ! -f .mockery.yaml ]; then \
		${ECHO_COMMAND} ${YELLOW_CONSOLE} "${CONSOLE_PREFIX} No .mockery.yml in ${SERVICE_NAME}; nothing to generate" ${BASE_CONSOLE}; \
	else \
		${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating mocks for ${SERVICE_NAME}" ${BASE_CONSOLE}; \
		${DOCKER_MOCKERY_COMMAND} mockery' || ${FAILURE_MESSAGE_HANDLER}; \
		${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Stamping SPDX headers on generated mocks" ${BASE_CONSOLE}; \
		sh ${BUILD_MAKEFILE_DIR}scripts/spdx-header-go.sh .; \
	fi
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: generate-mocks

generate-mocks-client: git-config-command gocache
	@if [ ! -d pkg/client ] || ([ ! -f pkg/client/.mockery.yml ] && [ ! -f pkg/client/.mockery.yaml ]); then \
		${ECHO_COMMAND} ${YELLOW_CONSOLE} "${CONSOLE_PREFIX} No pkg/client/.mockery.yml in ${SERVICE_NAME}; nothing to generate" ${BASE_CONSOLE}; \
	else \
		${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating mocks for ${SERVICE_NAME} pkg/client" ${BASE_CONSOLE}; \
		${DOCKER_MOCKERY_COMMAND} cd pkg/client && mockery' || ${FAILURE_MESSAGE_HANDLER}; \
		${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Stamping SPDX headers on generated mocks" ${BASE_CONSOLE}; \
		sh ${BUILD_MAKEFILE_DIR}scripts/spdx-header-go.sh pkg/client; \
	fi
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: generate-mocks-client

# Same as generate-mocks, without Docker. The Renovate post-upgrade hook runs
# inside the Renovate container, which has no Docker socket, so the containerised
# target fails there with "Cannot connect to the Docker daemon". No GOTOOLCHAIN
# override: the toolchain has to follow the module, or mockery is built by one Go
# and shells out to the `go list` of another.
generate-mocks-local: git-config-command
	@if [ ! -f .mockery.yml ] && [ ! -f .mockery.yaml ]; then \
		${ECHO_COMMAND} ${YELLOW_CONSOLE} "${CONSOLE_PREFIX} No .mockery.yml in ${SERVICE_NAME}; nothing to generate" ${BASE_CONSOLE}; \
		exit 0; \
	fi
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating mocks for ${SERVICE_NAME} (no container)" ${BASE_CONSOLE}
	@mkdir -p ${BIN_DIR}
	GOBIN=${BIN_DIR} go install github.com/vektra/mockery/v3@${MOCKERY_VERSION}
	${BIN_DIR}/mockery
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Stamping SPDX headers on generated mocks" ${BASE_CONSOLE}
	sh ${BUILD_MAKEFILE_DIR}scripts/spdx-header-go.sh .
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: generate-mocks-local
