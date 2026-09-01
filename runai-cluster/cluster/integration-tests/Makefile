GOPATH ?= ${HOME}/go
GOBIN ?= ${GOPATH}/bin
GINKGO_INSTALLED := $(shell command -v ginkgo 2> /dev/null)
# CI-only per-test capture: emit a JSON report when GOTESTSUM_ENABLED=1 (same
# knob as the unit path) so integration-test flakiness can be tracked. Empty by
# default so local runs are unchanged.
GINKGO_CAPTURE_FLAGS =
ifeq ($(GOTESTSUM_ENABLED),1)
GINKGO_CAPTURE_FLAGS = --json-report=test-report.json
endif
GINKGO_FLAGS = -r --fail-fast --randomize-all --randomize-suites --trace $(GINKGO_CAPTURE_FLAGS)
export GOMAXPROCS = 4

# Install ginkgo if not installed
install-ginkgo:
ifndef GINKGO_INSTALLED
	GOBIN=${GOBIN} go install github.com/onsi/ginkgo/v2/ginkgo
endif
.PHONY: install-ginkgo

# Run the ginkgo tests
run-ginkgo-tests:
	${GOBIN}/ginkgo ${GINKGO_FLAGS} ${TEST_PATH}
.PHONY: run-ginkgo-tests

external-workload-integrator-test: git-config-command install-ginkgo
	$(MAKE) TEST_PATH=./suites/external-workload-integrator run-ginkgo-tests
.PHONY: external-workload-integrator-test

project-controller-test: git-config-command install-ginkgo
	$(MAKE) TEST_PATH=./suites/project-controller run-ginkgo-tests
.PHONY: project-controller-test

nodepool-controller-test: git-config-command install-ginkgo
	$(MAKE) TEST_PATH=./suites/nodepool-controller run-ginkgo-tests
.PHONY: nodepool-controller-test

pod-group-assigner-test: git-config-command install-ginkgo
	$(MAKE) TEST_PATH=./suites/pod-group-assigner run-ginkgo-tests
.PHONY: pod-group-assigner-test

git-config-command:
ifdef GITHUB_TOKEN
	echo Creating gitconfig with github token
	git config --file ${HOME}/.gitconfig url.https://${GITHUB_TOKEN}@github.com/.insteadOf https://github.com/
endif
.PHONY: git-config-command
