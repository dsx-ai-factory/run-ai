#!/bin/bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# ============================================================
# Run:AI Kit Shared Utilities
# Source this file from any kit's start.sh:
#   source "${SCRIPT_DIR}/kit-common/kit-utils.sh"
# ============================================================

# ---------------------------------------------------------------
# setup_log_capture <kit-name>
# Sets LOG_FILE, RESULTS_LOG_FILE, redirects stdout/stderr to tee,
# and registers sync_logs on EXIT.
# ---------------------------------------------------------------
setup_log_capture() {
    local kit_name="$1"
    LOG_FILE="/tmp/${kit_name}-run.log"
    RESULTS_LOG_FILE="/app/e2e/results/${kit_name}-run.log"

    # Clean up any existing log file
    > "$LOG_FILE"

    # Redirect all output through tee so it appears on console AND in log
    if command -v stdbuf >/dev/null 2>&1; then
        exec > >(stdbuf -oL tee -a "$LOG_FILE") 2>&1
    else
        exec > >(tee -a "$LOG_FILE") 2>&1
    fi

    # Ensure logs are synced to results on any exit
    trap 'sync_logs' EXIT
}

# ---------------------------------------------------------------
# sync_logs
# Flushes buffers and copies the log file into the results directory.
# Called automatically on EXIT; can also be called manually.
# ---------------------------------------------------------------
sync_logs() {
    sync 2>/dev/null || true
    sleep 0.5
    if [ -d "/app/e2e/results" ] && [ -f "$LOG_FILE" ]; then
        cp "$LOG_FILE" "$RESULTS_LOG_FILE" 2>/dev/null || true
    fi
}

# ---------------------------------------------------------------
# require_env_var <VAR_NAME> <example-value>
# Exits with an error message if the named variable is unset or empty.
# ---------------------------------------------------------------
require_env_var() {
    local var_name="$1"
    local hint="$2"
    if [ -z "${!var_name}" ]; then
        echo "ERROR: ${var_name} environment variable is required"
        echo ""
        echo "Please run with:"
        echo "  docker run -e ${var_name}=${hint} ..."
        exit 1
    fi
}

# ---------------------------------------------------------------
# load_kubeconfig
# Resolves the kubeconfig destination from $KUBECONFIG (falling back to
# $HOME/.kube/config), loads the config from the mounted file at
# /kubeconfig/config or from an already-staged $KUBECONFIG, and exports
# KUBECONFIG_CONTENT and KUBECONFIG_B64 for use in write_env_file.
# ---------------------------------------------------------------
load_kubeconfig() {
    export KUBECONFIG="${KUBECONFIG:-${HOME:-/root}/.kube/config}"

    # kubectl accepts a colon-separated list; these kits need a single writable path.
    case "$KUBECONFIG" in
        *:*)
            KUBECONFIG="${KUBECONFIG%%:*}"
            echo "Note: KUBECONFIG contained multiple paths; using first: $KUBECONFIG"
            ;;
    esac

    local src=""
    if [ -d /kubeconfig ] && [ ! -x /kubeconfig ]; then
        echo "ERROR: /kubeconfig directory is not traversable by uid $(id -u)."
        echo "       Check that the directory mode allows execute for this user."
        ls -ldn /kubeconfig
        exit 1
    elif [ -f /kubeconfig/config ]; then
        src=/kubeconfig/config
    elif [ -s "$KUBECONFIG" ]; then
        src="$KUBECONFIG"
        echo "Using pre-staged kubeconfig at $KUBECONFIG"
    else
        echo "ERROR: Kubeconfig not found at /kubeconfig/config or $KUBECONFIG"
        echo ""
        echo "Please mount your kubeconfig file:"
        echo "  docker run -v /path/to/your/kubeconfig:/kubeconfig/config:ro ..."
        echo "or stage it at $KUBECONFIG before start."
        exit 1
    fi

    if [ ! -r "$src" ]; then
        echo "ERROR: $src exists but is not readable by uid $(id -u)."
        echo ""
        echo "This kit runs as an unprivileged user. A kubeconfig bind-mounted"
        echo "from the host keeps its host ownership and mode, so a mode-0600"
        echo "file owned by a different user cannot be read here."
        echo ""
        echo "Mounted file:"
        ls -ln "$src"
        exit 1
    fi

    echo "Setting up kubeconfig..."
    KUBECONFIG_CONTENT=$(cat "$src")

    # Use -ef (inode comparison) rather than a string compare so that symlinks
    # and double-mounts of the same host file are not written back to themselves.
    if ! [ "$src" -ef "$KUBECONFIG" ]; then
        if ! mkdir -p "$(dirname "$KUBECONFIG")" 2>/dev/null ||
           ! printf '%s\n' "$KUBECONFIG_CONTENT" > "$KUBECONFIG" 2>/dev/null; then
            echo "ERROR: Cannot write kubeconfig to $KUBECONFIG (uid $(id -u))."
            echo "  Set KUBECONFIG to a writable path, or pre-stage the file there."
            exit 1
        fi
        chmod 600 "$KUBECONFIG" 2>/dev/null || true
    fi
    echo "Kubeconfig loaded from $src"

    # Base64-encode for .env compatibility with E2E tests
    KUBECONFIG_B64=$(printf '%s\n' "$KUBECONFIG_CONTENT" | base64 -w 0)
    export KUBECONFIG_CONTENT
    export KUBECONFIG_B64
}

# ---------------------------------------------------------------
# run_preflight_checks
# Validates kubeconfig YAML, tests cluster connectivity, and
# checks for common namespaces. Exits on critical failures.
# ---------------------------------------------------------------
run_preflight_checks() {
    echo ""
    echo "Validating kubeconfig..."
    if [ ! -f "$KUBECONFIG" ]; then
        echo "CRITICAL: No kubeconfig file found at $KUBECONFIG"
        exit 1
    fi

    if ! kubectl config view --minify > /dev/null 2>&1; then
        echo "CRITICAL: Kubeconfig file is not valid YAML"
        echo "Please check your kubeconfig format"
        exit 1
    fi
    echo "Kubeconfig file structure is valid"

    echo "Testing cluster connectivity..."
    if kubectl cluster-info --request-timeout=10s > /dev/null 2>&1; then
        echo "Successfully connected to Kubernetes cluster"

        echo ""
        echo "Cluster Information:"
        kubectl cluster-info 2>/dev/null | head -5
        echo ""

        echo "Verifying namespace access..."
        if kubectl get namespaces --request-timeout=10s > /dev/null 2>&1; then
            namespace_count=$(kubectl get namespaces --no-headers 2>/dev/null | wc -l)
            echo "Can list namespaces (found $namespace_count)"
        else
            echo "Warning: Cannot list namespaces - permissions may be limited"
        fi

        if kubectl get namespace runai --request-timeout=10s > /dev/null 2>&1; then
            echo "Run:AI namespace 'runai' found"
        else
            echo "Warning: Run:AI namespace 'runai' not found"
            echo "  Some tests may fail if Run:AI is not installed"
        fi

        echo ""
        echo "Pre-flight validation PASSED - proceeding with tests"
        echo ""
    else
        echo "CRITICAL: Cannot connect to Kubernetes cluster"
        echo ""
        echo "Troubleshooting steps:"
        echo "1. Verify your kubeconfig is correct:"
        echo "   kubectl --kubeconfig=$KUBECONFIG cluster-info"
        echo ""
        echo "2. Check certificate data is not corrupted (no extra line breaks)"
        echo ""
        echo "3. Ensure cluster URL is reachable from this container"
        echo "   Cluster URL: $CLUSTER_URL"
        echo ""
        echo "4. Try mounting kubeconfig as file instead of pasting:"
        echo "   docker run -v /path/to/kubeconfig:/kubeconfig/config:ro ..."
        echo ""
        echo "Detailed error:"
        kubectl cluster-info 2>&1 | head -10
        echo ""
        echo "Exiting - cannot proceed without cluster connectivity"
        exit 1
    fi
}

# ---------------------------------------------------------------
# write_env_file
# Writes /app/e2e/.env from the current environment variables.
# Requires KUBECONFIG_CONTENT and KUBECONFIG_B64 to be set
# (call load_kubeconfig first).
# ---------------------------------------------------------------
write_env_file() {
    echo "Updating .env file with configuration..."
    cat > /app/e2e/.env << EOF
# Run:AI Kit Configuration
# Updated from environment variables

# Control Plane Configuration
CONTROL_PLANE_URL=${CONTROL_PLANE_URL}
APP_URL=${CONTROL_PLANE_URL}
CONTROL_PLANE_ADMIN_USERNAME=${CONTROL_PLANE_ADMIN_USERNAME:-admin@example.com}
CONTROL_PLANE_ADMIN_PASSWORD=${CONTROL_PLANE_ADMIN_PASSWORD:-Abcd!234}

# Cluster Configuration
CLUSTER_NAME_PREFIX=${CLUSTER_NAME_PREFIX}
KUBECONFIG_CLUSTER_NAME=${CLUSTER_NAME_PREFIX}
CLUSTER_URL=${CLUSTER_URL}
KUBERNETES_PLATFORM=${KUBERNETES_PLATFORM:-k8s}

# Test Configuration
LOG_LEVEL=${LOG_LEVEL:-info}
GPU_TYPE=${GPU_TYPE:-real}
TASK_TIMEOUT_MS=${TASK_TIMEOUT_MS:-300000}
HEADLESS_BROWSER=${HEADLESS_BROWSER:-true}
ENABLE_AUTH=${ENABLE_AUTH:-true}
LOCAL_RUN=${LOCAL_RUN:-false}

# Optional environment variables with defaults
NODE_TLS_REJECT_UNAUTHORIZED=${NODE_TLS_REJECT_UNAUTHORIZED:-0}
IS_AIRGAPPED=${IS_AIRGAPPED:-false}
RETRIES=${RETRIES:-0}

# Fallback values to prevent errors
E2E_TENANT_CLIENT_SECRET_STAGING=${E2E_TENANT_CLIENT_SECRET_STAGING:-dummy-secret}
E2E_TENANT_CLIENT_SECRET_TEST=${E2E_TENANT_CLIENT_SECRET_TEST:-dummy-secret}

# Allure Reporting
ALLURE_RESULTS_DIR=${ALLURE_RESULTS_DIR:-allure-results}
ALLURE_REPORT_DIR=${ALLURE_REPORT_DIR:-allure-report}
ALLURE_GENERATE_REPORT=${ALLURE_GENERATE_REPORT:-true}
ALLURE_SERVE_REPORT=${ALLURE_SERVE_REPORT:-false}
ALLURE_PORT=${ALLURE_PORT:-8080}

# Kubeconfig Configuration
KUBECONFIG_CONTENT=${KUBECONFIG_CONTENT}
KUBECONFIG_B64=${KUBECONFIG_B64}
EOF
    echo ".env file updated"
}

# ---------------------------------------------------------------
# set_default_env_vars
# Exports all optional env vars with their defaults, so they are
# available to the Playwright process even if not passed to docker.
# ---------------------------------------------------------------
set_default_env_vars() {
    echo "Setting required environment variables..."

    export CONTROL_PLANE_URL="${CONTROL_PLANE_URL:-https://localhost.example.com}"
    export APP_URL="${APP_URL:-$CONTROL_PLANE_URL}"
    export CLUSTER_NAME_PREFIX="${CLUSTER_NAME_PREFIX:-example-cluster}"
    export KUBECONFIG_CLUSTER_NAME="${KUBECONFIG_CLUSTER_NAME:-example-cluster}"
    export KUBERNETES_PLATFORM="${KUBERNETES_PLATFORM:-k8s}"
    export CONTROL_PLANE_ADMIN_USERNAME="${CONTROL_PLANE_ADMIN_USERNAME:-admin@example.com}"
    export CONTROL_PLANE_ADMIN_PASSWORD="${CONTROL_PLANE_ADMIN_PASSWORD:-Abcd!234}"

    export NODE_TLS_REJECT_UNAUTHORIZED="${NODE_TLS_REJECT_UNAUTHORIZED:-0}"
    export IS_AIRGAPPED="${IS_AIRGAPPED:-false}"
    export LOG_LEVEL="${LOG_LEVEL:-info}"
    export GPU_TYPE="${GPU_TYPE:-real}"
    export TASK_TIMEOUT_MS="${TASK_TIMEOUT_MS:-300000}"
    export HEADLESS_BROWSER="${HEADLESS_BROWSER:-true}"
    export ENABLE_AUTH="${ENABLE_AUTH:-true}"
    export LOCAL_RUN="${LOCAL_RUN:-true}"
    export RETRIES="${RETRIES:-1}"

    export E2E_TENANT_CLIENT_SECRET_STAGING="${E2E_TENANT_CLIENT_SECRET_STAGING:-dummy-secret}"
    export E2E_TENANT_CLIENT_SECRET_TEST="${E2E_TENANT_CLIENT_SECRET_TEST:-dummy-secret}"
    export CLUSTER_URL="${CLUSTER_URL:-https://$CLUSTER_NAME_PREFIX.example.com}"

    export ALLURE_RESULTS_DIR="${ALLURE_RESULTS_DIR:-allure-results}"
    export ALLURE_REPORT_DIR="${ALLURE_REPORT_DIR:-allure-report}"
    export ALLURE_GENERATE_REPORT="${ALLURE_GENERATE_REPORT:-true}"
    export ALLURE_SERVE_REPORT="${ALLURE_SERVE_REPORT:-false}"
    export ALLURE_PORT="${ALLURE_PORT:-8080}"

    echo "Environment variables set:"
    echo "  CONTROL_PLANE_URL: $CONTROL_PLANE_URL"
    echo "  APP_URL: $APP_URL"
    echo "  CLUSTER_NAME_PREFIX: $CLUSTER_NAME_PREFIX"
    echo "  KUBERNETES_PLATFORM: $KUBERNETES_PLATFORM"
    echo "  CONTROL_PLANE_ADMIN_USERNAME: $CONTROL_PLANE_ADMIN_USERNAME"
    echo "  LOG_LEVEL: $LOG_LEVEL"
    echo "  GPU_TYPE: $GPU_TYPE"
    echo "  RETRIES: $RETRIES"
}

# ---------------------------------------------------------------
# run_playwright_tests <grep-tag> <config-file>
# Runs Playwright with the given grep tag and config.
# NODE_OPTIONS is set and unset around the call.
# ---------------------------------------------------------------
run_playwright_tests() {
    local grep_tag="$1"
    local config_file="$2"

    echo "Running tests tagged with ${grep_tag}..."
    echo "Working directory: $(pwd)"
    echo "Config: ${config_file}"

    export NODE_OPTIONS="--openssl-legacy-provider --loader ts-node/esm"
    npx playwright test --grep="${grep_tag}" --config="${config_file}"
    local exit_code=$?
    unset NODE_OPTIONS

    echo "Checking test results..."
    if [ -d "allure-results" ]; then
        echo "Allure results directory found: $(ls -la allure-results/)"
    else
        echo "Warning: Allure results directory not found"
        echo "Current directory contents: $(ls -la)"
        find . -name "allure-results" -type d 2>/dev/null || echo "No allure-results found anywhere"
    fi

    return $exit_code
}

# ---------------------------------------------------------------
# check_and_generate_allure_report
# Generates the Allure HTML report from allure-results/.
# Falls back to /app/allure-results if the local dir is missing.
# ---------------------------------------------------------------
check_and_generate_allure_report() {
    echo "Generating Allure report..."
    # Unset NODE_OPTIONS so ts-node loader doesn't interfere with Allure (Java)
    unset NODE_OPTIONS

    if [ -d "allure-results" ]; then
        allure generate allure-results --clean -o allure-report
        echo "Allure report generated successfully"
    elif [ -d "/app/allure-results" ]; then
        echo "Found results in /app/allure-results, generating from there..."
        allure generate /app/allure-results --clean -o /app/allure-report
    else
        echo "ERROR: No Allure results found anywhere"
        exit 1
    fi
}

# ---------------------------------------------------------------
# copy_allure_report_to_results
# Copies the generated Allure report into the mounted results volume.
# ---------------------------------------------------------------
copy_allure_report_to_results() {
    if [ -d /app/e2e/results ] && [ -w /app/e2e/results ]; then
        echo "Copying Allure report to results directory..."
        if [ -d "allure-report" ]; then
            cp -r allure-report /app/e2e/results/
            echo "Allure report copied to results directory"
        elif [ -d "/app/allure-report" ]; then
            cp -r /app/allure-report /app/e2e/results/
            echo "Allure report copied from /app/allure-report"
        else
            echo "ERROR: No Allure report found to copy"
        fi

        echo "To view the report:"
        echo "  1. Navigate to the results/allure-report directory"
        echo "  2. Open index.html in your browser"
        echo "  3. Or run: cd results/allure-report && python3 -m http.server 8080"
    elif [ -d /app/e2e/results ]; then
        echo "ERROR: /app/e2e/results is not writable by uid $(id -u)."
        echo "  The report and archive cannot be written to the mounted volume."
        ls -ldn /app/e2e/results
    else
        echo "Results directory not mounted, report available at /app/e2e/allure-report in container"
    fi
}

# ---------------------------------------------------------------
# package_results <kit-name> <log-filename> <archive-prefix>
# Creates an archive (ZIP or TAR.GZ) of the results directory
# including the Allure report, logs, and metadata files.
# Must be called from inside the results directory.
# ---------------------------------------------------------------
package_results() {
    local kit_name="$1"
    local log_filename="$2"
    local archive_prefix="$3"

    echo ""
    echo "Packaging results for investigation..."

    if [ ! -d /app/e2e/results ]; then
        echo "Warning: Results directory not mounted, skipping package creation"
        return
    fi

    if [ ! -w /app/e2e/results ]; then
        echo "ERROR: /app/e2e/results is not writable by uid $(id -u)."
        echo "  The report and archive cannot be written to the mounted volume."
        ls -ldn /app/e2e/results
        return 1
    fi

    cd /app/e2e/results

    # Generate executive summary
    echo "Generating executive summary..."
    cat > executive-summary.md << EOF
# Run:AI ${kit_name} - Executive Summary

## Test Execution Overview
- **Execution Date**: $(date '+%Y-%m-%d %H:%M:%S UTC')
- **Cluster**: ${CLUSTER_NAME_PREFIX}
- **Control Plane**: ${CONTROL_PLANE_URL}
- **Platform**: ${KUBERNETES_PLATFORM}

## Test Results Summary
- **Total Tests**: $(find allure-report/data/test-cases -name "*.json" 2>/dev/null | wc -l)
- **Passed**: $(grep -r '"status":"passed"' allure-report/data/test-cases/ 2>/dev/null | wc -l)
- **Failed**: $(grep -r '"status":"failed"' allure-report/data/test-cases/ 2>/dev/null | wc -l)
- **Skipped**: $(grep -r '"status":"skipped"' allure-report/data/test-cases/ 2>/dev/null | wc -l)

## Environment Information
- **Cluster Name**: ${CLUSTER_NAME_PREFIX}
- **Control Plane URL**: ${CONTROL_PLANE_URL}
- **Kubernetes Platform**: ${KUBERNETES_PLATFORM}
- **Test Environment**: ${kit_name} Container

## Package Contents
- **Allure HTML Report**: Complete interactive test report
- **Raw Test Data**: JSON files with detailed test information
- **Test Attachments**: Logs, screenshots, and error details
- **Executive Summary**: This file with high-level overview

## How to Use This Package
1. **For Run:AI Team**: Extract the archive and open \`allure-report/index.html\`
2. **For Analysis**: Use the JSON files in \`allure-report/data/\` for programmatic analysis
3. **For Investigation**: Check \`allure-report/data/attachments/\` for detailed logs and errors
EOF

    # Generate CSV summary
    echo "Generating test results summary CSV..."
    cat > test-results-summary.csv << 'CSVHEADER'
Test Name,Status,Duration (ms),Tags,Error Details
CSVHEADER

    if [ -d "allure-report/data/test-cases" ]; then
        for test_file in allure-report/data/test-cases/*.json; do
            if [ -f "$test_file" ]; then
                test_name=$(grep -o '"name":"[^"]*"' "$test_file" | head -1 | sed 's/"name":"//;s/"//')
                status=$(grep -o '"status":"[^"]*"' "$test_file" | head -1 | sed 's/"status":"//;s/"//')
                duration=$(grep -o '"duration":[0-9]*' "$test_file" | head -1 | sed 's/"duration"://')
                tags_val=$(grep -o '"tag":"[^"]*"' "$test_file" | sed 's/"tag":"//;s/"//' | tr '\n' ';' | sed 's/;$//')
                echo "\"$test_name\",\"$status\",\"$duration\",\"$tags_val\"" >> test-results-summary.csv
            fi
        done
    fi

    # Create README
    echo "Creating package README..."
    cat > README.md << EOF
# Run:AI ${kit_name} Results Package

This package contains the complete results from a Run:AI ${kit_name} execution.

## Quick Start
1. **Open the HTML Report**: Navigate to \`allure-report/index.html\` and open in your browser
2. **View Executive Summary**: Read \`executive-summary.md\` for a high-level overview
3. **Analyze Test Data**: Use \`test-results-summary.csv\` for spreadsheet analysis

## Package Structure
\`\`\`
results-package/
├── README.md                    # This file
├── executive-summary.md         # High-level test summary
├── test-results-summary.csv     # Test results in CSV format
├── ${log_filename}              # Complete console output from the run
├── allure-report/               # Complete Allure HTML report
│   ├── index.html               # Main report entry point
│   ├── data/                    # Raw test data (JSON)
│   └── attachments/             # Test logs and screenshots
└── environment-info.json        # Test environment details
\`\`\`
EOF

    # Create environment-info.json
    echo "Creating environment information..."
    cat > environment-info.json << EOF
{
  "execution_date": "$(date -u '+%Y-%m-%dT%H:%M:%SZ')",
  "cluster_name": "${CLUSTER_NAME_PREFIX}",
  "control_plane_url": "${CONTROL_PLANE_URL}",
  "kubernetes_platform": "${KUBERNETES_PLATFORM}",
  "test_environment": "${kit_name}",
  "package_version": "1.0"
}
EOF

    # Final log sync before archiving
    echo "Final log sync before archiving..."
    sync_logs

    if [ -f "$RESULTS_LOG_FILE" ]; then
        LOG_SIZE=$(stat -f%z "$RESULTS_LOG_FILE" 2>/dev/null || stat -c%s "$RESULTS_LOG_FILE" 2>/dev/null || echo "0")
        echo "Run log ready for archiving (size: ${LOG_SIZE} bytes)"
    else
        echo "Warning: Run log file not found, creating placeholder..."
        echo "Log capture failed - this is a placeholder file" > "$RESULTS_LOG_FILE"
    fi

    # Create the archive
    ARCHIVE_NAME="${archive_prefix}-$(date +%Y%m%d-%H%M%S)"

    if command -v zip >/dev/null 2>&1; then
        zip -r "${ARCHIVE_NAME}.zip" . \
            -x "*.zip" \
            -x "allure-report/export/*" \
            -x "allure-report/history/*"
        ARCHIVE_RESULT=$?

        if [ $ARCHIVE_RESULT -eq 0 ]; then
            echo "ZIP package created successfully!"
            ls -la *.zip
        else
            echo "Warning: ZIP creation had issues, falling back to tar.gz..."
            tar -czf "${ARCHIVE_NAME}.tar.gz" \
                --exclude="*.zip" \
                --exclude="*.tar.gz" \
                --exclude="allure-report/export" \
                --exclude="allure-report/history" .
            echo "TAR.GZ package created successfully!"
            ls -la *.tar.gz
        fi
    else
        echo "zip not available, creating tar.gz..."
        tar -czf "${ARCHIVE_NAME}.tar.gz" \
            --exclude="*.zip" \
            --exclude="*.tar.gz" \
            --exclude="allure-report/export" \
            --exclude="allure-report/history" .
        echo "TAR.GZ package created successfully!"
        ls -la *.tar.gz
    fi

    # Verify log is in archive
    echo ""
    echo "Verifying log file is included in archive..."
    if [ -f "${ARCHIVE_NAME}.zip" ]; then
        if unzip -l "${ARCHIVE_NAME}.zip" | grep -q "${log_filename}"; then
            echo "Run log verified in ZIP archive"
        else
            echo "Warning: Run log may not be in the archive"
        fi
    elif [ -f "${ARCHIVE_NAME}.tar.gz" ]; then
        if tar -tzf "${ARCHIVE_NAME}.tar.gz" | grep -q "${log_filename}"; then
            echo "Run log verified in TAR.GZ archive"
        else
            echo "Warning: Run log may not be in the archive"
        fi
    fi

    echo ""
    echo "Results Package Ready!"
    echo "Package contains:"
    echo "  - Executive Summary (executive-summary.md)"
    echo "  - Test Results Summary (test-results-summary.csv)"
    echo "  - Complete Run Log (${log_filename})"
    echo "  - Complete Allure HTML Report (allure-report/)"
    echo "  - Environment Information (environment-info.json)"
    echo "  - Package Documentation (README.md)"
    echo ""
    echo "Package location: $(pwd)/"
}
