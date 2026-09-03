include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=sdk
##@ Basic Commands

test: test-go test-prometheus-rules ## Run unit-tests and prometheus rules tests.
.PHONY: test

format: format-go ## Format the code.
.PHONY: format

lint: ## Run linter.
	# no-op
.PHONY: lint

validate: git-config-command generate manifests copy-crds generate-cluster-config-values-schema
	git diff --exit-code || (echo "Please run 'make generate manifests generate-cluster-config-values-schema' and commit the changes" && exit 1)
.PHONY: validate

build: 
	# no-op
.PHONY: build

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=runai-operator crd:allowDangerousTypes=true,generateEmbeddedObjectMeta=true webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	@echo "**********************************************************"
	@echo "*     NOTE!                                              *"
	@echo "*  Whenever you change apis/ structures, make sure to    *"
	@echo "*  copy them to runai-charts repo.                       *"
	@echo "*  run: 'make copy-crds'                                 *"
	@echo "**********************************************************"

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: clients
clients: ## Generate clients.
	hack/update-client.sh

.PHONY: copy-crds
copy-crds:
	hack/copy-crds.sh

.PHONY: generate-cluster-config-values-schema
generate-cluster-config-values-schema: ## Generate values.schema.json for Helm chart based on RunaiConfigSpec
	hack/generate-values-schema.sh

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_VERSION ?= v5.0.0
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_GEN_VERSION ?= v0.17.0
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	{ curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) --output install_kustomize.sh && bash install_kustomize.sh $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); rm install_kustomize.sh; }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

.PHONY: promtool-test-image
promtool-test-image: ## Build the promtool test image.
	docker build -f Dockerfile.promtool-test \
		--build-arg GO_VERSION=$(GO_VERSION) \
		-t promtool-test .

.PHONY: test-prometheus-rules
test-prometheus-rules: promtool-test-image ## Run prometheus rules tests with promtool inside a container.
	${DOCKER_GO_BASE_COMMAND} -e GOOS=linux -e GOARCH=${ARCH} promtool-test go test -tags prometheus_rules ./apis/run/v1/prometheus/definitions/...
