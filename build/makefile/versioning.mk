# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
BLUE=\033[0;34m
NC=\033[0m

TAG_FORMAT?=${SHARED_PACKAGES_PATH_ABS}/${OPENAPI_CLIENT_NAME_GO}
# Constants
MAX_PATCH:=100

RUNAI_REPO_ROOT := $(shell git rev-parse --show-toplevel)
VERSIONING_HELPERS := $(RUNAI_REPO_ROOT)/build/scripts/versioning-helpers.sh
_CHECK_BC_PKG_DIR = $(RUNAI_REPO_ROOT)/$(TAG_FORMAT)
_VERSIONING_ENV = TAG_FORMAT="$(TAG_FORMAT)" MAX_PATCH="$(MAX_PATCH)"
_SOURCE_VERSIONING_HELPERS = . "$(VERSIONING_HELPERS)"

# versioning-helpers.sh uses bash-only syntax (<<<, read -p), so these
# recipes run explicitly under bash instead of make's default $(SHELL)
# (/bin/sh, which is dash on Debian/Ubuntu and would fail on it).
define _run_versioning
@bash -c 'set -e; $(_VERSIONING_ENV); $(_SOURCE_VERSIONING_HELPERS); $(1)'
endef

_require-client:
	@if [ ! -f "$(CURDIR)/pkg/client/go.mod" ] || [ "$(_CHECK_BC_PKG_DIR)" != "$(CURDIR)/pkg/client" ]; then \
		printf "$(RED)Error: client version bumping is only supported from a service with a pkg/client module.$(NC)\n"; \
		printf "$(YELLOW)TAG_FORMAT='$(TAG_FORMAT)' is not this service's pkg/client.$(NC)\n"; \
		exit 1; \
	fi
.PHONY: _require-client

_require-package:
	@case "$(TAG_FORMAT)" in \
		runai-common-packages/*|runai-application-packages/*) ;; \
		*) \
			printf "$(RED)Error: package version bumping is only supported for runai-common-packages and runai-application-packages.$(NC)\n"; \
			printf "$(YELLOW)TAG_FORMAT='$(TAG_FORMAT)' is outside that scope.$(NC)\n"; \
			exit 1 ;; \
	esac
.PHONY: _require-package

client-bump-version: _require-client _check-can-tag
	$(call _run_versioning,versioning_bump prod "$(_CHECK_BC_PKG_DIR)")
.PHONY: client-bump-version

client-bump-version-dev: _require-client
	$(call _run_versioning,versioning_bump dev "$(_CHECK_BC_PKG_DIR)")
.PHONY: client-bump-version-dev

package-bump-version: _require-package _check-can-tag
	$(call _run_versioning,versioning_bump prod "$(_CHECK_BC_PKG_DIR)")
.PHONY: package-bump-version

package-bump-version-dev: _require-package
	$(call _run_versioning,versioning_bump dev "$(_CHECK_BC_PKG_DIR)")
.PHONY: package-bump-version-dev

check-breaking-changes: _require-package
	@$(RUNAI_REPO_ROOT)/build/scripts/check-breaking-changes.sh "$(TAG_FORMAT)" "$(_CHECK_BC_PKG_DIR)"
.PHONY: check-breaking-changes

_validate-main-sync:
	@current_branch=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$current_branch" != "main" ]; then \
		echo "$(RED)Error: You must be on branch 'main' (currently on '$$current_branch')$(NC)"; \
		exit 42; \
	fi; \
	git fetch origin main > /dev/null 2>&1; \
	local_head=$$(git rev-parse main); \
	remote_head=$$(git rev-parse origin/main); \
	if [ "$$local_head" = "$$remote_head" ]; then \
		echo "$(GREEN)On branch 'main' and up to date with origin/main.$(NC)"; \
	else \
		echo "$(RED)Error: Local 'main' is not up to date with 'origin/main'.$(NC)"; \
		exit 42; \
	fi

_check-can-tag:
	@set -e;
	@$(MAKE) _validate-main-sync >/dev/null 2>&1; \
	EXIT_CODE=$$?; \
	if [ "$$EXIT_CODE" -eq 0 ]; then \
		printf "$(GREEN)Branch is up to date with origin/main. You can tag a production release.$(NC)\n"; \
		exit 0; \
	fi; \
   printf "$(RED)Branch is not up to date with origin/main. Make sure you are on main branch to tag a prod release.$(NC)\n"; \
   printf "$(YELLOW)Use the dev bump target (client-bump-version-dev / package-bump-version-dev)$(NC) to tag a dev release from your local branch.\n"; \
   printf "$(BLUE)Or checkout and pull main, then run this command again to tag a prod release.$(NC)\n"; \
	exit 42;

validate: validate-imports
.PHONY: validate

# Exemptions are matched by validate-replaces.sh against the WHOLE module path,
# so they are listed as bare module suffixes and fully qualified below. A bare
# suffix would otherwise also exempt e.g. .../pkg/client-fork.
VALIDATE_REPLACE_MODULE_PREFIX := github\.com/dsx-ai-factory/run-ai/

VALIDATE_IMPORTS_REPLACE_EXCEPTIONS = control-plane/authorization/pkg/client|control-plane/cluster-service/pkg/client|control-plane/org-unit-service/pkg/client|control-plane/workloads/pkg/client

VALIDATE_IMPORTS_REPLACE_ALLOWED_MODULES = control-plane/$(SERVICE_NAME)/pkg/client|$(VALIDATE_IMPORTS_REPLACE_EXCEPTIONS)

# Qualify every alternative: prefix the first, then each one following a `|`.
VALIDATE_IMPORTS_REPLACE_ALLOWED = $(VALIDATE_REPLACE_MODULE_PREFIX)$(subst |,|$(VALIDATE_REPLACE_MODULE_PREFIX),$(VALIDATE_IMPORTS_REPLACE_ALLOWED_MODULES))

VALIDATE_REPLACES_SCRIPT := $(BUILD_MAKEFILE_DIR)scripts/validate-replaces.sh

validate-imports:
	@set -e; \
	MODS=$$(find . -name 'go.mod'); \
	if [ -z "$$MODS" ]; then exit 0; fi; \
	if ! OUT=$$($(VALIDATE_REPLACES_SCRIPT) \
		'go.mod must not contain a replace for a client of another service' \
		'control-plane/[^/]+/pkg/client' \
		'$(VALIDATE_IMPORTS_REPLACE_ALLOWED)' \
		any $$MODS 2>&1); then \
		echo "$(RED)$$OUT$(NC)"; \
		exit 1; \
	fi; \
	if grep -qE '.*control-plane/.*/pkg/client v[0-9]+\.[0-9]+\.[0-9]+-dev.*' $$MODS; then \
		echo "$(RED)Error: go.mod must not require a runai-clients -dev version.$(NC)"; \
		echo "$(YELLOW)$$(grep -E '.*control-plane/.*/pkg/client v[0-9]+\.[0-9]+\.[0-9]+-dev.*' $$MODS)$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)Client import validation passed$(NC)";
.PHONY: validate-imports
