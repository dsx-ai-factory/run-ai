# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Run:AI Workloads Kit - Makefile
# This tool validates Run:AI workload functionality using the E2E test suite

.PHONY: help build test lint validate build-amd64 build-arm64 build-both-tar setup-buildx run run-rebuild force-rebuild package

# The image packages the e2e suite from this checkout, so the build context is
# the parent directory (e2e/): every build below runs from there with
# -f workloads_kit/Dockerfile, and e2e/.dockerignore filters what is sent.
# The Dockerfile relies on BuildKit (heredocs, BUILDPLATFORM stages).
export DOCKER_BUILDKIT = 1

# start.sh requires CONTROL_PLANE_URL, CLUSTER_NAME_PREFIX and CLUSTER_URL; the
# admin credentials have defaults. `make run` forwards them with `-e NAME`, so
# export them to the recipe environment (this keeps the password off the docker
# command line). They may come from the shell or from `make run NAME=value`.
export CONTROL_PLANE_URL CLUSTER_NAME_PREFIX CLUSTER_URL \
	CONTROL_PLANE_ADMIN_USERNAME CONTROL_PLANE_ADMIN_PASSWORD

help:
	@bash $(dir $(abspath $(lastword $(MAKEFILE_LIST))))../kit-common/print-help.sh "Workloads Kit"
	@echo "The image is built from this checkout of the e2e suite (build context: e2e/)."
	@echo "make run also requires CONTROL_PLANE_URL, CLUSTER_NAME_PREFIX and CLUSTER_URL;"
	@echo "CONTROL_PLANE_ADMIN_USERNAME / CONTROL_PLANE_ADMIN_PASSWORD are optional."
	@echo "The make run flow securely stages mode-0600 kubeconfigs for uid 65532."

build:
	@echo "» Building Run:AI Workloads Kit image..."
	@cd .. && docker build -t runai/runai-workloads-kit:latest -f workloads_kit/Dockerfile . \
		&& echo "✔ Image built successfully" \
		|| (echo "✘ Build failed"; exit 1)

# test, lint and validate are the targets artifact CI calls for a registered
# artifact (see artifact-metadata.json). Functional kit tests require a
# Kubernetes cluster and run through the parent e2e suite, so there is no
# standalone test command here.
test:
	@echo "No standalone tests; workloads kit tests run through the parent e2e suite."

lint:
	@echo "» Checking kit shell scripts for syntax errors..."
	@for f in start.sh ../kit-common/kit-utils.sh ../kit-common/print-help.sh; do \
		bash -n "$$f" && echo "  ✔ $$f" || exit 1; \
	done

validate:
	@echo "No additional validation; image construction is verified by make build."

setup-buildx:
	@echo "» Setting up Docker buildx..."
	@if ! docker buildx ls | grep -q workloads-kit-builder; then \
		echo "  Creating builder 'workloads-kit-builder'..."; \
		docker buildx create --name workloads-kit-builder --use --bootstrap; \
	else \
		echo "  Builder 'workloads-kit-builder' already exists"; \
		docker buildx use workloads-kit-builder; \
	fi
	@echo "✔ Buildx ready"

build-amd64: setup-buildx
	@echo "» Building AMD64 (x86_64) image..."
	@cd .. && docker buildx build \
		--platform linux/amd64 \
		-t runai/runai-workloads-kit-amd64:latest \
		-f workloads_kit/Dockerfile . \
		--load
	@echo "  Saving to tar.gz..."
	@docker save runai/runai-workloads-kit-amd64:latest | gzip > runai-workloads-kit-amd64.tar.gz
	@echo "✔ AMD64 image saved"
	@echo "  File: runai-workloads-kit-amd64.tar.gz ($$(du -h runai-workloads-kit-amd64.tar.gz | cut -f1))"
	@echo ""

build-arm64: setup-buildx
	@echo "» Building ARM64 image..."
	@cd .. && docker buildx build \
		--platform linux/arm64 \
		-t runai/runai-workloads-kit-arm64:latest \
		-f workloads_kit/Dockerfile . \
		--load
	@echo "  Saving to tar.gz..."
	@docker save runai/runai-workloads-kit-arm64:latest | gzip > runai-workloads-kit-arm64.tar.gz
	@echo "✔ ARM64 image saved"
	@echo "  File: runai-workloads-kit-arm64.tar.gz ($$(du -h runai-workloads-kit-arm64.tar.gz | cut -f1))"
	@echo ""

build-both-tar: setup-buildx
	@echo "» Building both architectures..."
	@echo ""
	@echo "  [1/2] AMD64 (x86_64)"
	@cd .. && docker buildx build \
		--platform linux/amd64 \
		-t runai/runai-workloads-kit-amd64:latest \
		-f workloads_kit/Dockerfile . \
		--load
	@docker save runai/runai-workloads-kit-amd64:latest | gzip > runai-workloads-kit-amd64.tar.gz
	@echo "  ✔ AMD64 saved -> runai-workloads-kit-amd64.tar.gz ($$(du -h runai-workloads-kit-amd64.tar.gz | cut -f1))"
	@echo ""
	@echo "  [2/2] ARM64"
	@cd .. && docker buildx build \
		--platform linux/arm64 \
		-t runai/runai-workloads-kit-arm64:latest \
		-f workloads_kit/Dockerfile . \
		--load
	@docker save runai/runai-workloads-kit-arm64:latest | gzip > runai-workloads-kit-arm64.tar.gz
	@echo "  ✔ ARM64 saved -> runai-workloads-kit-arm64.tar.gz ($$(du -h runai-workloads-kit-arm64.tar.gz | cut -f1))"
	@echo ""
	@echo "✔ Both architectures built successfully"
	@echo ""
	@echo "  Load on client:"
	@echo "  docker load -i <(gunzip -c runai-workloads-kit-<arch>.tar.gz)"
	@echo "  docker tag runai/runai-workloads-kit-<arch>:latest runai/runai-workloads-kit:latest"
	@echo ""

# The runtime is fixed at uid 65532, so a host-owned mode-0600 kubeconfig
# cannot be bind-mounted directly. Stage it through a short-lived Docker volume
# with the runtime uid/mode, then mount that volume at the configured path.
run:
	@echo "» Run:AI Workloads Kit"
	@echo ""
	@missing=""; \
	[ -n "$(KUBECONFIG_PATH)" ] || missing="$$missing KUBECONFIG_PATH"; \
	[ -n "$(CONTROL_PLANE_URL)" ] || missing="$$missing CONTROL_PLANE_URL"; \
	[ -n "$(CLUSTER_NAME_PREFIX)" ] || missing="$$missing CLUSTER_NAME_PREFIX"; \
	[ -n "$(CLUSTER_URL)" ] || missing="$$missing CLUSTER_URL"; \
	if [ -n "$$missing" ]; then \
		echo "✘ Missing required variables:$$missing"; \
		echo ""; \
		echo "  Usage: make run KUBECONFIG_PATH=/path/to/your/kubeconfig \\"; \
		echo "           CONTROL_PLANE_URL=https://runai.example.com \\"; \
		echo "           CLUSTER_NAME_PREFIX=my-cluster \\"; \
		echo "           CLUSTER_URL=https://k8s.example.com \\"; \
		echo "           [CONTROL_PLANE_ADMIN_USERNAME=admin@example.com] \\"; \
		echo "           [CONTROL_PLANE_ADMIN_PASSWORD=yourpassword]"; \
		echo ""; \
		exit 1; \
	fi
	@if ! docker image inspect runai/runai-workloads-kit:latest >/dev/null 2>&1; then \
		echo "✘ Image not found. Run 'make build' first."; \
		exit 1; \
	fi
	@if [ ! -f "$(KUBECONFIG_PATH)" ]; then \
		echo "✘ Kubeconfig not found at: $(KUBECONFIG_PATH)"; \
		exit 1; \
	fi
	@mkdir -p results
	@echo "  Kubeconfig : $(KUBECONFIG_PATH)"
	@echo "  Results    : ./results/"
	@echo ""
	@echo "  Starting tests..."
	@echo ""
	@set -eu; \
	KUBECONFIG_VOLUME="runai-workloads-kit-kubeconfig-$$$$"; \
	cleanup() { docker volume rm "$$KUBECONFIG_VOLUME" >/dev/null 2>&1 || true; }; \
	trap cleanup 0 1 2 15; \
	docker volume create "$$KUBECONFIG_VOLUME" >/dev/null; \
	docker run --rm \
		--user 0:0 \
		--entrypoint /bin/bash \
		-v "$(abspath $(KUBECONFIG_PATH)):/source/config:ro" \
		--mount "type=volume,source=$$KUBECONFIG_VOLUME,target=/home/runai/.kube" \
		runai/runai-workloads-kit:latest \
		-ec 'cp /source/config /home/runai/.kube/config && chown 65532:65532 /home/runai/.kube/config && chmod 0600 /home/runai/.kube/config'; \
	docker run --rm -it \
		-e CONTROL_PLANE_URL \
		-e CLUSTER_NAME_PREFIX \
		-e CLUSTER_URL \
		$(if $(CONTROL_PLANE_ADMIN_USERNAME),-e CONTROL_PLANE_ADMIN_USERNAME) \
		$(if $(CONTROL_PLANE_ADMIN_PASSWORD),-e CONTROL_PLANE_ADMIN_PASSWORD) \
		-v "$(PWD)/results:/app/e2e/results" \
		--mount "type=volume,source=$$KUBECONFIG_VOLUME,target=/home/runai/.kube" \
		runai/runai-workloads-kit:latest
	@echo ""
	@echo "✔ Run complete"
	@echo ""
	@if [ -d "results/allure-report" ]; then \
		echo "  Report  : results/allure-report/index.html"; \
	fi
	@if ls results/*.zip 2>/dev/null 1>&2; then \
		echo "  Package : $$(ls results/*.zip)"; \
	elif ls results/*.tar.gz 2>/dev/null 1>&2; then \
		echo "  Package : $$(ls results/*.tar.gz)"; \
	fi
	@echo ""

run-rebuild: build run

force-rebuild:
	@echo "» Force rebuilding..."
	@docker rmi runai/runai-workloads-kit:latest 2>/dev/null && echo "  Removed old image" || true
	@$(MAKE) build

package:
	@echo "» Results Package"
	@echo ""
	@if [ ! -d "results" ]; then \
		echo "  No results directory found. Run 'make run' first."; \
		exit 0; \
	fi
	@if [ -d "results/allure-report" ]; then \
		echo "  Report  : results/allure-report/index.html"; \
	fi
	@if ls results/*.zip 2>/dev/null 1>&2; then \
		echo "  Package : $$(ls -lh results/*.zip | awk '{print $$9, "(", $$5, ")"}')"; \
		echo ""; \
		echo "  Ready to share with the Run:AI team."; \
	elif ls results/*.tar.gz 2>/dev/null 1>&2; then \
		echo "  Package : $$(ls -lh results/*.tar.gz | awk '{print $$9, "(", $$5, ")"}')"; \
		echo ""; \
		echo "  Ready to share with the Run:AI team."; \
	else \
		echo "  No package found. Run 'make run' first."; \
	fi
	@echo ""
