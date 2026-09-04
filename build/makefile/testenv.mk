# Uses base.mk and golang.mk files from common/build/makefile

GO_ENVTEST_INNER_COMMAND=go test ./pkg/tests/... --ginkgo.fail-fast -timeout 30m

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.33
ENVTEST_VERSION=release-0.23

define GO_ENVTEST_CREATE_CLUSTER_COMMAND
$(KIND) delete cluster && \
$(KIND) create cluster && \
kubectl create -f ../runai-crd-upgrader/crds
endef

GO_ENVTEST_RUN_TEST_COMMAND=USE_EXISTING_CLUSTER=true \
    KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path --arch amd64)" \
    ${GO_ENVTEST_INNER_COMMAND}

define GO_ENVTEST_FULL_COMMAND
${GO_ENVTEST_CREATE_CLUSTER_COMMAND} && \
${GO_ENVTEST_RUN_TEST_COMMAND}
endef

define GO_ENVTEST_FULL_COMMAND_FOR_DOCKER
go install sigs.k8s.io/kind@v0.14.0 && \
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@${ENVTEST_VERSION} && \
cp /go/bin/linux_amd64/kind bin/kind && \
cp /go/bin/linux_amd64/setup-envtest bin/setup-envtest && \
${GO_ENVTEST_FULL_COMMAND}
endef
DOCKER_GO_ENVTEST_FULL_COMMAND=${DOCKER_GO_BASE_COMMAND_PRIVILEGED} -e GOOS=linux -e GOARCH=amd64 golang:${GO_VERSION} /bin/sh -c '${GO_ENVTEST_FULL_COMMAND_FOR_DOCKER}'

### Targets

# test-env: test-go
# 	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running kind env tests for ${SERVICE_NAME}" ${BASE_CONSOLE}
# 	${DOCKER_GO_ENVTEST_FULL_COMMAND} || ${FAILURE_MESSAGE_HANDLER}
# 	${SUCCESS_MESSAGE_HANDLER}
# .PHONY: test-env

envtest-docker-go: git-config-command  gocache
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running unit-tests for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_GO_COMMAND_LINUX_PRIVILEGED} make envtest-go || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}

envtest-go: envtest gocache
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running unit-tests for ${SERVICE_NAME}" ${BASE_CONSOLE}
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path --arch amd64)" \
    go test ./... -timeout 30m || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}

test-env-local: envtest kind test-go
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running kind env tests for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${GO_ENVTEST_FULL_COMMAND} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: test-env-local


####### TOOLS ########

KIND = $(BUILD_DIR)/kind
.PHONY: kind
kind: ## Download kind locally if necessary.
	GOBIN=$(BIN_DIR) go install sigs.k8s.io/kind@v0.20.0

ENVTEST = $(BUILD_DIR)/setup-envtest
.PHONY: envtest
envtest: ## Download envtest locally if necessary.
	GOBIN=$(BIN_DIR) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@${ENVTEST_VERSION}

HELM = $(BUILD_DIR)/helm
.PHONY: helm
helm: ## Download helm locally if necessary.
	GOBIN=$(BIN_DIR) go install helm.sh/helm/v3/cmd/helm@v3.10.0
