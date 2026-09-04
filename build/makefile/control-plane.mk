# CP-specific defaults for control-plane services.
#
# Include this file BEFORE build/makefile/index.mk in any control-plane
# service Makefile. The service must set SERVICE_NAME anywhere in the
# file — DOCKER_WORK_DIR uses recursive expansion so the value is
# resolved at recipe time.
#
# Typical CP service Makefile:
#   SERVICE_NAME=my-service
#   include ../../build/makefile/control-plane.mk
#   include ../../build/makefile/index.mk

# Image registry: nvcr.io/nvidia/runai by default; CI overrides
# DOCKER_REPO_BASE directly for prod releases (no ENVIRONMENT variable
# in CP convention — unlike RC, where ENVIRONMENT is part of the
# registry path).
DOCKER_REPO_BASE?=nvcr.io/nvidia/runai

# Host workspace path mounted as /local. CP services are at depth 2 from
# the monorepo root (control-plane/<svc>/), so the workspace mount needs
# two ../ to reach the root.
WORKSPACE_HOST_PATH?=${PWD}/../..

# Container workdir, resolved lazily so SERVICE_NAME can be set anywhere
# in the consuming Makefile.
DOCKER_WORK_DIR?=/local/control-plane/${SERVICE_NAME}
