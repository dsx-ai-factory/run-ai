#!/bin/bash
set -e

echo "=========================================="
echo "Run:AI Unified Scale Tests"
echo "Version: ${VERSION:-v1.0.0}"
echo "=========================================="

# Validate kubeconfig exists (check both locations for compatibility)
KUBECONFIG_PATH="/app/kubeconfig/config"
if [ ! -f "$KUBECONFIG_PATH" ]; then
    echo "Error: kubeconfig not found at $KUBECONFIG_PATH"
    echo "Please mount your kubeconfig: -v ~/.kube/config:/app/kubeconfig/config:ro"
    exit 1
fi

# Set KUBECONFIG environment variable
export KUBECONFIG="$KUBECONFIG_PATH"

# Source configuration file
if [ -f /app/config/test-config.env ]; then
    echo "Loading configuration from /app/config/test-config.env"
    set -a  # Export all variables
    source /app/config/test-config.env
    set +a
else
    echo "Error: Configuration file not found at /app/config/test-config.env"
    echo "Please mount your config file: -v ./test-config.env:/app/config/test-config.env:ro"
    exit 1
fi

# Display configuration summary
echo ""
echo "Test Configuration:"
echo "  Projects: ${PROJECT_COUNT:-N/A}"
echo "  Workloads: ${WORKLOAD_COUNT:-N/A}"
echo "  Resources: ${RESOURCES:-N/A}"
echo "  KWOK Enabled: ${KWOK_ENABLED:-false}"
echo "=========================================="
echo ""

# Generate timestamp for results
TIMESTAMP=$(date +"%Y-%m-%d-%H-%M-%S")
RESULTS_FILE="/app/results/test-results-${TIMESTAMP}.json"
RAW_DATA_FILE="/tmp/workload_raw_data.json"
METRICS_REPORT_FILE="/tmp/metrics_report.json"

# Validate workload-scale binary exists
if [ ! -x /app/bin/workload-scale ]; then
    echo "Error: workload-scale binary not found or not executable"
    exit 1
fi

# Run workload-scale binary
echo "Starting test execution..."
set +e  # Temporarily disable exit-on-error to capture exit code
/app/bin/workload-scale
TEST_EXIT_CODE=$?
set -e  # Re-enable exit-on-error

if [ "${TEST_EXIT_CODE:-0}" -ne 0 ]; then
    echo "Error: Test execution failed with exit code $TEST_EXIT_CODE"
fi

# Generate report if raw data exists
if [ -f "$RAW_DATA_FILE" ]; then
    echo ""
    echo "Generating metrics report..."
    export TEST_RESULTS_FILE="$RAW_DATA_FILE"
    export OUTPUT_FILE="$METRICS_REPORT_FILE"
    /app/bin/report-generator || echo "Warning: Report generation encountered errors"
else
    echo "Warning: No raw data file found at $RAW_DATA_FILE"
fi

# Save results to timestamped file
if [ -f "$METRICS_REPORT_FILE" ]; then
    mkdir -p /app/results || { echo "Error: Cannot create results directory"; exit 1; }
    cp "$METRICS_REPORT_FILE" "$RESULTS_FILE" || { echo "Error: Failed to save results"; exit 1; }
    echo ""
    echo "Results saved to: $(basename "$RESULTS_FILE")"
    echo ""
    echo "=========================================="
    echo "TEST RESULTS SUMMARY:"
    echo "=========================================="
    cat "$METRICS_REPORT_FILE"
    echo ""
else
    echo "Warning: No metrics report generated"
fi

# Run cleanup
echo "=========================================="
echo "Cleaning up test resources..."
/app/bin/clean-cluster || echo "Warning: Cleanup encountered errors"

echo "=========================================="
echo "Test execution completed"
echo "Full results available in: $(basename "$RESULTS_FILE")"
echo "=========================================="

exit "${TEST_EXIT_CODE:-0}"
