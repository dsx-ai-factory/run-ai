# Change for every new service
SERVICE_NAME=e2e

# Override Docker workdir and mount for non-control-plane path
DOCKER_WORK_DIR=/local/e2e
DOCKER_COMMAND=docker run --rm -w /local/e2e -v "${PWD}/../:/local" -u $(shell id -u):$(shell id -g)

REPORT_DIR=$(CURDIR)/reports

NODEJS_VERSION=22-bullseye
DOCKER_NODEJS_COMMAND=${DOCKER_COMMAND} -e NODE_OPTIONS="--max-old-space-size=4096" -e npm_config_cache=/tmp/.npmcache -e CYPRESS_CACHE_FOLDER=/tmp/.cypresscache node:${NODEJS_VERSION}

##@ Basic Commands 

test: typecheck ## Run lightweight tests (type-check only for e2e).
.PHONY: test

lint: ## Run linter and type-check.
	${DOCKER_NODEJS_COMMAND} /bin/sh -c 'npm i --save-dev && mkdir -p reports && ./node_modules/.bin/eslint . --ignore-path .eslintignore --quiet --format json --output-file reports/linter-result.json --ext .ts && npm run typecheck'
.PHONY: lint

validate: ## Run validate (no-op for e2e).
	# no-op
.PHONY: validate

build: ## Detect and cache changed playwright test files.
	@ echo "Detecting changed test files..."
	@ mkdir -p $(REPORT_DIR)
	@ BASE_REF=$${BASE_REF:-$${GITHUB_BASE_REF:-release-3.0}}; \
	  HEAD_SHA=$${GITHUB_SHA:-$$(git rev-parse HEAD)}; \
	  echo "Base ref: $$BASE_REF"; \
	  echo "Head sha: $$HEAD_SHA"; \
	  git -C .. fetch origin $$BASE_REF || true; \
	  # Playwright parallel tests
	  git -C .. diff --name-only origin/$$BASE_REF...$$HEAD_SHA -- \
	    e2e/src/test/suites/playwright/parallel/**/* \
	    | grep -v '^$$' > "$(REPORT_DIR)/playwright_parallel_modified.txt" || true; \
	  # Playwright serial tests (excluding phoenix)
	  git -C .. diff --name-only origin/$$BASE_REF...$$HEAD_SHA -- \
	    e2e/src/test/suites/playwright/serial/**/* \
	    ':(exclude)e2e/src/test/suites/playwright/serial/phoenix/**' \
	    | grep -v '^$$' > "$(REPORT_DIR)/playwright_serial_modified.txt" || true; \
	  # Convert to JSON arrays for caching/consumption
	  PW_PAR_JSON=$$(cat "$(REPORT_DIR)/playwright_parallel_modified.txt" 2>/dev/null | jq -R -s -c 'split("\n")[:-1]' 2>/dev/null || echo "[]"); \
	  PW_SER_JSON=$$(cat "$(REPORT_DIR)/playwright_serial_modified.txt" 2>/dev/null | jq -R -s -c 'split("\n")[:-1]' 2>/dev/null || echo "[]"); \
	  echo "$$PW_PAR_JSON" > "$(REPORT_DIR)/modified-playwright-parallel-tests.json"; \
	  echo "$$PW_SER_JSON" > "$(REPORT_DIR)/modified-playwright-serial-tests.json"; \
	  jq -n --argjson pw_parallel "$$PW_PAR_JSON" \
	        --argjson pw_serial "$$PW_SER_JSON" \
	        '{modified_pw_parallel: $pw_parallel, modified_pw_serial: $pw_serial}' \
	        > "$(REPORT_DIR)/changed-tests-summary.json"; \
	  echo "Changed tests cached under $(REPORT_DIR)"
.PHONY: build

typecheck:
	${DOCKER_NODEJS_COMMAND} /bin/sh -c 'npm i --save-dev && npm run typecheck'
.PHONY: typecheck

extra: ## Extra steps (no-op for e2e).
	# no-op
