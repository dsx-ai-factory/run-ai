#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


set -euo pipefail

# Script to generate values.schema.json for RunAI Helm chart
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
SDK_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Default output location (in the runai-chart directory)
DEFAULT_OUTPUT="${PROJECT_ROOT}/runai-cluster/cluster/runai-chart/values.schema.json"
OUTPUT_PATH="${1:-$DEFAULT_OUTPUT}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[schema-gen]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[schema-gen]${NC} $1"
}

error() {
    echo -e "${RED}[schema-gen]${NC} $1" >&2
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    error "Go is required but not installed"
    exit 1
fi

# Change to SDK directory to ensure proper module context
cd "${SDK_DIR}"

log "Generating JSON schema for RunAI Helm values..."
log "Output path: ${OUTPUT_PATH}"

# Build and run the generator
if ! go run hack/generate-values-schema.go "${OUTPUT_PATH}"; then
    error "Failed to generate schema"
    exit 1
fi

# Validate the generated JSON
if command -v jq &> /dev/null; then
    if cat "${OUTPUT_PATH}" | jq empty 2>/dev/null; then
        log "✅ Generated JSON is valid"
    else
        error "❌ Generated JSON is invalid"
        exit 1
    fi
else
    warn "⚠️  jq not found - skipping JSON validation"
fi

log "🎉 Successfully generated values.schema.json"
log "   File: ${OUTPUT_PATH}"

# Show file size for reference
if command -v wc &> /dev/null; then
    LINES=$(wc -l < "${OUTPUT_PATH}")
    SIZE=$(du -h "${OUTPUT_PATH}" | cut -f1)
    log "   Size: ${SIZE}, Lines: ${LINES}"
fi

echo
echo "Usage examples:"
echo "  # Validate values file against schema"
echo "  jsonschema -i values.yaml ${OUTPUT_PATH}"
echo
echo "  # Use in Helm (requires Helm 3.4+)"
echo "  helm install runai . --values values.yaml --validate"
