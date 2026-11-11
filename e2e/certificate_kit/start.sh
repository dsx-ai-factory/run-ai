#!/bin/bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


# ============================================================
# LOG CAPTURE SETUP
# Capture all output to a log file while still showing on console
# Uses line buffering to ensure logs are written immediately
# ============================================================
LOG_FILE="/tmp/certification-kit-run.log"
RESULTS_LOG_FILE="/app/e2e/results/certification-kit-run.log"

# Clean up any existing log file
> "$LOG_FILE"

# Create a named pipe for capturing output with line buffering
# Using stdbuf to reduce buffering issues
if command -v stdbuf >/dev/null 2>&1; then
    exec > >(stdbuf -oL tee -a "$LOG_FILE") 2>&1
else
    exec > >(tee -a "$LOG_FILE") 2>&1
fi

# Function to sync and copy logs to results directory
sync_logs() {
    # Flush stdio buffers
    sync 2>/dev/null || true
    sleep 0.5
    
    # Copy log to results directory if it exists
    if [ -d "/app/e2e/results" ] && [ -f "$LOG_FILE" ]; then
        cp "$LOG_FILE" "$RESULTS_LOG_FILE" 2>/dev/null || true
    fi
}

# Set up trap to sync logs on exit (ensures logs are captured even on errors)
trap 'sync_logs' EXIT

echo "🚀 Starting Run:AI Certification Kit..."
echo "📝 All output is being logged to: $LOG_FILE"

# Skip .env file loading - we require environment variables to be passed
echo "📋 Configuration from environment variables..."

# Validate all required parameters are provided
echo "🔍 Validating required parameters..."
echo ""

# Check for required environment variables
if [ -z "$CONTROL_PLANE_URL" ]; then
    echo "❌ ERROR: CONTROL_PLANE_URL environment variable is required"
    echo ""
    echo "Please run with:"
    echo "  docker run -e CONTROL_PLANE_URL=https://your-control-plane.com ..."
    exit 1
fi

if [ -z "$CLUSTER_NAME_PREFIX" ]; then
    echo "❌ ERROR: CLUSTER_NAME_PREFIX environment variable is required"
    echo ""
    echo "Please run with:"
    echo "  docker run -e CLUSTER_NAME_PREFIX=your-cluster-name ..."
    exit 1
fi

if [ -z "$CLUSTER_URL" ]; then
    echo "❌ ERROR: CLUSTER_URL environment variable is required"
    echo ""
    echo "Please run with:"
    echo "  docker run -e CLUSTER_URL=https://your-k8s-api.com ..."
    exit 1
fi

# Check for mounted kubeconfig
if [ ! -f /kubeconfig/config ]; then
    echo "❌ ERROR: Kubeconfig file not found at /kubeconfig/config"
    echo ""
    echo "Please mount your kubeconfig file:"
    echo "  docker run -v /path/to/your/kubeconfig:/kubeconfig/config:ro ..."
    exit 1
fi

# All required parameters present
echo "✅ All required parameters validated"
echo "   CONTROL_PLANE_URL: $CONTROL_PLANE_URL"
echo "   CLUSTER_NAME_PREFIX: $CLUSTER_NAME_PREFIX"
echo "   CLUSTER_URL: $CLUSTER_URL"
echo "   Kubeconfig: mounted at /kubeconfig/config"
echo ""

# Load kubeconfig from mounted file
KUBECONFIG_CONTENT=$(cat /kubeconfig/config)


# Update the .env file with user inputs
echo "📝 Updating .env file with user inputs..."
cat > /app/e2e/.env << EOF
# Run:AI Certification Kit Configuration
# Updated from user inputs

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
GPU_TYPE=${GPU_TYPE:-fake}
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
echo "✅ .env file updated with user inputs"
echo ""

# Setup kubeconfig - use user-provided content
echo "🔧 Setting up kubeconfig..."
export KUBECONFIG="${KUBECONFIG:-${HOME:-/root}/.kube/config}"
if [ -n "$KUBECONFIG_CONTENT" ]; then
    echo "📄 Using kubeconfig content provided by user"
    mkdir -p "$(dirname "$KUBECONFIG")"
    # With the legacy direct mount (-v kubeconfig:/kubeconfig/config:ro) Docker
    # resolves the image's /kubeconfig/config symlink and parks the read-only
    # bind mount on $KUBECONFIG itself — the content is already in place, and
    # writing to it would fail with EROFS. Only write when the target is
    # writable (the `make run` volume flow, or a plain filesystem path).
    if [ -w "$KUBECONFIG" ] || [ ! -e "$KUBECONFIG" ]; then
        printf '%s\n' "$KUBECONFIG_CONTENT" > "${KUBECONFIG:-${HOME:-/root}/.kube/config}"
        chmod 600 "${KUBECONFIG:-${HOME:-/root}/.kube/config}"
        echo "✅ Kubeconfig loaded from user-provided content"
    else
        echo "✅ Kubeconfig already mounted read-only at $KUBECONFIG — using it as-is"
    fi
    
    # Convert to base64 for .env file (for compatibility with existing E2E tests)
    KUBECONFIG_B64=$(printf '%s\n' "$KUBECONFIG_CONTENT" | base64 -w 0)
    echo "📝 Converted kubeconfig content to base64 for .env compatibility"
else
    echo "⚠️ No kubeconfig content provided, kubectl may not work properly"
    echo "   Please provide kubeconfig content when prompted"
fi

# Validate kubeconfig before proceeding
echo ""
echo "🔍 Validating kubeconfig..."
if [ ! -f "${KUBECONFIG:-${HOME:-/root}/.kube/config}" ]; then
    echo "❌ CRITICAL: No kubeconfig file found at ${KUBECONFIG:-${HOME:-/root}/.kube/config}"
    echo "Cannot proceed with certification tests without cluster access"
    exit 1
fi

# Check if kubeconfig has valid YAML syntax
if ! kubectl config view --minify > /dev/null 2>&1; then
    echo "❌ CRITICAL: Kubeconfig file is not valid YAML"
    echo "Please check your kubeconfig format"
    exit 1
fi

echo "✅ Kubeconfig file structure is valid"

# Test cluster connectivity with detailed error reporting
echo "🔍 Testing cluster connectivity..."
if kubectl cluster-info --request-timeout=10s > /dev/null 2>&1; then
    echo "✅ Successfully connected to Kubernetes cluster"
    
    # Show basic cluster info
    echo ""
    echo "📊 Cluster Information:"
    kubectl cluster-info 2>/dev/null | head -5
    echo ""
    
    # Test namespace access
    echo "🔍 Verifying namespace access..."
    if kubectl get namespaces --request-timeout=10s > /dev/null 2>&1; then
        namespace_count=$(kubectl get namespaces --no-headers 2>/dev/null | wc -l)
        echo "✅ Can list namespaces (found $namespace_count)"
    else
        echo "⚠️  Warning: Cannot list namespaces - permissions may be limited"
    fi
    
    # Check for runai namespace
    if kubectl get namespace runai --request-timeout=10s > /dev/null 2>&1; then
        echo "✅ Run:AI namespace 'runai' found"
    else
        echo "⚠️  Warning: Run:AI namespace 'runai' not found"
        echo "   Some tests may fail if Run:AI is not installed"
    fi
    
    echo ""
    echo "✅ Pre-flight validation PASSED - proceeding with certification tests"
    echo ""
else
    echo "❌ CRITICAL: Cannot connect to Kubernetes cluster"
    echo ""
    echo "Troubleshooting steps:"
    echo "1. Verify your kubeconfig is correct:"
    echo "   kubectl --kubeconfig=${KUBECONFIG:-${HOME:-/root}/.kube/config} cluster-info"
    echo ""
    echo "2. Check certificate data is not corrupted (no extra line breaks)"
    echo ""
    echo "3. Ensure cluster URL is reachable from this container"
    echo "   Cluster URL: $CLUSTER_URL"
    echo ""
    echo "4. Try mounting kubeconfig as file instead of pasting:"
    echo "   docker run -v /path/to/kubeconfig:/kubeconfig/config:ro ..."
    echo ""
    
    # Show the actual error
    echo "Detailed error:"
    kubectl cluster-info 2>&1 | head -10
    
    echo ""
    echo "❌ Exiting - cannot proceed without cluster connectivity"
    exit 1
fi

# Set required environment variables for E2E tests with defaults
echo "🔧 Setting required environment variables for E2E tests..."

# Set environment variables with defaults if not already set
export CONTROL_PLANE_URL="${CONTROL_PLANE_URL:-https://localhost.example.com}"
export APP_URL="${APP_URL:-$CONTROL_PLANE_URL}"
export CLUSTER_NAME_PREFIX="${CLUSTER_NAME_PREFIX:-example-cluster}"
export KUBECONFIG_CLUSTER_NAME="${KUBECONFIG_CLUSTER_NAME:-example-cluster}"
export KUBERNETES_PLATFORM="${KUBERNETES_PLATFORM:-k8s}"
export CONTROL_PLANE_ADMIN_USERNAME="${CONTROL_PLANE_ADMIN_USERNAME:-admin@example.com}"
export CONTROL_PLANE_ADMIN_PASSWORD="${CONTROL_PLANE_ADMIN_PASSWORD:-admin}"

# Optional environment variables with sensible defaults
export NODE_TLS_REJECT_UNAUTHORIZED="${NODE_TLS_REJECT_UNAUTHORIZED:-0}"
export IS_AIRGAPPED="${IS_AIRGAPPED:-false}"
export LOG_LEVEL="${LOG_LEVEL:-info}"
export GPU_TYPE="${GPU_TYPE:-fake}"
export TASK_TIMEOUT_MS="${TASK_TIMEOUT_MS:-300000}"
export HEADLESS_BROWSER="${HEADLESS_BROWSER:-true}"
export ENABLE_AUTH="${ENABLE_AUTH:-true}"
export LOCAL_RUN="${LOCAL_RUN:-true}"
export RETRIES="${RETRIES:-1}"

# Set defaults for missing variables that cause errors
export E2E_TENANT_CLIENT_SECRET_STAGING="${E2E_TENANT_CLIENT_SECRET_STAGING:-dummy-secret}"
export E2E_TENANT_CLIENT_SECRET_TEST="${E2E_TENANT_CLIENT_SECRET_TEST:-dummy-secret}"
export CLUSTER_URL="${CLUSTER_URL:-https://$CLUSTER_NAME_PREFIX.example.com}"

# Allure reporting configuration
export ALLURE_RESULTS_DIR="${ALLURE_RESULTS_DIR:-allure-results}"
export ALLURE_REPORT_DIR="${ALLURE_REPORT_DIR:-allure-report}"
export ALLURE_GENERATE_REPORT="${ALLURE_GENERATE_REPORT:-true}"
export ALLURE_SERVE_REPORT="${ALLURE_SERVE_REPORT:-false}"
export ALLURE_PORT="${ALLURE_PORT:-8080}"

echo "✅ Environment variables set:"
echo "   CONTROL_PLANE_URL: $CONTROL_PLANE_URL"
echo "   APP_URL: $APP_URL"
echo "   CLUSTER_NAME_PREFIX: $CLUSTER_NAME_PREFIX"
echo "   KUBERNETES_PLATFORM: $KUBERNETES_PLATFORM"
echo "   CONTROL_PLANE_ADMIN_USERNAME: $CONTROL_PLANE_ADMIN_USERNAME"
echo "   LOG_LEVEL: $LOG_LEVEL"
echo "   GPU_TYPE: $GPU_TYPE"
echo "   RETRIES: $RETRIES"

# Ensure we're in the right directory for E2E tests
cd /app/e2e

# Run Playwright tests by tag
echo "🧪 Running certification tests from E2E suite..."
echo "📋 Running tests tagged with @k8s-flav-cert-kit..."
echo "📁 Working directory: $(pwd)"
echo "📁 Test directory: src/test/suites/playwright/"

# Run only tests with the k8s-flav-cert-kit-temp tag using our custom certification kit config
# Set Node.js options to handle TypeScript config files properly and OpenSSL legacy provider
export NODE_OPTIONS="--openssl-legacy-provider --loader ts-node/esm"
npx playwright test --grep="@k8s-flav-cert-kit-temp" --config=playwright-config/certification-kit.config.ts

# Check if tests ran and results were generated
echo "🔍 Checking test results..."
if [ -d "allure-results" ]; then
    echo "✅ Allure results directory found: $(ls -la allure-results/)"
else
    echo "❌ Allure results directory not found!"
    echo "📁 Current directory contents: $(ls -la)"
    echo "📁 Looking for allure-results in subdirectories..."
    find . -name "allure-results" -type d 2>/dev/null || echo "No allure-results found"
fi

# Generate Allure report from the correct location
echo "📊 Generating Allure report..."
if [ -d "allure-results" ]; then
    # Unset NODE_OPTIONS for Allure to avoid ts-node loader conflicts
    echo "   Unsetting NODE_OPTIONS for Allure compatibility..."
    unset NODE_OPTIONS
    
    allure generate allure-results --clean -o allure-report
    echo "✅ Allure report generated successfully"
else
    echo "❌ Cannot generate Allure report - no results found"
    # Try to find results in other locations
    if [ -d "/app/allure-results" ]; then
        echo "📁 Found results in /app/allure-results, generating report from there..."
        allure generate /app/allure-results --clean -o /app/allure-report
    else
        echo "❌ No Allure results found anywhere!"
        exit 1
    fi
fi

# Copy report to results directory if mounted
if [ -d /app/e2e/results ]; then
    echo "📁 Copying Allure report to results directory..."
    if [ -d "allure-report" ]; then
        cp -r allure-report /app/e2e/results/
        echo "✅ Allure report copied to results directory"
    elif [ -d "/app/allure-report" ]; then
        cp -r /app/allure-report /app/e2e/results/
        echo "✅ Allure report copied to results directory from /app/allure-report"
    else
        echo "❌ No Allure report found to copy!"
    fi
    
    echo "🌐 To view the report:"
    echo "   1. Navigate to the results/allure-report directory"
    echo "   2. Open index.html in your browser"
    echo "   3. Or run: cd results/allure-report && python3 -m http.server 8080"
else
    echo "📁 Results directory not mounted, report available at /app/e2e/allure-report in container"
fi

# Package results for Run:AI team investigation
echo ""
echo "📦 Packaging results for Run:AI team investigation..."
if [ -d /app/e2e/results ]; then
    cd /app/e2e/results
    
    # Sync logs early in the packaging process
    echo "📝 Syncing run log to results directory..."
    sync_logs
    if [ -f "$RESULTS_LOG_FILE" ]; then
        echo "✅ Initial log sync completed"
    else
        echo "⚠️ Initial log sync - file not yet created"
    fi
    
    # Create a comprehensive results package
    echo "📋 Creating results package..."
    
    # Generate executive summary
    echo "📊 Generating executive summary..."
    cat > executive-summary.md << 'EOF'
# Run:AI Certification Kit - Executive Summary

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
- **Test Environment**: Certification Kit Container

## Package Contents
- **Allure HTML Report**: Complete interactive test report
- **Raw Test Data**: JSON files with detailed test information
- **Test Attachments**: Logs, screenshots, and error details
- **Executive Summary**: This file with high-level overview

## How to Use This Package
1. **For Run:AI Team**: Extract the ZIP file and open `allure-report/index.html`
2. **For Analysis**: Use the JSON files in `allure-report/data/` for programmatic analysis
3. **For Investigation**: Check `allure-report/data/attachments/` for detailed logs and errors

## Contact Information
- **Generated by**: Run:AI Certification Kit
- **For Support**: Contact Run:AI Support Team
EOF

    # Generate test results summary CSV
    echo "📊 Generating test results summary..."
    cat > test-results-summary.csv << 'EOF'
Test Name,Status,Duration (ms),Tags,Error Details
EOF

    # Extract test results from JSON files
    if [ -d "allure-report/data/test-cases" ]; then
        for test_file in allure-report/data/test-cases/*.json; do
            if [ -f "$test_file" ]; then
                # Extract test information using jq-like parsing (basic sed/awk approach)
                test_name=$(grep -o '"name":"[^"]*"' "$test_file" | head -1 | sed 's/"name":"//;s/"//')
                status=$(grep -o '"status":"[^"]*"' "$test_file" | head -1 | sed 's/"status":"//;s/"//')
                duration=$(grep -o '"duration":[0-9]*' "$test_file" | head -1 | sed 's/"duration"://')
                tags=$(grep -o '"tag":"[^"]*"' "$test_file" | sed 's/"tag":"//;s/"//' | tr '\n' ';' | sed 's/;$//')
                
                # Add to CSV
                echo "\"$test_name\",\"$status\",\"$duration\",\"$tags\"" >> test-results-summary.csv
            fi
        done
    fi

    # Create README for the package
    echo "📖 Creating package README..."
    cat > README.md << 'EOF'
# Run:AI Certification Kit Results Package

This package contains the complete results from a Run:AI Certification Kit execution.

## Quick Start
1. **Open the HTML Report**: Navigate to `allure-report/index.html` and open in your browser
2. **View Executive Summary**: Read `executive-summary.md` for a high-level overview
3. **Analyze Test Data**: Use `test-results-summary.csv` for spreadsheet analysis

## Package Structure
```
results-package/
├── README.md                    # This file
├── executive-summary.md         # High-level test summary
├── test-results-summary.csv     # Test results in CSV format
├── certification-kit-run.log   # Complete console output from the run
├── allure-report/              # Complete Allure HTML report
│   ├── index.html              # Main report entry point
│   ├── data/                   # Raw test data (JSON)
│   └── attachments/            # Test logs and screenshots
└── environment-info.json       # Test environment details
EOF

    # Create environment info JSON
    echo "🔧 Creating environment information..."
    cat > environment-info.json << EOF
{
  "execution_date": "$(date -u '+%Y-%m-%dT%H:%M:%SZ')",
  "cluster_name": "${CLUSTER_NAME_PREFIX}",
  "control_plane_url": "${CONTROL_PLANE_URL}",
  "kubernetes_platform": "${KUBERNETES_PLATFORM}",
  "test_environment": "certification-kit",
  "package_version": "1.0"
}
EOF

    # Final update to log file before archiving (capture as much as possible)
    echo "📝 Final log sync before archiving..."
    sync_logs
    
    # Verify the log file is present and has content
    if [ -f "$RESULTS_LOG_FILE" ]; then
        LOG_SIZE=$(stat -f%z "$RESULTS_LOG_FILE" 2>/dev/null || stat -c%s "$RESULTS_LOG_FILE" 2>/dev/null || echo "0")
        echo "✅ Run log ready for archiving (size: ${LOG_SIZE} bytes)"
    else
        echo "⚠️ Warning: Run log file not found, creating placeholder..."
        echo "Log capture failed - this is a placeholder file" > "$RESULTS_LOG_FILE"
    fi
    
    # Create the archive package name
    ARCHIVE_NAME="runai-certification-results-$(date +%Y%m%d-%H%M%S)"
    
    # Create the ZIP package
    echo "🗜️ Creating ZIP package..."
    if command -v zip >/dev/null 2>&1; then
        # Final log sync right before creating archive
        sync_logs
        
        zip -r "${ARCHIVE_NAME}.zip" . \
            -x "*.zip" \
            -x "allure-report/export/*" \
            -x "allure-report/history/*"
        ARCHIVE_RESULT=$?
        
        if [ $ARCHIVE_RESULT -eq 0 ]; then
            echo "✅ ZIP package created successfully!"
            ls -la *.zip
        else
            echo "⚠️ ZIP creation had issues, trying tar.gz as fallback..."
            tar -czf "${ARCHIVE_NAME}.tar.gz" \
                --exclude="*.zip" \
                --exclude="*.tar.gz" \
                --exclude="allure-report/export" \
                --exclude="allure-report/history" .
            echo "✅ TAR.GZ package created successfully!"
            ls -la *.tar.gz
        fi
    else
        echo "⚠️ ZIP command not available, creating tar.gz instead..."
        # Final log sync right before creating archive
        sync_logs
        
        tar -czf "${ARCHIVE_NAME}.tar.gz" \
            --exclude="*.zip" \
            --exclude="*.tar.gz" \
            --exclude="allure-report/export" \
            --exclude="allure-report/history" .
        echo "✅ TAR.GZ package created successfully!"
        
        # List the created package
        ls -la *.tar.gz
    fi
    
    # Verify log is included in the archive
    echo ""
    echo "🔍 Verifying log file is included in archive..."
    if [ -f "${ARCHIVE_NAME}.zip" ]; then
        if unzip -l "${ARCHIVE_NAME}.zip" | grep -q "certification-kit-run.log"; then
            echo "✅ Run log verified in ZIP archive"
        else
            echo "⚠️ Warning: Run log may not be in the archive"
        fi
    elif [ -f "${ARCHIVE_NAME}.tar.gz" ]; then
        if tar -tzf "${ARCHIVE_NAME}.tar.gz" | grep -q "certification-kit-run.log"; then
            echo "✅ Run log verified in TAR.GZ archive"
        else
            echo "⚠️ Warning: Run log may not be in the archive"
        fi
    fi

    echo ""
    echo "🎯 Results Package Ready for Run:AI Team!"
    echo "📦 Package contains:"
    echo "   • Executive Summary (executive-summary.md)"
    echo "   • Test Results Summary (test-results-summary.csv)"
    echo "   • Complete Run Log (certification-kit-run.log)"
    echo "   • Complete Allure HTML Report (allure-report/)"
    echo "   • Environment Information (environment-info.json)"
    echo "   • Package Documentation (README.md)"
    echo ""
    echo "📤 Ready to send to Run:AI team for investigation!"
    echo "📁 Package location: $(pwd)/"
    
else
    echo "⚠️ Results directory not mounted, skipping package creation"
fi

echo ""
echo "🎉 Run:AI Certification Kit completed!"
echo "📊 Check the generated Allure report for detailed results."

# Final log sync to capture completion message
sync_logs