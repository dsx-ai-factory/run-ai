include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=frontend
NODEJS_VERSION=22-bullseye
ENVIRONMENT ?= prod
OPENAPI_OUT_PATH=${OPENAPI_PUBLIC_SPEC_PATH}
OPENAPI_GENERATOR_CLIENT_CODE_PATH=src/swagger-models/generated
DOCKER_NODEJS_COMMAND=${DOCKER_COMMAND} -e ENVIRONMENT=${ENVIRONMENT} -e NODE_OPTIONS="--max-old-space-size=8192" -e npm_config_cache=/tmp/.npmcache ${DOCKERHUB_PROXY}library/node:${NODEJS_VERSION}
DOCKER_NODEJS_TYPECHECK_THRESHOLD_COMMAND=${DOCKER_NODEJS_COMMAND} /bin/sh -c 'npm i --save-dev && npm run typecheck:threshold' || ${FAILURE_MESSAGE_HANDLER}
DOCKER_NODEJS_LINT_THRESHOLDS_COMMAND=${DOCKER_NODEJS_COMMAND} /bin/sh -c 'npm i --save-dev && npm run lint:thresholds' || ${FAILURE_MESSAGE_HANDLER}
DOCKER_NODEJS_JS_CHECKS_COMMAND=${DOCKER_NODEJS_COMMAND} /bin/sh -c 'npm i --save-dev && npm run typecheck:threshold && npm run lint:thresholds' || ${FAILURE_MESSAGE_HANDLER}

OPENAPI_SPEC_SRC=../infra/openapi/specs/runai-public.yaml
OPENAPI_SPEC_DST=./dist/openapi/runai-api.yaml

##@ Basic Commands

test: test-js ## Run unit-tests.
.PHONY: test

lint: lint-js js-checks ## Run linter.
.PHONY: lint

validate: ## Run validate.
.PHONY: validate

purify: ## Purify the code.
	# no-op
.PHONY: purify

build: build-js copy-openapi-spec ## Build the service.#
.PHONY: build

copy-openapi-spec: ## Copy generated OpenAPI spec into the Docker build context.
	mkdir -p $(dir ${OPENAPI_SPEC_DST})
	cp ${OPENAPI_SPEC_SRC} ${OPENAPI_SPEC_DST}
.PHONY: copy-openapi-spec

extra: 
	if [ "$(ENVIRONMENT)" = "prod" ]; then \
		${DOCKER_NODEJS_COMMAND} /bin/sh -c 'npm i --save-dev && npm run build' || ${FAILURE_MESSAGE_HANDLER} \
	fi

js-checks: ## Run TS type check and ESLint thresholds (a11y, script-setup, security) in a single Docker container.
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running JS checks on service ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_NODEJS_JS_CHECKS_COMMAND}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: js-checks

typecheck-threshold: ## Check TS error count does not exceed threshold.
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running TS type check threshold on service ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_NODEJS_TYPECHECK_THRESHOLD_COMMAND}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: typecheck-threshold

lint-thresholds: ## Check ESLint a11y warning count and Options API file count do not exceed thresholds.
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running ESLint thresholds on service ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_NODEJS_LINT_THRESHOLDS_COMMAND}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: lint-thresholds
