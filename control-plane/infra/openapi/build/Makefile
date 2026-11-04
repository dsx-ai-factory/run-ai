# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../../build/makefile/control-plane.mk
include ../../../../build/makefile/index.mk

SERVICE_NAME=openapi/build

# Mount repo root (4 levels up) so git commands work and GIT_REPO_PATH=/local
# matches every other control-plane service.
WORKSPACE_HOST_PATH=${PWD}/../../../..
DOCKER_WORK_DIR=/local/control-plane/infra/openapi/build

# When openapi.mk targets run from this folder, resolve paths relative to self
OPENAPI_BUILD_PATH=.
# Siblings of build/ under infra/openapi/
OPENAPI_CENTRAL_SCRIPTS_PATH=../scripts
OPENAPI_CENTRAL_SPECS_PATH=../specs

# Spec paths (relative to this folder)
SPEC_FILE_SRC=../specs/runai-public.yaml
SPEC_FILE_SRC_SYSTEM=../specs/runai-system-scoped.yaml
SPEC_FILE_DST=./_build/runai-public.yaml
SPEC_FILE_DST_SYSTEM=./_build/runai-system-scoped.yaml

NODEJS_VERSION=20-alpine3.17

BACKWARD_COMPATIBILITY_SCRIPTS_PATH=${OPENAPI_CENTRAL_SCRIPTS_PATH}/backward-compatibility

##@ Basic Commands

test: ## Run unit-tests.
	# no-op
.PHONY: test

lint: ## Run linters.
	# no-op
.PHONY: lint

validate: validate-openapi-build-uptodate validate-openapi-public-spec validate-tags validate-openapi-spec-python validate-openapi-backwards-compatibility
.PHONY: validate

# Generated files that openapi-build-public-spec overwrites in place.
OPENAPI_DRIFT_FILES=_build/runaiApi.swagger.yaml _build/runai-system-scoped.yaml ${OPENAPI_CENTRAL_SPECS_PATH}/runai-public-raw.yaml ${OPENAPI_CENTRAL_SPECS_PATH}/runai-public.yaml ${OPENAPI_CENTRAL_SPECS_PATH}/runai-system-scoped.yaml

# The build rewrites the committed files, so snapshot them first and compare after.
# Replaces the plain build as the first step of validate; regenerated files are left in place.
validate-openapi-build-uptodate:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Checking the committed merged specs match a fresh build" ${BASE_CONSOLE}
	@ snapshot=$$(mktemp -d) && trap 'rm -rf "$$snapshot"' EXIT && \
	for f in ${OPENAPI_DRIFT_FILES}; do cp "$$f" "$$snapshot/$$(echo "$$f" | tr '/' '_')" || exit 1; done && \
	$(MAKE) openapi-build-public-spec && \
	drift=0 && \
	for f in ${OPENAPI_DRIFT_FILES}; do \
		diff -w -B "$$snapshot/$$(echo "$$f" | tr '/' '_')" "$$f" >/dev/null || { echo "ERROR: $$f differs from a fresh build"; drift=1; }; \
	done && \
	[ "$$drift" = 0 ] || { echo "Run 'make build' in control-plane/infra/openapi/build and commit the regenerated specs"; exit 1; }
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: validate-openapi-build-uptodate

validate-openapi-spec-python:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating Python client code using OpenAPI Generator" ${BASE_CONSOLE}
	${DOCKER_OPENAPI_BASE_COMMAND_SERVER} generate -i ${SPEC_FILE_SRC} -o /tmp -g python >/dev/null 2>&1 || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: validate-openapi-spec-python

validate-tags:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating public openapi spec" ${BASE_CONSOLE}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Checking for x-internal tags" ${BASE_CONSOLE}
	${DOCKER_LINUX_COMMAND} /bin/sh -c "cat ${OPENAPI_BUILD_PATH}/_build/runaiApi.swagger.yaml | grep -i 'x-internal' && exit 1 || exit 0"

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Checking for x-authorization tags" ${BASE_CONSOLE}
	${DOCKER_LINUX_COMMAND} /bin/sh -c "cat ${OPENAPI_BUILD_PATH}/_build/runaiApi.swagger.yaml | grep -i 'x-authorization' && exit 1 || exit 0"
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: validate-tags

validate-openapi-backwards-compatibility:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating backwards compatibility of public openapi spec" ${BASE_CONSOLE};
	${DOCKER_COMMAND_PRIVILEGED} -e SEVERITY_LEVELS_FILE=${BACKWARD_COMPATIBILITY_SCRIPTS_PATH}/oasdiff-levels.txt -e ERR_IGNORE_FILE=${BACKWARD_COMPATIBILITY_SCRIPTS_PATH}/oasdiff-err-ignore.txt -e SOURCE_FILE_PATH=${SPEC_FILE_SRC} -e BASE_BRANCH_NAME=${GITHUB_BASE_REF} -e GITHUB_TOKEN=${GITHUB_TOKEN} -e GIT_REPO_PATH=/local -e ALLOF_MERGE_SCRIPT_PATH=${BACKWARD_COMPATIBILITY_SCRIPTS_PATH}/allof-merge.js --entrypoint sh ${DOCKERHUB_PROXY}library/golang:${GO_ALPINE_VERSION} -c 'sh ${BACKWARD_COMPATIBILITY_SCRIPTS_PATH}/setup.sh && sh ${BACKWARD_COMPATIBILITY_SCRIPTS_PATH}/validate.sh' || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: validate-openapi-backwards-compatibility

prepare-specs: openapi-build-public-spec
	mkdir -p _build
	cp ${SPEC_FILE_SRC} ${SPEC_FILE_DST}
	cp ${SPEC_FILE_SRC_SYSTEM} ${SPEC_FILE_DST_SYSTEM}
.PHONY: prepare-specs

build: openapi-build-public-spec
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Copying built spec for local use" ${BASE_CONSOLE}
	# Validate source file exists
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating source file exists: ${OPENAPI_PUBLIC_SPEC_BUILD_PATH}" ${BASE_CONSOLE}
	@ test -f "${OPENAPI_PUBLIC_SPEC_BUILD_PATH}" || (echo "${RED_CONSOLE}ERROR: Source file ${OPENAPI_PUBLIC_SPEC_BUILD_PATH} does not exist${BASE_CONSOLE}" && exit 1)
	# Copy built file to output location
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Copying ${OPENAPI_PUBLIC_SPEC_BUILD_PATH} to ${SPEC_FILE_DST}" ${BASE_CONSOLE}
	cp "${OPENAPI_PUBLIC_SPEC_BUILD_PATH}" ${SPEC_FILE_DST} || ${FAILURE_MESSAGE_HANDLER}
	# Validate copy succeeded and files are identical
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating copy succeeded and files are identical" ${BASE_CONSOLE}
	@ diff -q "${OPENAPI_PUBLIC_SPEC_BUILD_PATH}" ${SPEC_FILE_DST} || (echo "${RED_CONSOLE}ERROR: Copy failed - files are not identical${BASE_CONSOLE}" && exit 1)
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: build
