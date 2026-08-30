include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk
# Makefile for packaging, building, testing, linting, and validating a Helm chart
# Set the name of your Helm chart
CHART_NAME := control-plane

# Set the version of your Helm chart
VERSION ?= 0.0.0-devel

# Set the containers registry
REGISTRY ?= "nvidia/runai"

REGISTRY_URL ?= "nvcr.io"
# Set the directory where your Helm chart files are located
CHART_DIR := ./

OUTPUT_DIR ?= ./

ifeq ($(UNAME_S),Darwin)
SED_INPLACE=sed -i ''
else
SED_INPLACE=sed -i
endif

.PHONY: build test test-secrets lint validate package

# Build the Helm chart 
build:
	@echo "Building Helm chart: $(CHART_NAME)"
	yq -i e ".global.image.registry=\"$(REGISTRY_URL)/$(REGISTRY)\"" values.yaml
	yq -i e ".global.image.tag=\"$(VERSION)\"" values.yaml
	yq -i e ".notificationsService.image.registry=\"$(REGISTRY_URL)/$(REGISTRY)\"" values.yaml
	yq -i e ".notificationsService.image.tag=\"$(VERSION)\"" values.yaml
	yq -i e ".telemetryService.image.registry=\"$(REGISTRY_URL)/$(REGISTRY)\"" values.yaml
	yq -i e ".telemetryService.image.tag=\"$(VERSION)\"" values.yaml
	${SED_INPLACE} "s/ubi:latest/ubi:$(VERSION)/g" values.yaml
	rm -f $(CHART_NAME).tgz
	helm dep up	$(CHART_DIR)
	helm package $(CHART_DIR) --destination $(OUTPUT_DIR) --app-version $(VERSION) --version $(VERSION)
	mv  $(OUTPUT_DIR)$(CHART_NAME)*.tgz $(CHART_NAME).tgz

# Run tests for the Helm chart
test: build test-secrets
	@echo "Running tests for Helm chart: $(CHART_NAME)"
	docker run -t --rm -v $(CHART_DIR):/apps $(DOCKERHUB_PROXY)helmunittest/helm-unittest:3.17.2-0.8.1 .

# Cross-document integrity check: every control-plane service's secretRef must
# resolve to a (secret, key) pair declared in the rendered output, evaluated
# under several values scenarios. Backed by tests/integrity (Go + terratest).
test-secrets:
	@echo "Running secret-reference integrity tests across scenarios"
	cd tests/integrity && go test -count=1 ./...

# Lint the Helm chart
lint: build
	@echo "Linting Helm chart: $(CHART_NAME)"
	helm lint $(CHART_DIR)

# Validate the Helm chart
validate: build
	@echo "Validating Helm chart: $(CHART_NAME)"
	helm template --set global.config.kubernetesDistribution=openshift $(CHART_DIR)
	helm template --set global.config.kubernetesDistribution=openshift $(CHART_DIR) > manifests.yaml
	$(DOCKER_COMMAND) $(DOCKERHUB_PROXY)openpolicyagent/conftest:v0.59.0 test -o github -p /local/common/opa-policies/kubernetes control-plane-chart/manifests.yaml

extra:
	helm dep update	$(CHART_DIR)

# Run all the steps: lint, validate, build, and package
