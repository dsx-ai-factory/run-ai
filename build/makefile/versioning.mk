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

_require-package:
	@case "$(TAG_FORMAT)" in \
		runai-common-packages/*|runai-application-packages/*) ;; \
		*) \
			printf "$(RED)Error: package version bumping is only supported for runai-common-packages and runai-application-packages.$(NC)\n"; \
			printf "$(YELLOW)TAG_FORMAT='$(TAG_FORMAT)' is outside that scope.$(NC)\n"; \
			exit 1 ;; \
	esac
.PHONY: _require-package

client-bump-version: _check-can-tag _bump-version
.PHONY: client-bump-version

client-bump-version-dev: _bump-version-dev
.PHONY: client-bump-version-dev

# versioning-helpers.sh uses bash-only syntax (<<<, read -p), so these two
# recipes run explicitly under bash instead of make's default $(SHELL)
# (/bin/sh, which is dash on Debian/Ubuntu and would fail on it).
package-bump-version: _require-package _check-can-tag
	@bash -c ' \
	set -e; \
	$(_VERSIONING_ENV); \
	$(_SOURCE_VERSIONING_HELPERS); \
	versioning_refresh_tags; \
	latest_prod_tag=$$(versioning_latest_prod_tag); \
	echo "Latest prod tag: $$latest_prod_tag"; \
	set +e; \
	$(RUNAI_REPO_ROOT)/build/scripts/check-breaking-changes.sh "$(TAG_FORMAT)" "$(_CHECK_BC_PKG_DIR)" "$$latest_prod_tag"; \
	RESULT=$$?; \
	set -e; \
	bump_kind=$$(versioning_bump_kind_from_result $$RESULT); \
	versioning_publish_prod_bump "$$bump_kind" "$$latest_prod_tag"; \
	'
.PHONY: package-bump-version

package-bump-version-dev: _require-package
	@bash -c ' \
	set -e; \
	$(_VERSIONING_ENV); \
	$(_SOURCE_VERSIONING_HELPERS); \
	versioning_refresh_tags; \
	latest_prod_tag=$$(versioning_latest_prod_tag); \
	latest_dev_tag=$$(versioning_latest_dev_tag); \
	echo "Latest prod tag: $$latest_prod_tag"; \
	echo "Latest dev tag: $$latest_dev_tag"; \
	set +e; \
	$(RUNAI_REPO_ROOT)/build/scripts/check-breaking-changes.sh "$(TAG_FORMAT)" "$(_CHECK_BC_PKG_DIR)" "$$latest_prod_tag"; \
	RESULT=$$?; \
	set -e; \
	bump_kind=$$(versioning_bump_kind_from_result $$RESULT); \
	target_version=$$(versioning_next_dev_target_version "$$bump_kind" "$$latest_prod_tag" "$$latest_dev_tag"); \
	new_tag=$$(versioning_next_dev_tag "$$target_version"); \
	versioning_create_tag "$$new_tag" "Dev release $$new_tag"; \
	'
.PHONY: package-bump-version-dev

check-breaking-changes: _require-package
	@$(RUNAI_REPO_ROOT)/build/scripts/check-breaking-changes.sh "$(TAG_FORMAT)" "$(_CHECK_BC_PKG_DIR)"
.PHONY: check-breaking-changes

_bump-version-dev:
	@set -e; \
	git tag -l '$(TAG_FORMAT)*' | xargs -r git tag -d > /dev/null 2>&1; \
	git fetch --tags --force origin > /dev/null 2>&1; \
	latest_tag=$$(git tag -l '$(TAG_FORMAT)/v*' | sort -V | tail -1); \
	echo "Latest tag: $$latest_tag"; \
	if [ -z "$$latest_tag" ]; then \
		new_tag="$(TAG_FORMAT)/v0.0.0-dev.0"; \
	elif echo "$$latest_tag" | grep -Eq '^$(TAG_FORMAT)/v[0-9]+\.[0-9]+\.[0-9]+-dev\.[0-9]+$$'; then \
		base=$$(echo $$latest_tag | sed -E 's#(.*-dev\.)([0-9]+)$$#\1#'); \
		dev_ver=$$(echo $$latest_tag | sed -E 's#.*-dev\.([0-9]+)$$#\1#'); \
		new_dev_ver=$$((dev_ver + 1)); \
		new_tag="$$base$$new_dev_ver"; \
	else \
		ver=$$(echo $$latest_tag | sed -E 's#^$(TAG_FORMAT)/v([0-9]+)\.([0-9]+)\.([0-9]+).*#\1.\2.\3#'); \
		major=$$(echo $$ver | cut -d. -f1); \
		minor=$$(echo $$ver | cut -d. -f2); \
		patch=$$(echo $$ver | cut -d. -f3); \
		new_patch=$$((patch + 1)); \
		if [ "$$new_patch" -ge $(MAX_PATCH) ]; then \
			new_patch=0; \
			minor=$$((minor + 1)); \
		fi; \
		new_tag="$(TAG_FORMAT)/v$${major}.$${minor}.$${new_patch}-dev.0"; \
	fi; \
	git tag $$new_tag -m "Dev release $$new_tag"; \
	git push origin $$new_tag > /dev/null 2>&1; \
	echo "$(GREEN)Published tag: $$new_tag$(NC)";
.PHONY: _bump-version-dev

_bump-version:
	@set -e; \
	git tag -l '*$(TAG_FORMAT)*' | xargs -r git tag -d > /dev/null 2>&1; \
	git fetch --tags --force origin > /dev/null 2>&1; \
	echo '$(TAG_FORMAT)/v*'; \
	latest_dev_tag=$$(git tag -l '$(TAG_FORMAT)/v*-dev*' | sort -V | tail -1); \
	latest_prod_tag=$$(git tag -l '$(TAG_FORMAT)/v*' | sort -V | grep -v dev | tail -1); \
	echo "Latest dev tag: $$latest_dev_tag"; \
	echo "Latest prod tag: $$latest_prod_tag"; \
	prod_version=$$(echo $$latest_prod_tag | sed -E 's#^$(TAG_FORMAT)/v([0-9]+\.[0-9]+\.[0-9]+).*#\1#'); \
	if [ -z "$$latest_dev_tag" ]; then \
	  dev_version="0.0.0"; \
	else \
	  dev_version=$$(echo $$latest_dev_tag | sed -E 's#^$(TAG_FORMAT)/v([0-9]+\.[0-9]+\.[0-9]+)-dev\.[0-9]+#\1#'); \
	fi; \
	prod_major=$$(echo $$prod_version | cut -d. -f1); \
	prod_minor=$$(echo $$prod_version | cut -d. -f2); \
	prod_patch=$$(echo $$prod_version | cut -d. -f3); \
	dev_major=$$(echo $$dev_version | cut -d. -f1); \
	dev_minor=$$(echo $$dev_version | cut -d. -f2); \
	dev_patch=$$(echo $$dev_version | cut -d. -f3); \
	if [ "$$dev_major" -gt "$$prod_major" ] || \
	   ([ "$$dev_major" -eq "$$prod_major" ] && [ "$$dev_minor" -gt "$$prod_minor" ]) || \
	   ([ "$$dev_major" -eq "$$prod_major" ] && [ "$$dev_minor" -eq "$$prod_minor" ] && [ "$$dev_patch" -gt "$$prod_patch" ]); then \
		cand_major=$$dev_major; \
		cand_minor=$$dev_minor; \
		cand_patch=$$dev_patch; \
	else \
		cand_major=$$prod_major; \
		cand_minor=$$prod_minor; \
		cand_patch=$$((prod_patch + 1)); \
	fi; \
	if [ "$$cand_patch" -ge $(MAX_PATCH) ]; then \
		cand_patch=0; \
		cand_minor=$$((cand_minor + 1)); \
	fi; \
	new_tag="$(TAG_FORMAT)/v$${cand_major}.$${cand_minor}.$${cand_patch}"; \
	git tag $$new_tag -m "Prod release $$new_tag"; \
	git push origin $$new_tag > /dev/null 2>&1; \
	echo "$(GREEN)Published tag: $$new_tag$(NC)";
.PHONY: _bump-version

_validate-master-sync:
	@current_branch=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$current_branch" != "master" ]; then \
		echo "$(RED)Error: You must be on branch 'master' (currently on '$$current_branch')$(NC)"; \
		exit 42; \
	fi; \
	git fetch origin master > /dev/null 2>&1; \
	local_head=$$(git rev-parse master); \
	remote_head=$$(git rev-parse origin/master); \
	if [ "$$local_head" = "$$remote_head" ]; then \
		echo "$(GREEN)On branch 'master' and up to date with origin/master.$(NC)"; \
	else \
		echo "$(RED)Error: Local 'master' is not up to date with 'origin/master'.$(NC)"; \
		exit 42; \
	fi

_check-can-tag:
	@set -e;
	@$(MAKE) _validate-master-sync >/dev/null 2>&1; \
	EXIT_CODE=$$?; \
	if [ "$$EXIT_CODE" -eq 0 ]; then \
		printf "$(GREEN)Branch is up to date with origin/master. You can tag a production release.$(NC)\n"; \
		exit 0; \
	fi; \
   printf "$(RED)Branch is not up to date with origin/master. Make sure you are on master branch to tag a prod release.$(NC)\n"; \
   printf "$(YELLOW)Use the dev bump target (client-bump-version-dev / package-bump-version-dev)$(NC) to tag a dev release from your local branch.\n"; \
   printf "$(BLUE)Or checkout and pull master, then run this command again to tag a prod release.$(NC)\n"; \
	exit 42;

validate: validate-imports
.PHONY: validate

# Exemptions are matched by validate-replaces.sh against the WHOLE module path,
# so they are listed as bare module suffixes and fully qualified below. A bare
# suffix would otherwise also exempt e.g. .../pkg/client-fork.
VALIDATE_REPLACE_MODULE_PREFIX := github\.com/run-ai/runai/

VALIDATE_IMPORTS_REPLACE_EXCEPTIONS = control-plane/authorization/pkg/client|control-plane/cluster-service/pkg/client

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
