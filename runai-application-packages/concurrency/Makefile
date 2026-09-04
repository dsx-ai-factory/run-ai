include ../../build/makefile/index.mk

SERVICE_NAME=concurrency
TAG_FORMAT=runai-application-packages/$(SERVICE_NAME)
DOCKER_WORK_DIR=/local/runai-application-packages/${SERVICE_NAME}
DOCKER_COMMAND=${DOCKER_COMMAND_WORK_DIR}

test: test-go
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate:
	# no op
.PHONY: validate

build: 
	# no op
.PHONY: build
