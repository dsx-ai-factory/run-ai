include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
# Helm chart Makefile — no SERVICE_NAME convention. Recipes assume the
# container workdir is /local/runai-cluster (one level above cluster/),
# matching the legacy behavior when ${SUB_DIR} and ${SERVICE_NAME} were
# both empty.
DOCKER_WORK_DIR=/local/runai-cluster
# Makefile for packaging, building, testing, linting, and validating a Helm chart

# Set the name of your Helm chart
CHART_NAME := runai-cluster

# Set the version of your Helm chart
VERSION ?= 0.0.0-dev

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

.PHONY: build test lint validate package
# Build the Helm chart 
build:
	@echo "Building Helm chart: $(CHART_NAME)"
	yq -i e ".version=\"$(VERSION)\"" Chart.yaml
	yq -i e ".runai-operator.image.tag=\"$(VERSION)\"" values.yaml
	yq -i e ".runai-operator.image.registry=\"$(REGISTRY_URL)/$(REGISTRY)\"" values.yaml
	yq -i e ".crdUpgrader.image.registry=\"$(REGISTRY_URL)/$(REGISTRY)\"" values.yaml
	yq -i e ".crdUpgrader.image.tag=\"$(VERSION)\"" values.yaml
	yq -i e ".helmHooksHandler.image.registry=\"$(REGISTRY_URL)/$(REGISTRY)\"" values.yaml
	yq -i e ".helmHooksHandler.image.tag=\"$(VERSION)\"" values.yaml
	yq -i e ".kartaMigrate.image.registry=\"$(REGISTRY_URL)/$(REGISTRY)\"" values.yaml
	yq -i e ".kartaMigrate.image.tag=\"$(VERSION)\"" values.yaml
	rm -f $(CHART_NAME).tgz
	helm dep up	$(CHART_DIR)
	helm template $(CHART_DIR) > manifests.yaml
	helm package $(CHART_DIR) --destination $(OUTPUT_DIR) --app-version $(VERSION) --version $(VERSION)
	mv $(OUTPUT_DIR)$(CHART_NAME)*.tgz  $(CHART_NAME).tgz

# Run tests for the Helm chart
test: build
	@echo "Running tests for Helm chart: $(CHART_NAME)"
	docker run -t --rm -v $(CHART_DIR):/apps helmunittest/helm-unittest:3.17.2-0.8.1 . -f 'tests/**/*_test.yaml'

# Lint the Helm chart
lint: build
	@echo "Linting Helm chart: $(CHART_NAME)"
	helm lint $(CHART_DIR)

# Validate the Helm chart
validate: build
	@echo "Validating Helm chart: $(CHART_NAME)"
	helm template $(CHART_DIR)
	helm template $(CHART_DIR) -f values.yaml -f ../../../common/opa-policies/kubernetes/runai-cluster-values.yaml --set openShift.mode=Enabled > manifests.yaml
	$(DOCKER_COMMAND) openpolicyagent/conftest:v0.59.0 test -o github -p /local/common/opa-policies/kubernetes cluster/runai-chart/manifests.yaml

build-airgapped:
	## todo

# Make Extra
extra:
	yq -i e ".version=\"$(VERSION)\"" Chart.yaml
	helm dep update	$(CHART_DIR)

# Run all the steps: lint, validate, build, and package
