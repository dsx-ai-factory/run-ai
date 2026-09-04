include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

# This Makefile sits two levels deep (scale-tests/unified_tests/), which
# matches control-plane.mk's depth-2 default, so the workspace root is two
# levels up.
WORKSPACE_HOST_PATH=${PWD}/../..

BIN_DIR=$(shell pwd)/bin

test:
	# no-op
.PHONY: test

lint:
	# no-op
.PHONY: lint

validate:
	# no-op
.PHONY: validate

purify:
	# no-op
.PHONY: purify

build: git-config-command bin ## Build the service.
	go build -o $(BIN_DIR) ./cmd/*
.PHONY: build

bin:
	mkdir -p $(BIN_DIR)

# ==============================================================================
# Docker Targets for External Client Distribution
# ==============================================================================

# Docker image configuration
IMAGE_NAME ?= runai/unified-tests
VERSION ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "latest")
LATEST_TAG = latest

# Build Docker image

# Push Docker image to registry

# Test Docker image locally
.PHONY: docker-test
docker-test:
	@echo "Testing Docker image with small-scale config..."
	@if [ ! -f client-dist/examples/small-scale-test.env ]; then \
		echo "Error: Example config not found"; \
		exit 1; \
	fi
	@mkdir -p test-results
	docker run --rm \
		-v ~/.kube/config:/root/.kube/config:ro \
		-v $(shell pwd)/test-results:/app/results \
		-v $(shell pwd)/client-dist/examples/small-scale-test.env:/app/config/test-config.env:ro \
		$(IMAGE_NAME):$(LATEST_TAG)

# Create distribution package for external clients
.PHONY: package
package:
	@echo "Creating distribution package..."
	@./bundle.sh $(IMAGE_NAME) $(VERSION)

# Clean Docker artifacts
.PHONY: docker-clean
docker-clean:
	@echo "Cleaning Docker artifacts..."
	@docker rmi -f $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):$(LATEST_TAG) 2>/dev/null || true
	@rm -rf dist/ test-results/
	@echo "Docker artifacts cleaned"

# Display help for Docker targets
.PHONY: docker-help
docker-help:
	@echo "Docker Targets:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-push   - Push Docker image to registry"
	@echo "  docker-test   - Test Docker image locally with small-scale config"
	@echo "  package       - Create distribution package for external clients"
	@echo "  docker-clean  - Clean Docker artifacts"
	@echo ""
	@echo "Environment variables:"
	@echo "  IMAGE_NAME    - Docker image name (default: runai/unified-tests)"
	@echo "  VERSION       - Image version (default: git describe)"
