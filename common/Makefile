SHELL := /bin/bash
REPO_ROOT := $(shell git rev-parse --show-toplevel)
COMMON_DIR := $(REPO_ROOT)/common

build:
	@echo "NA"

test:
	@echo "NA"

lint:
	@echo "NA"

validate:
	@echo "Validating Rego files under common/..."
	@if ! command -v opa >/dev/null 2>&1; then \
		echo "OPA not found, installing..."; \
		OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
		ARCH=$$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/'); \
		curl -sSL -o /tmp/opa "https://openpolicyagent.org/downloads/latest/opa_$${OS}_$${ARCH}_static" && \
		chmod +x /tmp/opa && \
		sudo mv /tmp/opa /usr/local/bin/opa || { echo "Error: OPA installation failed"; exit 1; }; \
	fi; \
	failed=false; \
	while IFS= read -r file; do \
		rel="$${file#$(REPO_ROOT)/}"; \
		echo "Validating Rego: $$rel"; \
		if ! opa check --v0-compatible "$$file" 2>&1; then \
			echo "Error: $$rel - Rego syntax check failed"; \
			failed=true; \
		fi; \
	done < <(find "$(COMMON_DIR)" -name "*.rego" | sort); \
	if [ "$$failed" = "true" ]; then \
		echo "Validation failed!"; \
		exit 1; \
	fi; \
	echo "Validation completed successfully"
