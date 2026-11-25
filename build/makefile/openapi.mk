# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Uses general.mk file from infra/makefile
####generators docs:  https://github.com/OpenAPITools/openapi-generator/tree/master/docs/generators

VERSIONED_CLIENT?=false

## Versions

REDOCLY_CLI_VERSION=1.12.0
OPENAPI_GENERATOR_VERSION_SERVER=v7.7.0
OPENAPI_GENERATOR_VERSION_CLIENT_GO=v7.7.0
OPENAPI_GENERATOR_VERSION_CLIENT_TS?=v6.0.1

TEMPLATES_FOLDER_VERSION_SERVER?=${OPENAPI_GENERATOR_VERSION_SERVER}
TEMPLATES_FOLDER_VERSION_CLIENT_GO?=${OPENAPI_GENERATOR_VERSION_CLIENT_GO}
TEMPLATES_FOLDER_VERSION_CLIENT_TS?=${OPENAPI_GENERATOR_VERSION_CLIENT_TS}

## Paths

OPENAPI_OUT_DIR=swagger/_build
OPENAPI_OUT_PATH=${OPENAPI_OUT_DIR}/swagger.yaml
OPENAPI_OUT_NO_INTERNAL=${OPENAPI_OUT_DIR}/swagger.yaml.no_internal
OPENAPI_OUT_PATH_TO_COPY=${OPENAPI_OUT_DIR}/swagger.yaml
OPENAPI_HTML_PATH=${OPENAPI_OUT_DIR}/swagger.html
OPENAPI_JSON_PATH=${OPENAPI_OUT_DIR}/swagger.json
OPENAPI_TMP_OUT_PATH=../tmp/${SERVICE_NAME}.tmp.yaml
OPENAPI_SRC_PATH=swagger/source/openapi.yaml
OPENAPI_GENERATOR_SERVER_CODE_PATH?=internal/generated
OPENAPI_GENERATOR_CLIENT_SWAGGER_PATH?=${OPENAPI_OUT_DIR}/swagger.yaml
OPENAPI_GENERATOR_CLIENT_JS_CODE_PATH=/tmp/generated
OPENAPI_GENERATOR_INTERNAL_CODE_PATH=pkg/api
OPENAPI_GENERATOR_SERVER_TMP_CODE_PATH=../tmp/${SERVICE_NAME}/generated
OPENAPI_GENERATOR_CLIENT_TMP_CODE_PATH=../tmp/${OPENAPI_CLIENT_NAME_GO}/generated

OPENAPI_CENTRAL_SCRIPTS_PATH=../infra/openapi/scripts
OPENAPI_CENTRAL_SPECS_PATH=../infra/openapi/specs
OPENAPI_BUILD_PATH?=../infra/openapi/build
OPENAPI_BUILD_DIR=${OPENAPI_BUILD_PATH}/_build
OPENAPI_SYSTEM_SCOPED_SPEC_BUILD_PATH=${OPENAPI_BUILD_DIR}/runai-system-scoped.yaml
OPENAPI_PUBLIC_SPEC_BUILD_PATH=${OPENAPI_BUILD_DIR}/runaiApi.swagger.yaml
OPENAPI_PUBLIC_CLEANUP_CONFIG_PATH=${OPENAPI_BUILD_PATH}/config/redocly.yaml
OPENAPI_SYSTEM_SCOPED_CLEANUP_CONFIG_PATH=${OPENAPI_BUILD_PATH}/config/redocly-system-scoped.yaml
OPENAPI_PUBLIC_SPEC_PATH=${OPENAPI_CENTRAL_SPECS_PATH}/runai-public.yaml
OPENAPI_GENERATOR_SERVER_TEMPLATES_PATH=../infra/openapi/templates/go-server-${TEMPLATES_FOLDER_VERSION_SERVER}
OPENAPI_GENERATOR_CLIENT_GO_TEMPLATES_PATH=../infra/openapi/templates/go-client-${TEMPLATES_FOLDER_VERSION_CLIENT_GO}
OPENAPI_GENERATOR_CLIENT_TS_TEMPLATES_PATH=../infra/openapi/templates/typescript-axios-${TEMPLATES_FOLDER_VERSION_CLIENT_TS}
OPENAPI_CLIENT_NAME_TS?=${SERVICE_NAME}-client
OPENAPI_CLIENT_NAME_GO?=${SERVICE_NAME}-client
OPENAPI_CLIENT_SUB_FOLDER?=generated
OPENAPI_GENERATOR_CLIENT_CODE_PATH=${SHARED_PACKAGES_PATH}/${OPENAPI_CLIENT_NAME_GO}/${OPENAPI_CLIENT_SUB_FOLDER}
OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH=../frontend/src/swagger-models/${OPENAPI_CLIENT_NAME_TS}

# Configurations
OPENAPI_GENERATOR_DEBUG?=false
OPENAPI_GENERATOR_ADDITIONAL_PROPERTY_ADD_RESPONSE_HEADERS?=false

### OpenAPI

DOCKER_OPENAPI_COMMAND=${DOCKER_COMMAND} ${DOCKERHUB_PROXY}redocly/cli:${REDOCLY_CLI_VERSION}
DOCKER_OPENAPI_COMMAND_WORK_DIR=${DOCKER_COMMAND_WORK_DIR} ${DOCKERHUB_PROXY}redocly/cli:${REDOCLY_CLI_VERSION}
DOCKER_OPENAPI_COMMAND_WORK_DIR_HIGH_MEM=${DOCKER_COMMAND_WORK_DIR} -e NODE_OPTIONS="--max-old-space-size=16384" ${DOCKERHUB_PROXY}redocly/cli:${REDOCLY_CLI_VERSION}

DOCKER_OPENAPI_BASE_COMMAND_CLIENT_GO=${DOCKER_COMMAND_WORK_DIR} -e JAVA_OPTS="-Dlog.level=error " ${DOCKERHUB_PROXY}openapitools/openapi-generator-cli:${OPENAPI_GENERATOR_VERSION_CLIENT_GO}
DOCKER_OPENAPI_BASE_COMMAND_CLIENT_TS=${DOCKER_COMMAND_WORK_DIR} -e JAVA_OPTS="-Dlog.level=error " ${DOCKERHUB_PROXY}openapitools/openapi-generator-cli:${OPENAPI_GENERATOR_VERSION_CLIENT_TS}
DOCKER_OPENAPI_BASE_COMMAND_SERVER=${DOCKER_COMMAND_WORK_DIR} -e JAVA_OPTS="-Dlog.level=error " ${DOCKERHUB_PROXY}openapitools/openapi-generator-cli:${OPENAPI_GENERATOR_VERSION_SERVER}

DOCKER_OPENAPI_SERVER_GENERATOR_COMMAND=${DOCKER_OPENAPI_BASE_COMMAND_SERVER} generate \
	--additional-properties outputAsLibrary=true \
	--additional-properties packageVersion=${GO_VERSION_BASE} \
	--additional-properties gitUserId=run-ai \
	--additional-properties sourceFolder=${OPENAPI_GENERATOR_INTERNAL_CODE_PATH} \
	--additional-properties gitRepoId=${SERVICE_NAME} \
	--additional-properties packageName=api \
	--additional-properties router=chi \
	--additional-properties hideGenerationTimestamp=true \
	--additional-properties enumClassPrefix=true \
	--additional-properties addResponseHeaders=${OPENAPI_GENERATOR_ADDITIONAL_PROPERTY_ADD_RESPONSE_HEADERS} \
	--openapi-normalizer REMOVE_X_INTERNAL=true \
	-t ${OPENAPI_GENERATOR_SERVER_TEMPLATES_PATH} \
	-g go-server
ifeq (${OPENAPI_GENERATOR_DEBUG}, true)
	DOCKER_OPENAPI_SERVER_GENERATOR_COMMAND += --global-property debugOpenAPI,debugModels,debugOperations,debugSupportingFiles
endif

DOCKER_OPENAPI_CLIENT_GO_GENERATOR_COMMAND=${DOCKER_OPENAPI_BASE_COMMAND_CLIENT_GO} generate \
    --git-user-id ${GITHUB_PROJECT_NAME}/${GITHUB_REPO_NAME} \
	--git-repo-id ${SHARED_PACKAGES_PATH_ABS}/${OPENAPI_CLIENT_NAME_GO} \
	--additional-properties generateInterfaces=true,packageVersion=1.17 \
	--additional-properties enumClassPrefix=true \
	--openapi-normalizer REMOVE_X_INTERNAL=true \
	-t ${OPENAPI_GENERATOR_CLIENT_GO_TEMPLATES_PATH} \
	-g go

DOCKER_OPENAPI_CLIENT_JS_GENERATOR_COMMAND=${DOCKER_OPENAPI_BASE_COMMAND_CLIENT_TS} generate \
	-g typescript-axios

### Targets

openapi-bundle-generic: git-config-command
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Bundling openapi files for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_OPENAPI_COMMAND_WORK_DIR} bundle ${OPENAPI_SRC_PATH} -o ${OPENAPI_OUT_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-bundle-generic

openapi-bundle-json-generic: git-config-command
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Bundling openapi files for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_OPENAPI_COMMAND_WORK_DIR} bundle ${OPENAPI_SRC_PATH} -o ${OPENAPI_JSON_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-bundle-json-generic

openapi-html-doc-generic: git-config-command
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Bundling openapi files for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_OPENAPI_COMMAND_WORK_DIR_HIGH_MEM} build-docs -o ${OPENAPI_HTML_PATH} -t /local/control-plane/infra/openapi/templates/build-docs/template.hbs ${OPENAPI_OUT_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-html-doc-generic

openapi-copy-central-spec: git-config-command
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Copying openapi spec for service ${SERVICE_NAME} to ${OPENAPI_CENTRAL_SPECS_PATH}" ${BASE_CONSOLE}
	@ cp ${OPENAPI_OUT_PATH_TO_COPY} ${OPENAPI_CENTRAL_SPECS_PATH}/${SERVICE_NAME}.yaml || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
. PHONY : openapi-bundle-central-spec

openapi-bundle: openapi-bundle-generic openapi-bundle-json-generic openapi-html-doc-generic openapi-copy-central-spec openapi-build-public-spec 
.PHONY: openapi-bundle

openapi-build: openapi-bundle-generic openapi-bundle-json-generic openapi-html-doc-generic openapi-copy-central-spec openapi-build-public-spec openapi-generate-code 
.PHONY: openapi-build

openapi-lint-generic: git-config-command
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Linting openapi spec for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_OPENAPI_COMMAND} lint ${OPENAPI_OUT_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-lint

openapi-generate-code-server-go: git-config-command openapi-code-server-go-compose openapi-code-server-go-format
.PHONY: openapi-generate-code-server-go

openapi-code-server-go-compose: git-config-command
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating openapi server code for ${SERVICE_NAME} to ${OPENAPI_GENERATOR_SERVER_CODE_PATH}" ${BASE_CONSOLE}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Removing old generated code from ${OPENAPI_GENERATOR_SERVER_CODE_PATH}" ${BASE_CONSOLE}
	rm -f -r ${OPENAPI_GENERATOR_SERVER_CODE_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${DOCKER_OPENAPI_SERVER_GENERATOR_COMMAND} \
		-i ${OPENAPI_OUT_PATH} \
		-o ${OPENAPI_GENERATOR_SERVER_CODE_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-code-server-go-compose

openapi-code-server-go-format: git-config-command gocache
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Formatting openapi generated code for ${SERVICE_NAME} to ${OPENAPI_GENERATOR_SERVER_CODE_PATH}" ${BASE_CONSOLE}
	${DOCKER_GO_COMMAND} \
	go fmt ${OPENAPI_GENERATOR_SERVER_CODE_PATH}/pkg/api/*.go || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-code-server-go-format

openapi-generate-code-client-go: openapi-generate-code-client-go-raw openapi-code-client-go-format openapi-fix-client-go
.PHONY: openapi-generate-code-client-go

openapi-code-client-go-format: git-config-command
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Formatting openapi generated code for ${OPENAPI_CLIENT_NAME_GO} in ${OPENAPI_GENERATOR_CLIENT_CODE_PATH}" ${BASE_CONSOLE}
	${DOCKER_GO_COMMAND} \
	go fmt ${OPENAPI_GENERATOR_CLIENT_CODE_PATH}/*.go || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-code-client-go-format

openapi-generate-code-client-js:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating openapi js client code for ${SERVICE_NAME} to ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH}" ${BASE_CONSOLE}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Removing old generated js client code from ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH}" ${BASE_CONSOLE}

	rm -f -r ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${DOCKER_OPENAPI_CLIENT_JS_GENERATOR_COMMAND} \
		--skip-validate-spec \
		-i ${OPENAPI_GENERATOR_CLIENT_SWAGGER_PATH} \
		-t ${OPENAPI_GENERATOR_CLIENT_TS_TEMPLATES_PATH} \
		-o ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
	sed 's/, RequestArgs/, type RequestArgs/' < ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH}/api.ts > /tmp/api.ts
	cp /tmp/api.ts ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH}/api.ts
	sed 's/, RequestArgs/, type RequestArgs/'  < ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH}/common.ts > /tmp/common.ts
	cp /tmp/common.ts ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH}/common.ts
	@# Drop the unused git_push.sh scaffolding, matching openapi-fix-client-go on
	@# the Go side. Needed here too: openapi-check-client-code-uptodate-js diffs
	@# this output against the committed tree, so if only one side drops the file
	@# the check fails on "Only in ...: git_push.sh".
	rm -f ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH}/git_push.sh
.PHONY: openapi-generate-code-client-js

openapi-generate-code-client-go-raw: git-config-command
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating openapi go client code for ${SERVICE_NAME} to ${OPENAPI_GENERATOR_CLIENT_CODE_PATH}" ${BASE_CONSOLE}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Removing old generated go client code from ${OPENAPI_GENERATOR_CLIENT_CODE_PATH}" ${BASE_CONSOLE}
	rm -f -r ${OPENAPI_GENERATOR_CLIENT_CODE_PATH} || ${FAILURE_MESSAGE_HANDLER}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating openapi go client code for ${SERVICE_NAME} to ${OPENAPI_GENERATOR_CLIENT_CODE_PATH}" ${BASE_CONSOLE}
	@${DOCKER_OPENAPI_CLIENT_GO_GENERATOR_COMMAND} \
		-i ${OPENAPI_OUT_PATH} \
		-o ${OPENAPI_GENERATOR_CLIENT_CODE_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-generate-code-client-go-raw

openapi-fix-client-go:
	@${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Deleting ${OPENAPI_CLIENT_NAME_GO} tests folder and unused generator scaffolding" ${BASE_CONSOLE}
	@# git_push.sh is openapi-generator scaffolding we never use. Dropping it here
	@# - the same cleanup control-plane/notifications-service already does - keeps
	@# it out of the tree entirely, so it neither ships nor shows up in the
	@# regenerate-and-diff check further down this file. This target runs on both
	@# the real output path and the temp path used by
	@# openapi-check-client-code-uptodate-go, so both sides stay consistent.
	@${DOCKER_GO_COMMAND} /bin/sh -c 'cd ${OPENAPI_GENERATOR_CLIENT_CODE_PATH} && (rm -rf test || echo "no test folder") && rm -f git_push.sh' || ${FAILURE_MESSAGE_HANDLER}

	@${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Handling go.mod and go.sum for ${OPENAPI_CLIENT_NAME_GO}" ${BASE_CONSOLE}
	@${DOCKER_GO_COMMAND} /bin/sh -c '\
		echo "${CONSOLE_PREFIX} Moving go.mod and go.sum from ${OPENAPI_GENERATOR_CLIENT_CODE_PATH}"; \
		for f in go.mod go.sum; do \
			if [ -f "${OPENAPI_GENERATOR_CLIENT_CODE_PATH}/$$f" ] && [ ! -f "${OPENAPI_GENERATOR_CLIENT_CODE_PATH}/../$$f" ]; then \
				mv "${OPENAPI_GENERATOR_CLIENT_CODE_PATH}/$$f" "${OPENAPI_GENERATOR_CLIENT_CODE_PATH}/../$$f"; \
			elif [ -f "${OPENAPI_GENERATOR_CLIENT_CODE_PATH}/$$f" ]; then \
				rm -f "${OPENAPI_GENERATOR_CLIENT_CODE_PATH}/$$f"; \
			fi; \
		done; \
		echo "${CONSOLE_PREFIX} Running go mod tidy in ${SHARED_PACKAGES_PATH}/${OPENAPI_CLIENT_NAME_GO}"; \
		cd "${SHARED_PACKAGES_PATH}/${OPENAPI_CLIENT_NAME_GO}" && echo go mod tidy'
	@${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} ${SERVICE_NAME} Client build finished" ${BASE_CONSOLE}
.PHONY: openapi-fix-client-go

openapi-generate-code: openapi-generate-code-server-go openapi-generate-code-client-go openapi-generate-code-client-js
.PHONY: openapi-generate-code

openapi-check-yaml-uptodate-generic: 
	make openapi-bundle-generic OPENAPI_OUT_PATH=${OPENAPI_TMP_OUT_PATH}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating diff between ${OPENAPI_TMP_OUT_PATH} and ${OPENAPI_OUT_PATH} for ${SERVICE_NAME} service " ${BASE_CONSOLE}
	${DOCKER_LINUX_COMMAND} diff -r -w -B ${OPENAPI_TMP_OUT_PATH} ${OPENAPI_OUT_PATH} || ${FAILURE_MESSAGE_HANDLER}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating diff between ${OPENAPI_TMP_OUT_PATH} and ${OPENAPI_CENTRAL_SPECS_PATH}/${SERVICE_NAME}.yaml for ${SERVICE_NAME} service " ${BASE_CONSOLE}
	${DOCKER_LINUX_COMMAND} diff -r -w -B ${OPENAPI_TMP_OUT_PATH} ${OPENAPI_CENTRAL_SPECS_PATH}/${SERVICE_NAME}.yaml || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-check-yaml-uptodate-generic

openapi-check-service-field: openapi-bundle-generic
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Check if x-service field is present on all paths in ${OPENAPI_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '[.paths[] | to_entries | .[] | select(.key == "get" or .key == "post" or .key == "put" or .key == "delete" or .key == "patch") | select(.value."x-service" == null and .value."x-runai-team" == null)] | length == 0' ${OPENAPI_OUT_PATH} | grep -q "true" || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}

openapi-validate-public-service-routes-published:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-authorization fields from spec ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-no-authorization?) | del(.. | .x-authorize?))' -i ${OPENAPI_TMP_OUT_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-test-coverage fields from spec ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-test-coverage?))' -i ${OPENAPI_TMP_OUT_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-runai-team and x-service fields from spec ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-runai-team?) | del(.. | .x-service?))' -i ${OPENAPI_TMP_OUT_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-default fields from spec ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-default?))' -i ${OPENAPI_TMP_OUT_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-audit fields from spec ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-audit?))' -i ${OPENAPI_TMP_OUT_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup internal paths from spec ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | ./internal/*?))' -i ${OPENAPI_TMP_OUT_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-internal routes from spec ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_OPENAPI_COMMAND} bundle ${OPENAPI_TMP_OUT_PATH} -o ${OPENAPI_TMP_OUT_PATH} --config ${OPENAPI_PUBLIC_CLEANUP_CONFIG_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Checkin if public paths show up in public api ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_COMMAND} -e OPENAPI_TMP_OUT_PATH=${OPENAPI_TMP_OUT_PATH} -e OPENAPI_PUBLIC_SPEC_PATH=${OPENAPI_PUBLIC_SPEC_PATH} --entrypoint sh ${DOCKER_YQ_IMG} ${OPENAPI_CENTRAL_SCRIPTS_PATH}/validate_service_public_paths.sh || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-validate-public-service-routes-published

openapi-validate-auth-exists:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating components.securitySchemes.bearerAuth exists in ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '.components.securitySchemes | keys | .[] | test("(?i)bearerAuth")' ${OPENAPI_TMP_OUT_PATH} | grep -q true || ${FAILURE_MESSAGE_HANDLER}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating security.[bearerAuth] exists in ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} '.security.[] | keys | .[] | test("(?i)bearerAuth")' ${OPENAPI_TMP_OUT_PATH} | grep -q true || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-validate-auth-exists

openapi-validate-yaml: openapi-check-yaml-uptodate-generic openapi-validate-auth-exists
.PHONY: openapi-validate-yaml-uptodate

openapi-check-server-code-uptodate-go: git-config-command
	make openapi-generate-code-server-go OPENAPI_GENERATOR_SERVER_CODE_PATH=${OPENAPI_GENERATOR_SERVER_TMP_CODE_PATH}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating diff between ${OPENAPI_GENERATOR_SERVER_TMP_CODE_PATH}/${OPENAPI_GENERATOR_INTERNAL_CODE_PATH} and ${OPENAPI_GENERATOR_SERVER_CODE_PATH}/${OPENAPI_GENERATOR_INTERNAL_CODE_PATH} for ${SERVICE_NAME} service " ${BASE_CONSOLE}
	${DOCKER_LINUX_COMMAND} diff -r -w -B ${OPENAPI_GENERATOR_SERVER_TMP_CODE_PATH}/${OPENAPI_GENERATOR_INTERNAL_CODE_PATH} ${OPENAPI_GENERATOR_SERVER_CODE_PATH}/${OPENAPI_GENERATOR_INTERNAL_CODE_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-check-code-uptodate-generic

openapi-check-client-code-uptodate-go: git-config-command
	make openapi-generate-code-client-go OPENAPI_GENERATOR_CLIENT_CODE_PATH=${OPENAPI_GENERATOR_CLIENT_TMP_CODE_PATH}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating diff between ${OPENAPI_GENERATOR_CLIENT_TMP_CODE_PATH} and ${OPENAPI_GENERATOR_CLIENT_CODE_PATH} for ${SERVICE_NAME} service " ${BASE_CONSOLE}
	${DOCKER_LINUX_COMMAND} diff -r -w -B --exclude=go.sum --exclude=go.mod --exclude=api --exclude=docs --exclude=mocks --exclude=test ${OPENAPI_GENERATOR_CLIENT_CODE_PATH} ${OPENAPI_GENERATOR_CLIENT_TMP_CODE_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-check-code-uptodate-generic

openapi-check-client-code-uptodate-js: git-config-command
	make openapi-generate-code-client-js OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH=${OPENAPI_GENERATOR_CLIENT_TMP_CODE_PATH}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Validating diff between ${OPENAPI_GENERATOR_CLIENT_TMP_CODE_PATH} and ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH} for ${SERVICE_NAME} service " ${BASE_CONSOLE}
	${DOCKER_LINUX_COMMAND} diff -r -w -B ${OPENAPI_GENERATOR_CLIENT_TS_CODE_PATH} ${OPENAPI_GENERATOR_CLIENT_TMP_CODE_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-check-code-uptodate-generic

validate-openapi-public-spec:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} validate-openapi-public-spec ${OPENAPI_TMP_OUT_PATH}" ${BASE_CONSOLE}
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Run lint on spec ${OPENAPI_BUILD_PATH}/_build/runaiApi.swagger.yaml" ${BASE_CONSOLE}
	${DOCKER_OPENAPI_COMMAND} lint --config ${OPENAPI_BUILD_PATH}/config/lint.yaml ${OPENAPI_BUILD_PATH}/_build/runaiApi.swagger.yaml --format=stylish
	${SUCCESS_MESSAGE_HANDLER};
.PHONY: validate-openapi-public-spec

openapi-validate-code: openapi-check-server-code-uptodate-go openapi-check-client-code-uptodate-go openapi-check-client-code-uptodate-js
.PHONY: openapi-validate-code-uptodate

openapi-validate: openapi-validate-yaml openapi-validate-code openapi-validate-public-service-routes-published openapi-check-service-field
.PHONY: openapi-validate

## Public specs

openapi-generate-public-spec:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating public spec for runai control-plane to ${OPENAPI_PUBLIC_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_NODEJS_COMMAND} /bin/sh -c 'cd ${OPENAPI_BUILD_PATH}/scripts/cli && npm --cache /tmp/.npm install \
	&& ./node_modules/.bin/openapi-merge-cli --config ../../config/merge-config.json' || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Store raw swagger for runai control-plane to ${OPENAPI_PUBLIC_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_LINUX_COMMAND} cp ${OPENAPI_BUILD_DIR}/runaiApi.swagger.yaml ${OPENAPI_CENTRAL_SPECS_PATH}/runai-public-raw.yaml || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-generate-public-spec

openapi-cleanup-spec:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-authorization fields from spec ${OPENAPI_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-no-authorization?) | del(.. | .x-authorize?))' -i ${OPENAPI_SPEC_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-test-coverage fields from spec ${OPENAPI_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-test-coverage?))' -i ${OPENAPI_SPEC_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-runai-team and x-service fields from spec ${OPENAPI_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-runai-team?) | del(.. | .x-service?))' -i ${OPENAPI_SPEC_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-default fields from spec  ${OPENAPI_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-default?))' -i ${OPENAPI_SPEC_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-audit fields from spec ${OPENAPI_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-audit?))' -i ${OPENAPI_SPEC_PATH}
	
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup internal paths from spec ${OPENAPI_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | ./internal/*?))' -i ${OPENAPI_SPEC_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup routes from spec ${OPENAPI_SPEC_PATH} based on ${OPENAPI_CLEANUP_CONFIG_PATH}" ${BASE_CONSOLE}
	${DOCKER_OPENAPI_COMMAND} bundle ${OPENAPI_SPEC_PATH} -o ${OPENAPI_SPEC_PATH} --config ${OPENAPI_CLEANUP_CONFIG_PATH}

	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Cleanup x-scope fields from spec ${OPENAPI_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_YQ_COMMAND} e '. as $$yaml | ($$yaml | del(.. | .x-scope?))' -i ${OPENAPI_SPEC_PATH}
	${SUCCESS_MESSAGE_HANDLER};
.PHONY: openapi-cleanup-spec

openapi-publish-spec:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Copying ${SOURCE_SPEC_PATH} spec for runai control-plane to ${DEST_SPEC_PATH}" ${BASE_CONSOLE}
	${DOCKER_LINUX_COMMAND} cp ${SOURCE_SPEC_PATH} ${DEST_SPEC_PATH} || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-publish-spec

openapi-build-public-spec: openapi-generate-public-spec
	$(MAKE) openapi-cleanup-spec OPENAPI_SPEC_PATH=${OPENAPI_PUBLIC_SPEC_BUILD_PATH} OPENAPI_CLEANUP_CONFIG_PATH=${OPENAPI_PUBLIC_CLEANUP_CONFIG_PATH}
	$(MAKE) openapi-publish-spec SOURCE_SPEC_PATH=${OPENAPI_PUBLIC_SPEC_BUILD_PATH} DEST_SPEC_PATH=${OPENAPI_CENTRAL_SPECS_PATH}/runai-public.yaml
	${SUCCESS_MESSAGE_HANDLER}
	$(MAKE) openapi-build-system-scoped-spec
.PHONY: openapi-build-public-spec

openapi-build-system-scoped-spec: openapi-generate-system-scoped-spec
	$(MAKE) openapi-cleanup-spec OPENAPI_SPEC_PATH=${OPENAPI_SYSTEM_SCOPED_SPEC_BUILD_PATH} OPENAPI_CLEANUP_CONFIG_PATH=${OPENAPI_SYSTEM_SCOPED_CLEANUP_CONFIG_PATH}
	$(MAKE) openapi-publish-spec SOURCE_SPEC_PATH=${OPENAPI_SYSTEM_SCOPED_SPEC_BUILD_PATH} DEST_SPEC_PATH=${OPENAPI_CENTRAL_SPECS_PATH}/runai-system-scoped.yaml
.PHONY: openapi-build-system-scoped-spec

openapi-generate-system-scoped-spec:
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Generating system scoped spec for runai control-plane" ${BASE_CONSOLE}
	${DOCKER_NODEJS_COMMAND} /bin/sh -c 'cd ${OPENAPI_BUILD_PATH}/scripts/cli && npm --cache /tmp/.npm install \
	&& ./node_modules/.bin/openapi-merge-cli --config ../../config/merge-config-system-system-scoped.json' || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: openapi-generate-system-scoped-spec
 
