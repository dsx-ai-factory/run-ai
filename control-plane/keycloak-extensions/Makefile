# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# Change for every new service
SERVICE_NAME=keycloak-extensions
GRADLE_IMAGE=${DOCKERHUB_PROXY}library/gradle:8.12.0-jdk21
##@ Basic Commands
REPORTS_DIR=runai-custom-providers/build/reports/tests/test/classes/
PARSE_SCRIPT=runai-custom-providers/scripts/parse_html_report.py
GRADLE_USER_HOME=/local/control-plane/keycloak-extensions/.gradle

test: test-gradle test-go ## Run unit-tests (Gradle SPI providers + Go init binary).
.PHONY: test

test-gradle: ## Run the Gradle (Java SPI providers) unit-tests.
	@echo "Running tests..."
	${DOCKER_COMMAND} -e GRADLE_OPTS="-Dorg.gradle.native=false" \
	-e GRADLE_USER_HOME="${GRADLE_USER_HOME}" \
	${GRADLE_IMAGE} \
	/bin/sh -c \
	'gradle test -p runai-custom-providers; \
	 TEST_EXIT_CODE=$$?; \
	 echo "Listing HTML reports directory:" && \
	 ls -R runai-custom-providers/build/reports/tests/test/classes/ && \
	 echo "Parsing HTML reports:" && \
	 python3 $(PARSE_SCRIPT) $(REPORTS_DIR); \
     rm -rf runai-custom-providers/build/reports/tests/test/classes/; \
	 exit $$TEST_EXIT_CODE'
.PHONY: test-gradle

lint: lint-go ## Run linters (Go init binary; the Gradle side has none).
.PHONY: lint

build: build-gradle build-go ## Build the keycloak-extensions image artifacts (Gradle JAR + Go init binary).
.PHONY: build

build-gradle: ## Build the Gradle (Java SPI providers + OTel agent) artifacts.
	${DOCKER_COMMAND} -e GRADLE_OPTS="-Dorg.gradle.native=false" \
	-e GRADLE_USER_HOME="${GRADLE_USER_HOME}" \
	${GRADLE_IMAGE} \
	/bin/sh -c \
	'rm -rf /tmp/.gradle/caches && \
	mkdir -p bin && \
	curl -sfL https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v2.29.0/opentelemetry-javaagent.jar -o bin/opentelemetry-javaagent.jar && \
	echo "546531ca690a8603d2923b6db26bbda35c6409327b1e610430ae33c2f8f68050  bin/opentelemetry-javaagent.jar" | sha256sum -c - && \
	gradle build -p runai-custom-providers && \
	cp ./runai-custom-providers/build/libs/runai-custom-providers.jar bin/runai-custom-providers.jar'
.PHONY: build-gradle

validate: validate-mockery ## Run validate.
	# no-op
.PHONY: validate
