# RC-specific defaults for runai-cluster services.
#
# Include this file BEFORE build/makefile/index.mk in any runai-cluster
# service Makefile. Each service then picks its container workdir by
# assigning DOCKER_WORK_DIR to one of the three pre-built values below.
#
# Typical RC service Makefile:
#   SERVICE_NAME=my-service
#   include ../../../build/makefile/runai-cluster.mk
#   include ../../../build/makefile/index.mk
#   DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}   # or _ENGINE or _COMMON

# Image registry: nvcr.io/nvidia/runai by default. CI overrides
# DOCKER_REPO_BASE for release builds; ENVIRONMENT is left in place
# for downstream Makefiles that still branch on it.
ENVIRONMENT?=lab
DOCKER_REPO_BASE?=nvcr.io/nvidia/runai

# Host workspace path mounted as /local. RC services are at depth 3 from
# the monorepo root (runai-cluster/{cluster,engine,common}/<svc>/), so
# the workspace mount needs three ../ to reach the root.
WORKSPACE_HOST_PATH?=${PWD}/../../..

# golangci-lint config path: RC services share one lint config under
# runai-cluster/common/build/lint/, three levels up from a depth-3 service.
# Overrides build/makefile/golang.mk's CP default (../infra/lint/.golangci.yaml).
GOLANG_LINTER_CONFIG_PATH?=../../common/build/lint/.golangci.yaml

# Enable legacy RC --new-from-rev linting: golangci-lint only flags code
# changed since the base branch. CP services lint all files; RC services
# preserve the legacy "new code only" filter to avoid surfacing pre-existing
# issues unrelated to a given PR. Override to empty in a service Makefile
# to lint everything.
LINT_NEW_FROM_REV ?= $(LINT_REV)

# Pre-built workdirs for the three RC service categories. Each service
# sets DOCKER_WORK_DIR to one of these — see the example header above.
# Recursive (`=`) so SERVICE_NAME can be set anywhere in the consuming
# Makefile.
DOCKER_WORK_DIR_CLUSTER=/local/runai-cluster/cluster/${SERVICE_NAME}
DOCKER_WORK_DIR_COMMON=/local/runai-cluster/common/${SERVICE_NAME}
