#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# bundle.sh - Create distribution package for unified tests external client
# Usage: ./bundle.sh [IMAGE_NAME] [VERSION]

set -e

# Configuration
IMAGE_NAME=${1:-runai/unified-tests}
VERSION=${2:-$(git rev-parse --short HEAD 2>/dev/null || echo "latest")}
IMAGE_TAG="${IMAGE_NAME}:latest"
BUNDLE_NAME="unified-tests-client-${VERSION}"
DIST_DIR="dist"
TEMP_BUNDLE_DIR="${DIST_DIR}/${BUNDLE_NAME}"
REPO_ROOT="../.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

step() {
    echo -e "${BLUE}==> $1${NC}"
}

# Validate prerequisites
if ! command -v docker &> /dev/null; then
    error "Docker is not installed. Please install Docker and try again."
fi

if ! docker info &> /dev/null; then
    error "Docker is not running. Please start Docker and try again."
fi

# Display configuration
echo "=========================================="
echo "Run:AI Unified Tests - Bundle Creation"
echo "=========================================="
echo "Image Name: ${IMAGE_NAME}"
echo "Version: ${VERSION}"
echo "Bundle Name: ${BUNDLE_NAME}.tar.gz"
echo "=========================================="
echo ""

# Step 1: Prepare Go dependencies
step "Step 1/7: Preparing Go dependencies..."
if [ -d "vendor" ]; then
    info "Vendor directory already exists"
else
    info "Running go mod vendor..."
    go mod vendor || error "Failed to run go mod vendor"
fi

# Step 2: Build Docker image
step "Step 2/7: Building Docker image..."
info "Building ${IMAGE_TAG} from repo root..."
cd "${REPO_ROOT}"
docker build -t "${IMAGE_TAG}" -f scale-tests/unified_tests/Dockerfile . || error "Failed to build Docker image"
cd - > /dev/null
info "Docker image built successfully"

# Step 3: Save Docker image to tarball
step "Step 3/7: Saving Docker image to tarball..."
mkdir -p "${DIST_DIR}"
docker save "${IMAGE_TAG}" -o "${DIST_DIR}/docker-image.tar" || error "Failed to save Docker image"
IMAGE_SIZE=$(du -h "${DIST_DIR}/docker-image.tar" | cut -f1)
info "Docker image saved: ${IMAGE_SIZE}"

# Step 4: Create bundle directory structure
step "Step 4/7: Creating bundle directory structure..."
rm -rf "${TEMP_BUNDLE_DIR}"
mkdir -p "${TEMP_BUNDLE_DIR}"
info "Created ${TEMP_BUNDLE_DIR}"

# Step 5: Copy client-dist contents (excluding README.md)
step "Step 5/7: Copying distribution files..."
# Copy everything except README.md into the bundle
find client-dist -mindepth 1 -maxdepth 1 ! -name 'README.md' -exec cp -r {} "${TEMP_BUNDLE_DIR}/" \;
info "Copied client-dist contents (excluding README.md)"

# Copy README.md to dist directory (outside the bundle)
cp client-dist/README.md "${DIST_DIR}/"
info "Copied README.md to ${DIST_DIR}/"

# Step 6: Move Docker image to bundle root
step "Step 6/7: Adding Docker image to bundle..."
mv "${DIST_DIR}/docker-image.tar" "${TEMP_BUNDLE_DIR}/"
info "Docker image added to bundle"

# Step 7: Create compressed tarball
step "Step 7/7: Creating compressed archive..."
cd "${DIST_DIR}"
tar -czf "${BUNDLE_NAME}.tar.gz" "${BUNDLE_NAME}" || error "Failed to create tarball"
cd - > /dev/null

# Get final size
BUNDLE_SIZE=$(du -h "${DIST_DIR}/${BUNDLE_NAME}.tar.gz" | cut -f1)

# Cleanup temporary directory
rm -rf "${TEMP_BUNDLE_DIR}"

# Success
echo ""
echo "=========================================="
info "✓ Bundle created successfully!"
echo "=========================================="
echo "Bundle: ${DIST_DIR}/${BUNDLE_NAME}.tar.gz"
echo "README: ${DIST_DIR}/README.md"
echo "Size: ${BUNDLE_SIZE}"
echo ""
echo "To distribute:"
echo "  1. Share both ${BUNDLE_NAME}.tar.gz and README.md with external clients"
echo ""
echo "Clients should:"
echo "  1. Read README.md for prerequisites and setup"
echo "  2. Extract: tar -xzf ${BUNDLE_NAME}.tar.gz"
echo "  3. Load image: cd ${BUNDLE_NAME} && docker load < docker-image.tar"
echo "  4. Run tests: ./run-test.sh examples/small-scale-test.env"
echo "=========================================="
