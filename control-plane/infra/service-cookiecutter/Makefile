include ../../../build/makefile/control-plane.mk
include ../../../build/makefile/index.mk

BASE_DIR=../../
TMP_DIR=tmp
SERVICE_NAME=example-service
# Legacy mount semantics: at depth 3 (control-plane/infra/service-cookiecutter/),
# WORKSPACE_HOST_PATH=${PWD}/../.. resolves to control-plane/, so /local
# inside the container is control-plane/ on host.
DOCKER_WORK_DIR=/local/infra/service-cookiecutter

test: ## Run test.
	# no op
.PHONY: test

lint: ## Run linter.
	# no op
.PHONY: lint

purify: ## Purify the code.
	# no op
.PHONY: purify

build: ## build the code.
	# no op
.PHONY: build

generate:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_COOKIECUTTER_COMMAND} ../service-cookiecutter --replay --config-file config.yaml -f -o ${BASE_DIR} || ${FAILURE_MESSAGE_HANDLER}
.PHONY: generate
