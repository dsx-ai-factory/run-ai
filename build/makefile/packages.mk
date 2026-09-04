##@ Packages

# Hook the package checks into the shared `validate` target. Recipes stay with
# each service's own `validate`; these only add prerequisites, so they run for
# every service's `make validate` even when that service overrides the recipe
# with a no-op.
#
# imports/replaces/dev-versions are consumer-side checks (a module must not
# depend on, locally replace, or pin a -dev version of a package), so they run
# for every consumer — including the packages themselves.
validate: validate-package-imports validate-package-replaces validate-package-dev-versions

# Gate on the module path RELATIVE to the repo root (derived from this shared
# makefile's own location), matching only the first path segment. Using the
# absolute $(CURDIR) would false-positive when the checkout itself lives under
# a directory named like a package root (e.g. /tmp/runai-common-packages/...).
PKG_REPO_ROOT := $(abspath $(BUILD_MAKEFILE_DIR)/../..)
PKG_MODULE_REL := $(patsubst $(PKG_REPO_ROOT)/%,%,$(CURDIR))
ifneq ($(filter runai-common-packages runai-application-packages,$(firstword $(subst /, ,$(PKG_MODULE_REL)))),)

# Override golang.mk's CP default (../infra/lint/.golangci.yaml) to the
# .golangci.yaml at each package tree's root. Modules live one level under
# their tree (runai-{common,application}-packages/<module>/), so ../ points
# at the tree root. Each tree owns its own config and can tune linters
# independently.
GOLANG_LINTER_CONFIG_PATH = ../.golangci.yaml
endif

validate-package-imports: ## Ensure services don't import the run-ai/runai-(application|common)-packages modules.
	@set -e; \
	MODS=$$(find . -name 'go.mod'); \
	if [ -z "$$MODS" ]; then exit 0; fi; \
	if [ "$(SERVICE_NAME)" != "notifications-service" ]; then \
		if grep -qE 'github\.com/run-ai/runai-(application|common)-packages' $$MODS; then \
			echo "$(RED)Error: go.mod must not import from run-ai/runai-application-packages or run-ai/runai-common-packages$(NC)"; \
			echo "$(YELLOW)$$(grep -E 'github\.com/run-ai/runai-(application|common)-packages' $$MODS)$(NC)"; \
			exit 1; \
		fi; \
	fi; \
	echo "$(GREEN)Go package import validation passed$(NC)";
.PHONY: validate-package-imports

# A package may locally replace itself (e.g. integration-tests pointing to ../).
# Services leave this empty so all local replaces are rejected. validate-replaces.sh
# matches this against the whole module path, so runai-common-packages/logging does
# not also exempt .../logging-extra.
VALIDATE_PACKAGE_REPLACE_ALLOWED := github.com/dsx-ai-factory/run-ai/runai-common-packages/(auth|safemapcache)
ifneq ($(filter runai-common-packages runai-application-packages,$(firstword $(subst /, ,$(PKG_MODULE_REL)))),)
VALIDATE_PACKAGE_REPLACE_ALLOWED := github\.com/dsx-ai-factory/run-ai/$(PKG_MODULE_REL)
endif

VALIDATE_PACKAGE_REPLACES_SCRIPT := $(BUILD_MAKEFILE_DIR)scripts/validate-replaces.sh

validate-package-replaces: ## Ensure the service's own go.mod doesn't locally replace a common/application package.
	@set -e; \
	MODS=$$(find . -name 'go.mod'); \
	if [ -z "$$MODS" ]; then exit 0; fi; \
	if ! OUT=$$($(VALIDATE_PACKAGE_REPLACES_SCRIPT) \
		'go.mod must not contain a local (filesystem path) replace for a common/application package' \
		'(^|/)runai-(application|common)-packages/[^/]+' \
		'$(VALIDATE_PACKAGE_REPLACE_ALLOWED)' \
		local-only $$MODS 2>&1); then \
		echo "$(RED)$$OUT$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)Go package replace validation passed$(NC)";
.PHONY: validate-package-replaces

validate-package-dev-versions: ## Ensure no go.mod requires a common/application package -dev version.
	@set -e; \
	MODS=$$(find . -name 'go.mod'); \
	if [ -z "$$MODS" ]; then exit 0; fi; \
	if grep -qE '.*runai-(application|common)-packages/.* v[0-9]+\.[0-9]+\.[0-9]+-dev.*' $$MODS; then \
		echo "$(RED)Error: go.mod must not require a runai-(application|common)-packages -dev version.$(NC)"; \
		echo "$(YELLOW)$$(grep -E '.*runai-(application|common)-packages/.* v[0-9]+\.[0-9]+\.[0-9]+-dev.*' $$MODS)$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)Go package dev version validation passed$(NC)";
.PHONY: validate-package-dev-versions
