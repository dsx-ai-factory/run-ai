#!/bin/bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


# ============================================================
# Run:AI Workloads Kit - Entry Point
# ============================================================

# Determine script directory so kit-utils.sh can be sourced
# regardless of where the script is called from.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source shared kit utilities
source "${SCRIPT_DIR}/kit-common/kit-utils.sh"

# ---------------------------------------------------------------
# Kit-specific configuration
# ---------------------------------------------------------------
KIT_NAME="runai-workloads-kit"
KIT_DISPLAY_NAME="Run:AI Workloads Kit"
PLAYWRIGHT_TAG="@k8s-flav-workloads-kit"
PLAYWRIGHT_CONFIG="playwright-config/runai-workloads-kit.config.ts"
ARCHIVE_PREFIX="runai-workloads-results"

# ---------------------------------------------------------------
# 1. Start log capture
# ---------------------------------------------------------------
setup_log_capture "$KIT_NAME"

echo "Starting ${KIT_DISPLAY_NAME}..."
echo "All output is being logged to: $LOG_FILE"

# ---------------------------------------------------------------
# 2. Validate required environment variables
# ---------------------------------------------------------------
echo "Validating required parameters..."
require_env_var "CONTROL_PLANE_URL"    "https://your-control-plane.com"
require_env_var "CLUSTER_NAME_PREFIX"  "your-cluster-name"
require_env_var "CLUSTER_URL"          "https://your-k8s-api.com"

echo "All required parameters validated"
echo "  CONTROL_PLANE_URL:   $CONTROL_PLANE_URL"
echo "  CLUSTER_NAME_PREFIX: $CLUSTER_NAME_PREFIX"
echo "  CLUSTER_URL:         $CLUSTER_URL"
echo ""

# ---------------------------------------------------------------
# 3. Load and validate kubeconfig
# ---------------------------------------------------------------
load_kubeconfig
run_preflight_checks

# ---------------------------------------------------------------
# 4. Write .env file and set environment defaults
# ---------------------------------------------------------------
write_env_file
set_default_env_vars

# ---------------------------------------------------------------
# 5. Run workloads kit tests
# ---------------------------------------------------------------
cd /app/e2e
echo "Running ${KIT_DISPLAY_NAME} tests..."
run_playwright_tests "$PLAYWRIGHT_TAG" "$PLAYWRIGHT_CONFIG"

# ---------------------------------------------------------------
# 6. Generate and publish Allure report
# ---------------------------------------------------------------
check_and_generate_allure_report
copy_allure_report_to_results

# ---------------------------------------------------------------
# 7. Package results for investigation
# ---------------------------------------------------------------
package_results "$KIT_NAME" "${KIT_NAME}-run.log" "$ARCHIVE_PREFIX"

echo ""
echo "${KIT_DISPLAY_NAME} completed!"
echo "Check the generated Allure report for detailed results."

# Final log sync to capture completion message (also runs via EXIT trap)
sync_logs
