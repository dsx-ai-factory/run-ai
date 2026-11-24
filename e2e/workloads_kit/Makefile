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

help:
	@bash $(dir $(abspath $(lastword $(MAKEFILE_LIST))))../kit-common/print-help.sh "Workloads Kit"
	@echo "The image is built from this checkout of the e2e suite (build context: e2e/)."

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

run:
	@echo "» Run:AI Workloads Kit"
	@echo ""
	@if [ -z "$(KUBECONFIG_PATH)" ]; then \
		echo "✘ KUBECONFIG_PATH is required"; \
		echo ""; \
		echo "  Usage: make run KUBECONFIG_PATH=/path/to/your/kubeconfig"; \
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
	@docker run --rm -it \
		-v $(PWD)/results:/app/e2e/results \
		-v $(KUBECONFIG_PATH):/kubeconfig/config:ro \
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
