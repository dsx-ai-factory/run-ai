#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# run-test.sh - Helper script for running Run:AI unified scale tests
# Usage: ./run-test.sh [config-file] [results-dir]

set -e

# Configuration
CONFIG_FILE=${1:-test-config.env}
RESULTS_DIR=${2:-./results}
IMAGE_NAME=${RUNAI_TESTS_IMAGE:-runai/unified-tests:latest}
KUBECONFIG_PATH=${KUBECONFIG:-~/.kube/config}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

info() {
    echo -e "${GREEN}$1${NC}"
}

warn() {
    echo -e "${YELLOW}Warning: $1${NC}"
}

# Validate Docker is installed
if ! command -v docker &> /dev/null; then
    error "Docker is not installed. Please install Docker and try again."
fi

# Validate Docker is running
if ! docker info &> /dev/null; then
    error "Docker is not running. Please start Docker and try again."
fi

# Validate config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    error "Config file '$CONFIG_FILE' not found. Please provide a valid config file."
fi

# Validate kubeconfig exists
if [ ! -f "$KUBECONFIG_PATH" ]; then
    error "Kubeconfig not found at '$KUBECONFIG_PATH'. Please ensure kubectl is configured."
fi

# Create results directory
mkdir -p "$RESULTS_DIR"

# Display test information
echo "=========================================="
echo "Run:AI Unified Scale Tests"
echo "=========================================="
echo "Config file: $CONFIG_FILE"
echo "Results directory: $RESULTS_DIR"
echo "Docker image: $IMAGE_NAME"
echo "Kubeconfig: $KUBECONFIG_PATH"
echo "=========================================="
echo ""

# Check if kubectl can connect to cluster
info "Validating cluster connectivity..."
if ! kubectl cluster-info &> /dev/null; then
    warn "Cannot connect to Kubernetes cluster. Tests may fail."
fi

info "Starting test execution..."
echo ""

# Resolve absolute paths
CONFIG_FILE_ABS="$(cd "$(dirname "$CONFIG_FILE")" && pwd)/$(basename "$CONFIG_FILE")"
RESULTS_DIR_ABS="$(cd "$RESULTS_DIR" 2>/dev/null && pwd || (mkdir -p "$RESULTS_DIR" && cd "$RESULTS_DIR" && pwd))"

# Run tests
docker run --rm \
  -v "$KUBECONFIG_PATH:/app/kubeconfig/config:ro" \
  -v "$RESULTS_DIR_ABS:/app/results" \
  -v "$CONFIG_FILE_ABS:/app/config/test-config.env:ro" \
  "$IMAGE_NAME"

# Check exit code
EXIT_CODE=$?

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    info "✓ Tests completed successfully"
    info "Results available in: $RESULTS_DIR/"
else
    error "Tests failed with exit code $EXIT_CODE"
fi

exit $EXIT_CODE
