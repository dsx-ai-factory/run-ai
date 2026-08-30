include ../../../build/makefile/control-plane.mk
include ../../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=infra-openapi

# Legacy mount semantics: WORKSPACE_HOST_PATH=${PWD}/../.. expanded at
# depth 3 (control-plane/infra/openapi/) is control-plane/, so /local
# inside the container maps to control-plane/ on host.
DOCKER_WORK_DIR=/local/infra/openapi

# Variables for x-test coverage checking in OpenAPI spec
XTEST_SPEC_FILE=./specs/runai-public-raw.yaml
XTEST_SCRIPT_DIR=./scripts/xtestcoverage
XTEST_SCRIPT_FILE_NAME=check-swagger-test-paths.py

# Variables for OpenAPI pattern checking
PATTERN_CHECKER_SCRIPT_DIR=./scripts/openapi-scanner
PATTERN_CHECKER_SCRIPT=check_patterns.sh
ENABLE_PATTERN_VALIDATION?=true
FAIL_ON_PATTERN_ERRORS?=false

test: ## Run unit-tests.
	# no-op
.PHONY: test

lint:  ## Run linters.
	# no-op
.PHONY: lint

build:
	# no-op
.PHONY: build

validate: validate-openapi-x-test-coverage validate-openapi-patterns
.PHONY: validate

validate-openapi-x-test-coverage:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Checking x-test coverage API" ${BASE_CONSOLE}
	${DOCKER_COMMAND_PRIVILEGED_WORK_DIR} ${DOCKER_PYTHON_IMG} bash -c 'pip install -r ${XTEST_SCRIPT_DIR}/requirements.txt && python ${XTEST_SCRIPT_DIR}/${XTEST_SCRIPT_FILE_NAME} ${XTEST_SPEC_FILE}' || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: validate-openapi-x-test-coverage

validate-openapi-patterns:
ifeq ($(ENABLE_PATTERN_VALIDATION),true)
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Checking poetry installation" ${BASE_CONSOLE}
	@ which poetry > /dev/null 2>&1 || (${ECHO_COMMAND} ${YELLOW_CONSOLE} "${CONSOLE_PREFIX} Poetry not found, installing..." ${BASE_CONSOLE} && pip3 install poetry)
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Installing poetry dependencies" ${BASE_CONSOLE}
	cd ${PATTERN_CHECKER_SCRIPT_DIR} && poetry install
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Checking OpenAPI pattern changes" ${BASE_CONSOLE}
ifeq ($(FAIL_ON_PATTERN_ERRORS),true)
	cd ${PATTERN_CHECKER_SCRIPT_DIR} && bash ${PATTERN_CHECKER_SCRIPT} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
else
	cd ${PATTERN_CHECKER_SCRIPT_DIR} && bash ${PATTERN_CHECKER_SCRIPT} || (${ECHO_COMMAND} ${YELLOW_CONSOLE} "${CONSOLE_PREFIX} Pattern validation failed but continuing (FAIL_ON_PATTERN_ERRORS is not true)" ${BASE_CONSOLE} && true)
	${SUCCESS_MESSAGE_HANDLER}
endif
else
	@ ${ECHO_COMMAND} ${YELLOW_CONSOLE} "${CONSOLE_PREFIX} Skipping OpenAPI pattern validation (ENABLE_PATTERN_VALIDATION is not true)" ${BASE_CONSOLE}
endif
.PHONY: validate-openapi-patterns
