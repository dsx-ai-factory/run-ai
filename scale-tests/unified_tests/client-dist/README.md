# Run:AI Unified Scale Tests

Validate your Run:AI installation at scale with comprehensive performance testing.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Quick Start](#quick-start)
4. [Configuration Reference](#configuration-reference)
5. [Running Tests](#running-tests)
6. [Understanding Results](#understanding-results)
7. [Advanced Usage](#advanced-usage)
8. [FAQ](#faq)

---

## Overview

The Run:AI Unified Scale Tests validate your Run:AI installation by simulating workload patterns at scale. Tests use simple sleep pods (not actual training/inference workloads) to measure Run:AI's control plane, scheduling, and orchestration performance—not GPU compute capabilities.

**What Gets Measured:**
- Workload creation and pod scheduling speed
- System scalability under load (hundreds of projects, thousands of workloads)
- Control plane synchronization performance

**Use Cases:**
- Before production: Validate installation before going live
- Capacity planning: Understand cluster performance at different scales
- Post-upgrade: Verify performance after upgrading Run:AI
- Troubleshooting: Diagnose performance issues

**Test Resources Created:**
- Run:AI projects (namespaces)
- Training workloads with sleep pods (configurable GPU/CPU requests)
- Node pools (real nodes or simulated with KWOK)

All test resources are automatically cleaned up after completion.

---

## Prerequisites

For complete installation and setup instructions, see the [Run:AI Documentation](https://run-ai-docs.nvidia.com/).

### Required

1. **Docker**
   - Docker Desktop (macOS/Windows) or Docker Engine (Linux)
   - Version: 20.10 or newer
   - Install: https://docs.docker.com/get-docker/

2. **kubectl Access**
   - kubectl configured with cluster-admin permissions
   - Verify: `kubectl cluster-info`
   - Your kubeconfig must point to the cluster you want to test

3. **Run:AI Installed**
   - Run:AI cluster and control plane must be installed and operational
   - Required operators: Run:AI cluster, training-operator
   - Verify: `kubectl get pods -n runai`
   - Installation guide: https://run-ai-docs.nvidia.com/self-hosted/getting-started/installation/install-using-helm

4. **Run:AI Control Plane Access**
   - **Cluster ID**: Get your cluster UUID using `runai cluster list --json | jq -r '.clusters[0].uuid'`
   - **Control Plane URL**: Your Run:AI control plane URL (e.g., https://your-company.run.ai)
   - **Credentials**: Username and password for control plane access
   - Get these from your Run:AI administrator or use `runai login` to authenticate

### Optional

- **Prometheus**: For metrics collection and visualization (tests work without it)
- **Grafana**: For visualizing test metrics
- **KWOK Operator**: For simulating nodes without real hardware (recommended for large-scale tests)

### Permissions Required

- Kubernetes cluster-admin permissions
- Full Run:AI permissions

---

## Quick Start

Get started in 5 minutes with a small-scale test.

### Step 1: Extract the Distribution Package

```bash
tar -xzf unified-tests-client-<version>.tar.gz
cd unified-tests-client-<version>
```

Replace `<version>` with the actual version from your downloaded file.

### Step 2: Load the Docker Image

```bash
docker load < docker-image.tar
```

This loads the pre-built test image into your local Docker. Verify it loaded successfully:

```bash
docker images | grep runai/unified-tests
```

You should see `runai/unified-tests` with the `latest` tag.

### Step 3: Configure Control Plane Access

Edit the example config file with your Run:AI Control Plane details:

```bash
# Edit small-scale-test.env
vi examples/small-scale-test.env

# Add your control plane information:
CLUSTER_ID=your-cluster-id
CONTROL_PLANE_URL=https://your-company.run.ai
CONTROL_PLANE_USERNAME=admin@your-company.com
CONTROL_PLANE_PASSWORD=your-password
```

### Step 4: Verify Prerequisites

```bash
# Check Docker
docker --version

# Check kubectl connectivity
kubectl cluster-info

# Check Run:AI is running
kubectl get pods -n runai
```

### Step 5: Run a Small-Scale Test

```bash
# Create results directory
mkdir -p results

# Run the test (uses small-scale-test.env example)
./run-test.sh examples/small-scale-test.env
```

The test will:
1. Create 4 projects
2. Create 100 training workloads
3. Monitor workload creation and scheduling
4. Generate performance metrics
5. Clean up all test resources
6. Save results to `results/test-results-YYYY-MM-DD-HH-MM-SS.json`

**Expected Duration**: 5-10 minutes

### Step 6: View Results

```bash
# View success rate and P95 latencies
jq '{
  success_rate: .performance_metrics.success_failure_analysis.success_rate,
  first_pod_creation_p95: .performance_metrics.workload_first_pod_creation.latency_ms.p95,
  pod_scheduling_p95: .performance_metrics.workload_pod_scheduling.latency_ms.p95,
  whole_lifecycle_p95: .performance_metrics.workload_whole_lifecycle.latency_ms.p95
}' results/test-results-*.json
```

Look for:
- `success_rate` should be high (determine acceptable threshold for your environment)
- P95 latencies indicate performance characteristics

---

## Configuration Reference

### Configuration File Format

Tests are configured using `.env` files with environment variables. Three example configurations are provided:

| Configuration | Projects | Workloads | Duration | Use Case |
|---------------|----------|-----------|----------|----------|
| `small-scale-test.env` | 4 | 100 | 5-10 min | Quick validation, smoke tests |
| `medium-scale-test.env` | 50 | 1,000 | 15-30 min | Standard production validation |
| `large-scale-test.env` | 250 | 10,000 | 1-2 hours | Extreme scale, capacity planning |

### Core Parameters

```bash
# Resource tier: small, medium, large, or custom
RESOURCES=medium

# Number of projects to create (1-500)
PROJECT_COUNT=50

# Number of workloads to create (1-10000)
WORKLOAD_COUNT=1000

# Workload type: Training or Distributed
WORKLOAD_TYPE=Training

# Workers per distributed workload
WORKLOAD_WORKER_COUNT=1

# Workload behavior
RUN_INFINITE_WORKLOADS=false
WORKLOAD_SLEEP_TIMES=180
```

### Resource Tiers

Resource tiers determine CPU and memory allocated to Run:AI services during testing:

| Tier | CPU per Service | Memory per Service | When to Use |
|------|----------------|-------------------|-------------|
| `small` | 0.5-1 CPU | 2-3Gi | Development clusters, limited resources |
| `medium` | 2 CPU | 7Gi | Standard production clusters |
| `large` | 4+ CPU | 24Gi+ | High-performance clusters |
| `custom` | User-defined | User-defined | Advanced users with specific requirements |

### Node Pool Configuration

```bash
# Number of node pools to create
# Set to 0 to use pre-existing nodepools
NODEPOOL_COUNT=2

# Number of nodes per pool
NODES_PER_POOL=10

# Use KWOK virtual nodes (true/false)
# KWOK simulates nodes without real hardware
KWOK_ENABLED=true
```

**Three Node Pool Options:**

1. **Use Pre-existing Nodepools** (recommended for existing clusters)
   - Set `NODEPOOL_COUNT=0`
   - Test uses all existing nodepools in your Run:AI cluster
   - Requires: At least 1 nodepool already configured in Run:AI Control Plane

2. **Create KWOK Virtual Nodepools** (recommended for scale testing)
   - Set `NODEPOOL_COUNT > 0` and `KWOK_ENABLED=true`
   - Test creates virtual nodes using KWOK (no real hardware needed)
   - Allows testing massive scale without real hardware

3. **Create Nodepools for Real Nodes**
   - Set `NODEPOOL_COUNT > 0` and `KWOK_ENABLED=false`
   - Test creates Run:AI nodepool configurations
   - Requires: Real nodes with matching labels already exist in Kubernetes

### Advanced Parameters

```bash
# GPUs per workload
GPU_PER_WORKLOAD=1

# Enable node chaos testing (simulates node failures)
NODE_CHAOS_ENABLED=false

# Enable Prometheus metrics collection
# Set to false if Prometheus is not available
PROMETHEUS_REMOTE_WRITE_ENABLED=false

# Cleanup after test completes
CLUSTER_CLEANUP=true

# Grouping label for organizing results
SCALE_GROUP=client-validation
```

### Creating Custom Configurations

Copy the template and customize:

```bash
cp examples/test-config.env.template my-custom-test.env
# Edit my-custom-test.env with your parameters
./run-test.sh my-custom-test.env
```

---

## Running Tests

### Using the Helper Script (Recommended)

The `run-test.sh` script provides the easiest way to run tests:

```bash
# Use default config (test-config.env in current directory)
./run-test.sh

# Use a specific config file
./run-test.sh examples/medium-scale-test.env

# Use custom results directory
./run-test.sh examples/small-scale-test.env my-test-results/
```

The script validates prerequisites and provides helpful error messages.

### Using Docker Directly

For more control, run Docker directly.

**First, ensure the Docker image is loaded** (if you haven't already):
```bash
docker load < docker-image.tar
```

Then run the container:

```bash
docker run --rm \
  -v ~/.kube/config:/app/kubeconfig/config:ro \
  -v $(pwd)/results:/app/results \
  -v $(pwd)/examples/medium-scale-test.env:/app/config/test-config.env:ro \
  runai/unified-tests:latest
```

**Volume Mounts Explained:**
- `-v ~/.kube/config:/app/kubeconfig/config:ro` - Mount your kubeconfig (read-only)
- `-v $(pwd)/results:/app/results` - Mount results directory for output
- `-v $(pwd)/test-config.env:/app/config/test-config.env:ro` - Mount config file (read-only)

### Monitoring Test Progress

Tests output progress to stdout in real-time:

```
==========================================
Run:AI Unified Scale Tests
Version: v1.0.0
==========================================

Test Configuration:
  Projects: 50
  Workloads: 1000
  Resources: medium
  KWOK Enabled: true
==========================================

Starting test execution...
[Progress messages...]
```

You can also watch test resources in your cluster:

```bash
# Watch projects being created
kubectl get projects -A --watch

# Watch workloads
kubectl get workloads -A --watch

# Watch pods
kubectl get pods -A --watch
```

---

## Understanding Results

### Result Files

Each test run creates a timestamped JSON file:

```
results/test-results-2026-02-04-14-30-45.json
```

The filename includes the date and time the test started.

### Result Structure

The results JSON contains several top-level sections:

```json
{
  "id": "scaleTestRun_20260208-123401",
  "metadata": { ... },
  "performance_metrics": {
    "success_failure_analysis": { ... },
    "workload_first_pod_creation": { ... },
    "workload_pod_scheduling": { ... },
    "workload_sync_to_cluster": { ... },
    "workload_sync_to_control_plane": { ... },
    "workload_whole_lifecycle": { ... },
    "prometheus_crash_rate": "0",
    "prometheus_target_sync_failed_total": { ... }
  },
  "environment": { ... }
}
```

### Success/Failure Analysis

The `success_failure_analysis` section provides an overview of test results:

```json
"success_failure_analysis": {
  "total_workloads": 10,
  "success_count": 10,
  "failure_count": 0,
  "success_rate": 100,
  "failure_breakdown": {},
  "top_failure_reasons": [],
  "stage_failures": {
    "CreatedOnCluster": { "success_count": 10, "failure_count": 0 },
    "FirstPodCreated":  { "success_count": 10, "failure_count": 0 },
    "FirstPodBound":    { "success_count": 10, "failure_count": 0 },
    "PhaseChangeSynced":{ "success_count": 10, "failure_count": 0 },
    "WholeLifecycle":   { "success_count": 10, "failure_count": 0 }
  }
}
```

**Stage failures** break down where workloads failed in their lifecycle:
- **CreatedOnCluster**: Workload resource was created on the Kubernetes cluster
- **FirstPodCreated**: First pod was created for the workload
- **FirstPodBound**: First pod was bound (scheduled) to a node
- **PhaseChangeSynced**: Workload phase change was synced to the control plane
- **WholeLifecycle**: Full lifecycle completed successfully

### Latency Metrics

Each latency metric has the following structure:

```json
"workload_first_pod_creation": {
  "pending_rate": 0,
  "latency_ms": {
    "mean": 1100,
    "max": 2000,
    "min": 1000,
    "p50": 1000,
    "p75": 1000,
    "p95": 2000,
    "std_dev": 300
  },
  "sample_size": 10
}
```

| Metric | Description |
|--------|-------------|
| **workload_first_pod_creation** | Time from workload submission to first pod created. Measures Run:AI's ability to process workload requests. |
| **workload_pod_scheduling** | Time for Kubernetes to schedule (bind) a pod to a node. Measures scheduler performance under load. |
| **workload_sync_to_cluster** | Time for Run:AI to sync workload state from control plane to cluster. |
| **workload_sync_to_control_plane** | Time for Run:AI to sync workload state from cluster back to control plane. |
| **workload_whole_lifecycle** | End-to-end time from workload submission to completion. Measures overall system performance. |

**Latency fields:**
- **mean**: Average latency across all workloads
- **min/max**: Fastest and slowest observed latency
- **p50, p75, p95**: Percentiles (50% / 75% / 95% of workloads completed faster than this value)
- **std_dev**: Standard deviation — high values indicate inconsistent performance
- **pending_rate**: Fraction of workloads that remained pending (did not complete this stage)
- **sample_size**: Number of workloads measured

**Focus on P95**: This percentile represents typical worst-case performance.

### Metadata

The `metadata` section contains test configuration and throughput info:

```bash
# Key metadata fields
jq '.metadata | {
  workload_creation_throughput_per_sec,
  workload_creation_duration_seconds,
  workload_creation_success_count,
  workload_creation_failed_count
}' results/test-results-*.json
```

- **workload_creation_throughput_per_sec**: Rate at which workloads were submitted to the API
- **workload_creation_duration_seconds**: Total time spent creating all workloads
- **workload_creation_success_count / failed_count**: API-level success/failure counts

### Prometheus Health Metrics

These metrics indicate whether monitoring infrastructure was stable during the test:

- **prometheus_crash_rate**: Rate of Prometheus container crashes (should be "0")
- **prometheus_target_sync_failed_total**: Per-component sync failures (all should be 0)
- **prometheus_remote_write_samples_failed_total**: Failed metric samples (should be "0")

### Interpreting Results

✅ **Good Results:**
- High `success_rate` (determine acceptable threshold for your environment)
- Consistent P95 metrics across runs
- Low `std_dev` relative to `mean`
- `pending_rate` is 0 for all metrics

⚠️ **Warning Signs:**
- High `std_dev` (inconsistent latencies)
- Non-zero `pending_rate`
- Stage failures concentrated at a specific stage

❌ **Failed Test:**
- Low success rate
- Very high P95 latencies
- `failure_breakdown` shows specific error patterns
- Check `top_failure_reasons` for root cause


---

## Advanced Usage

### Using KWOK Virtual Nodes

KWOK (Kubernetes WithOut Kubelet) simulates nodes without provisioning real infrastructure.

**Benefits:**
- No real hardware needed for simulated nodes
- Test scheduler behavior at massive scale
- Fast node provisioning

**Limitations:**
- Pods don't actually run (no workload execution)
- Cannot test actual GPU workloads
- Focus is on control plane and scheduler performance

**Configuration:**
```bash
# Enable KWOK
KWOK_ENABLED=true
NODES_PER_POOL=1000  # Can simulate many nodes

# Use real nodes
KWOK_ENABLED=false
NODES_PER_POOL=10    # Limited by actual cluster size
```

**When to Use KWOK:**
- ✅ Testing scheduler scalability
- ✅ Control plane performance testing
- ✅ Capacity planning (theoretical limits)
- ❌ GPU workload validation
- ❌ Actual workload execution testing

### Custom Resource Configurations

For fine-grained control over resource allocation, use `RESOURCES=custom`:

```bash
RESOURCES=custom

# Define resources for each service
# See examples/test-config.env.template for full list
SCHEDULING_SERVICES_CPU_REQUEST=2
SCHEDULING_SERVICES_MEMORY_REQUEST=4Gi
```

See `examples/test-config.env.template` for full list of services.

### Running Without Prometheus

If Prometheus is not available in your cluster:

```bash
PROMETHEUS_REMOTE_WRITE_ENABLED=false
```

**Impact:**
- Tests still run and generate results
- Metrics only available in JSON output file
- No real-time Grafana visualization

### Chaos Testing

Test system resilience with node chaos:

```bash
NODE_CHAOS_ENABLED=true
NODE_CHAOS_COUNT=1
NODE_CHAOS_DOWN_DURATION=5m
```

**What It Does:**
- Randomly terminates nodes during testing
- Simulates real-world failure scenarios
- Validates Run:AI's fault tolerance

**Use Cases:**
- Resilience testing
- Disaster recovery validation
- Production readiness assessment

**Warning**: Only use on non-production clusters.

### Running Multiple Tests in Parallel

**Not Recommended**: Tests create cluster-wide resources and may interfere with each other.

**Better Approach**: Run tests sequentially:

```bash
#!/bin/bash
for config in examples/*.env; do
    echo "Running test: $config"
    ./run-test.sh "$config"
    echo "Waiting 5 minutes before next test..."
    sleep 300
done
```

---

## FAQ

### How long do tests take?

| Scale | Duration | Workloads |
|-------|----------|-----------|
| Small | 5-10 minutes | 100 |
| Medium | 15-30 minutes | 1,000 |
| Large | 1-2 hours | 10,000 |

Duration depends on cluster performance and resources.

### Do tests affect production workloads?

Tests create their own projects and workloads with unique naming. However:
- Tests consume cluster resources (CPU, memory, GPU)
- Scheduler processes test workloads
- **Recommendation**: Run on non-production clusters or during maintenance windows

### Can I stop a test mid-execution?

Yes, but cleanup may not complete:

```bash
# Stop Docker container
docker ps  # Find container ID
docker stop <container-id>

# Manual cleanup may be required
kubectl get projects -A | grep runai-test-
kubectl delete project <test-project-names>
```

### Where are results stored?

Results are stored in the `results/` directory (or custom directory specified):

```
results/
├── test-results-2026-02-04-14-30-45.json
├── test-results-2026-02-04-16-15-22.json
└── test-results-2026-02-04-18-45-33.json
```

Each file is named with the test start time.

### Can I run tests on multiple clusters?

Yes, change your kubeconfig context before running:

```bash
# Switch cluster
kubectl config use-context cluster-1
./run-test.sh

# Switch to another cluster
kubectl config use-context cluster-2
./run-test.sh
```

### Can I run tests on a small cluster?

Yes, you can run tests with reduced scale to match your cluster size:

```bash
# Minimal test configuration
PROJECT_COUNT=2
WORKLOAD_COUNT=50
RESOURCES=small
NODEPOOL_COUNT=0  # Use pre-existing nodepools
```

Results will still be valid for your cluster's capabilities.

### Can I integrate this into CI/CD?

Yes, the Docker image can be used in CI/CD pipelines:

```yaml
# Example GitLab CI
test-runai:
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker run --rm
        -v $KUBECONFIG:/app/kubeconfig/config:ro
        -v $(pwd)/results:/app/results
        -v $(pwd)/test-config.env:/app/config/test-config.env:ro
        runai/unified-tests:latest
  artifacts:
    paths:
      - results/
```

### How do I update to a newer version?

```bash
# Extract new distribution package
tar -xzf unified-tests-client-<new-version>.tar.gz
cd unified-tests-client-<new-version>

# Load the new Docker image
docker load < docker-image.tar

# Run tests with new version
./run-test.sh examples/small-scale-test.env
```

