# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

## Colored outputs

RED_CONSOLE="\033[0;31m"
GREEN_CONSOLE="\033[0;32m"
YELLOW_CONSOLE="\033[0;33m"
BASE_CONSOLE="\033[0m"
CONSOLE_PREFIX="[$(shell date +'%Y-%m-%d %H:%M:%S')]"
FAILURE_MESSAGE_HANDLER=(${ECHO_COMMAND} ${RED_CONSOLE} "${CONSOLE_PREFIX} Failed to run make target $@ for service ${SERVICE_NAME}" ${BASE_CONSOLE}; exit 1)
SUCCESS_MESSAGE_HANDLER=(${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Successfully ran make target $@ for service ${SERVICE_NAME}" ${BASE_CONSOLE})

## Tags

DOCKER_TAG?=0.0.0-dev
DOCKER_REPO_BASE?=nvcr.io/nvidia/runai
DOCKER_REPO_FULL=${DOCKER_REPO_BASE}/${SERVICE_NAME}
DOCKER_IMAGE_NAME=${DOCKER_REPO_FULL}:${DOCKER_TAG}

# Proxy prefix for public Docker Hub images. Default is empty (pull directly
# from Docker Hub); override to a mirror (e.g. DOCKERHUB_PROXY=mirror.example/dockerhub/)
# in environments where direct Docker Hub pulls are rate-limited or blocked.
DOCKERHUB_PROXY?=

# Path to the service's Dockerfile (relative to the service dir). Override
# for services that ship multiple Dockerfiles or use non-standard names.
DOCKERFILE_PATH?=./Dockerfile

# Extra args appended to `docker build` (e.g. --build-arg, --secret).
DOCKER_BUILD_ADDITIONAL_ARGS?=

# Volume mount suffix used as `-v host:container${MOUNT_FLAGS}`. Set to `:z`
# on SELinux-enabled Linux to relabel the volume; empty elsewhere.
MOUNT_FLAGS?=

# Host path mounted into containers as /local. Default works for services
# at depth 2 from the monorepo root (e.g., control-plane/<svc>/). Services
# at deeper paths must override: runai-cluster/cluster/<svc>/ →
WORKSPACE_HOST_PATH?=${PWD}/../..

## Commands

# Service-side workdir inside the container. Default targets control-plane;
# runai-cluster services override to /local/runai-cluster/<sub>/${SERVICE_NAME}.
DOCKER_WORK_DIR?=/local/control-plane/${SERVICE_NAME}
DOCKER_RUN_ARGS?=
DOCKER_COMMAND=docker run --rm ${DOCKER_RUN_ARGS} -w ${DOCKER_WORK_DIR} -v "${WORKSPACE_HOST_PATH}:/local${MOUNT_FLAGS}" -u $(shell id -u):$(shell id -g)
DOCKER_COMMAND_WORK_DIR=docker run --rm ${DOCKER_RUN_ARGS} -w ${DOCKER_WORK_DIR} -v "${WORKSPACE_HOST_PATH}:/local${MOUNT_FLAGS}" -u $(shell id -u):$(shell id -g)
DOCKER_COMMAND_PRIVILEGED=docker run --rm ${DOCKER_RUN_ARGS} -w ${DOCKER_WORK_DIR} -v "${WORKSPACE_HOST_PATH}:/local${MOUNT_FLAGS}" -v ${DOCKER_SOCK_PATH}:${DOCKER_SOCK_PATH} -v ~/.docker/config.json:/root/.docker/config.json:ro -e TESTCONTAINERS_HOST_OVERRIDE=localhost
DOCKER_COMMAND_PRIVILEGED_WORK_DIR=docker run --rm ${DOCKER_RUN_ARGS} -w ${DOCKER_WORK_DIR} -v "${WORKSPACE_HOST_PATH}:/local${MOUNT_FLAGS}" -v ${DOCKER_SOCK_PATH}:${DOCKER_SOCK_PATH} -v ~/.docker/config.json:/root/.docker/config.json:ro -e TESTCONTAINERS_HOST_OVERRIDE=localhost
DOCKER_LINUX_COMMAND=${DOCKER_COMMAND} ${DOCKERHUB_PROXY}library/debian:12.1-slim
DOCKER_CURL_COMMAND=${DOCKER_COMMAND} ${DOCKERHUB_PROXY}alpine/curl:8.12.1
DOCKER_YQ_IMG=${DOCKERHUB_PROXY}mikefarah/yq:4.35.1
DOCKER_YQ_COMMAND=${DOCKER_COMMAND} ${DOCKER_YQ_IMG}
DOCKER_COOKIECUTTER_COMMAND=docker run --rm --platform linux/amd64 -w ${DOCKER_WORK_DIR} -v "${WORKSPACE_HOST_PATH}/../:/local" -u $(shell id -u):$(shell id -g) ${DOCKERHUB_PROXY}datamindedbe/cookiecutter:1.6.0 cookiecutter
DOCKER_PYTHON_IMG=${DOCKERHUB_PROXY}library/python:3.13
DOCKER_PYTHON_COMMAND=${DOCKER_COMMAND_PRIVILEGED} -e SPEC_FILE_SRC=${SPEC_FILE_SRC} ${DOCKER_PYTHON_IMG} bash -c 'pip install -r ${PYTHON_REQ_PATH}/requirements.txt && python ${PYTHON_REQ_PATH}/${PYTHON_SCRIPT_FILE_NAME}'

## Github

GITHUB_PROJECT_NAME=dsx-ai-factory
GITHUB_REPO_NAME=run-ai

## Paths

SHARED_PACKAGES_PATH=../../runai-clients
SHARED_PACKAGES_PATH_ABS=runai-clients
DOCKER_SOCK_PATH=/var/run/docker.sock


################### OS DEPENDENCY ##########################
# Build the project
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(filter arm arm64 aarch64,$(UNAME_M)),)
ARCH?=amd64
else
ARCH?=arm64
endif


ifeq ($(UNAME_S),Linux)
ECHO_COMMAND=echo -e
OS?=linux


else ifeq ($(UNAME_S),Darwin)
ECHO_COMMAND=echo
OS?=darwin

else
$(error "The OS is not supported")
endif

############################################################

### Targets

##@ Basic Commands
git-config-command:
ifdef GITHUB_TOKEN
	echo Creating gitconfig with github token
	git config --file ${HOME}/.gitconfig url.https://${GITHUB_TOKEN}@github.com/.insteadOf https://github.com/
endif
.PHONY: git-config-command

##@ General

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help

docker-build-generic: git-config-command
	docker buildx build ${DOCKER_BUILD_ADDITIONAL_ARGS} --build-arg SERVICE_NAME=${SERVICE_NAME} -f ${DOCKERFILE_PATH} -t ${DOCKER_IMAGE_NAME} --platform=linux/amd64 --attest type=sbom --attest type=provenance,mode=max --load .
.PHONY: docker-build

docker-push-generic:
	docker push ${DOCKER_IMAGE_NAME}
.PHONY: push

# Regenerate this component's NOTICE from its resolved dependency graph.
# Requires go-licenses for Go components: go install github.com/google/go-licenses@latest
BASE_MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

notice:
	@python3 ${BASE_MAKEFILE_DIR}../scripts/generate-notice.py ${CURDIR}
.PHONY: notice
