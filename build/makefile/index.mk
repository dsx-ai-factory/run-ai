# Shared build infrastructure for the runai monorepo.
#
# Services include this file via:
#   include <relative-path>/build/makefile/index.mk
#
# This index is location-independent: it computes its own directory from
# MAKEFILE_LIST so each sub-include resolves correctly regardless of how
# deep the consuming service sits in the tree.

BUILD_MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

include $(BUILD_MAKEFILE_DIR)base.mk
include $(BUILD_MAKEFILE_DIR)golang.mk
include $(BUILD_MAKEFILE_DIR)openapi.mk
include $(BUILD_MAKEFILE_DIR)javascript.mk
include $(BUILD_MAKEFILE_DIR)versioning.mk
include $(BUILD_MAKEFILE_DIR)testenv.mk
include $(BUILD_MAKEFILE_DIR)packages.mk
