include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

SERVICE_NAME=workload-common
##@ Basic Commands

test: ## Run unit-tests.
	@echo "Running tests..."
	 ${DOCKER_GO_COMMAND} \
	 /bin/sh -c \
	 'for name in *; do \
	 if [ ! -d "$${name}" ] || [ "$${name}" = "vendor" ]; then continue; fi; \
	 echo "Testing $${name}"; \
	 cd $${name} && \
	 go test -coverprofile ./cover.out -v ./... ; \
	 cd ../; \
	 done'
	@echo "Tests completed."
.PHONY: test

lint: ## Run linter.
	# no op
.PHONY: lint

validate: ## Run validate.
	# no-op
.PHONY: validate

build: ## Run builds.
	@echo "Running builds..."
	 ${DOCKER_GO_COMMAND} \
	 /bin/sh -c \
	 'for name in *; do \
	 if [ ! -d "$${name}" ] || [ "$${name}" = "vendor" ]; then continue; fi; \
	 echo "Building $${name}"; \
	 cd $${name} && \
	 go build ; \
	 cd ../; \
	 done'
	@echo "Builds completed."
.PHONY: build
