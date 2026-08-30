# Resource Sweep Scripts

Scripts for triggering scale tests with custom resource configurations to find optimal allocations for Run:AI services.

## Prerequisites

```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

## Quick Start

### Single Test Run

```bash
# With custom resources
python 00_trigger_formula_tests.py \
  --subset large-configs \
  --branch your-branch \
  --custom-resources \
  --scheduling-cpu-request 2 \
  --scheduling-memory-request 7Gi \
  --workload-cpu-request 4 \
  --workload-memory-request 24Gi \
  --sync-cpu-request 4 \
  --sync-memory-request 24Gi

# With predefined tiers (small/medium/large)
python 00_trigger_formula_tests.py --subset large-configs --branch your-branch
```

### Parameter Sweep

```bash
# Sweep one service group
python parameter_sweep_scheduling.py \
  --sweep-groups scheduling \
  --cpu-start 1.0 --cpu-end 2.0 --cpu-step 0.5 \
  --memory-start 2.0 --memory-end 4.0 --memory-step 1.0 \
  --sync-cpu-request 1.0 --sync-memory-request 2Gi \
  --workload-cpu-request 1.5 --workload-memory-request 3Gi \
  --branch your-branch

# Sweep multiple service groups together
python parameter_sweep_scheduling.py \
  --sweep-groups scheduling workload \
  --cpu-start 1.0 --cpu-end 2.0 --cpu-step 0.5 \
  --memory-start 2.0 --memory-end 4.0 --memory-step 1.0 \
  --branch your-branch
```

## Available Flags

### 00_trigger_formula_tests.py

**Basic:**
- `--subset`: Which tests to run (small-configs, medium-configs, large-configs, edge-cases, all)
- `--branch`: Git branch to run from
- `--dry-run`: Show what would run without executing
- `--delay`: Seconds between triggering tests (default: 5)
- `--run-group`: Custom run group name

**Custom Resources:**
- `--custom-resources`: Enable custom resource mode (required for custom values)
- `--scheduling-cpu-request`, `--scheduling-cpu-limit`
- `--scheduling-memory-request`, `--scheduling-memory-limit`
- `--workload-cpu-request`, `--workload-cpu-limit`
- `--workload-memory-request`, `--workload-memory-limit`
- `--sync-cpu-request`, `--sync-cpu-limit`
- `--sync-memory-request`, `--sync-memory-limit`
- `--training-cpu-request`, `--training-cpu-limit`
- `--training-memory-request`, `--training-memory-limit`
- `--gpu-operator-cpu-request`, `--gpu-operator-cpu-limit`
- `--gpu-operator-memory-request`, `--gpu-operator-memory-limit`

### parameter_sweep_scheduling.py

**Basic:**
- `--sweep-groups`: Service groups to sweep (scheduling, sync, workload). Default: scheduling
- `--subset`: Which test configs to use (default: small-configs)
- `--branch`: Git branch to run from
- `--dry-run`: Show what would run without executing

**Sweep Range:**
- `--cpu-start`, `--cpu-end`, `--cpu-step` (defaults: 0.5, 2.0, 0.1)
- `--memory-start`, `--memory-end`, `--memory-step` (defaults: 0.5, 2.0, 0.1)
- `--sleep-between`: Seconds between runs (default: 60)
- `--run-group-prefix`: Prefix for run group names (default: 'sweep')

**Static Resources (for non-swept groups):**
- Same as 00_trigger_formula_tests.py custom resource flags

## Test Configurations

- **small-configs**: 30 nodes, 480 workloads, 10 projects
- **medium-configs**: 100 nodes, 1600 workloads, 50 projects
- **large-configs**: 4 tests ranging from 250-500 nodes, 3000-12000 workloads
- **edge-cases**: Sparse/dense node-to-workload ratios, high project counts

## Monitoring

```bash
# List recent runs
gh run list --workflow=scale-test-manual-trigger.yml --repo run-ai/runai

# Watch specific run
gh run watch <run-id> --repo run-ai/runai
```

Results are available in your Grafana and Thanos endpoints.
