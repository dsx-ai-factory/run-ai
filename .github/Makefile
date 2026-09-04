SHELL := /bin/bash
REPO_ROOT := $(shell git rev-parse --show-toplevel)
GITHUB_DIR := $(REPO_ROOT)/.github
SCAN_DIR ?= $(GITHUB_DIR)

build:
	@echo "NA"

test:
	@echo "NA"

lint:
	@echo "NA"

validate:
	@echo "Validating all files under $(SCAN_DIR)..."
	@failed=false; \
	while IFS= read -r file; do \
		basename="$$(basename "$$file")"; \
		rel="$${file#$(REPO_ROOT)/}"; \
		ext="$${basename##*.}"; \
		if [[ "$$ext" == "$$basename" ]] || [[ "$$basename" == .* ]]; then \
			continue; \
		fi; \
		ext_lower=$$(echo "$$ext" | tr '[:upper:]' '[:lower:]'); \
		case "$$ext_lower" in \
			pyc|lock|txt|disabled) \
				continue \
				;; \
			py) \
				echo "Validating Python: $$rel"; \
				if ! python3 -m py_compile "$$file" 2>&1; then \
					echo "Error: $$rel - Python syntax check failed"; \
					failed=true; \
				fi \
				;; \
			sh) \
				echo "Validating Bash: $$rel"; \
				if ! bash -n "$$file" 2>&1; then \
					echo "Error: $$rel - Bash syntax check failed"; \
					failed=true; \
				fi \
				;; \
			yml|yaml) \
				echo "Validating YAML: $$rel"; \
				if ! yq eval '.' "$$file" >/dev/null 2>&1; then \
					echo "Error: $$rel - YAML syntax check failed"; \
					failed=true; \
				fi \
				;; \
			json) \
				echo "Validating JSON: $$rel"; \
				if ! python3 -m json.tool "$$file" >/dev/null 2>&1; then \
					echo "Error: $$rel - JSON syntax check failed"; \
					failed=true; \
				fi \
				;; \
			md) \
				echo "Validating Markdown: $$rel"; \
				if ! python3 -c "import sys; open(sys.argv[1], encoding='utf-8').read()" "$$file" 2>&1; then \
					echo "Error: $$rel - Markdown file is not valid UTF-8"; \
					failed=true; \
				fi \
				;; \
			*) \
				echo "Error: Unsupported file type '.$$ext' in $$rel"; \
				failed=true \
				;; \
		esac; \
	done < <(find "$(SCAN_DIR)" -type f | sort); \
	if [ "$$failed" = "true" ]; then \
		echo "Validation failed!"; \
		exit 1; \
	fi; \
	echo "Validation completed successfully"
