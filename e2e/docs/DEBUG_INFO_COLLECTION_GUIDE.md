# Debug Info Collection - Test Failure Diagnostics Guide

## Overview

The Debug Info Collection feature automatically collects diagnostic information when tests fail, providing comprehensive debugging data directly in Allure reports. This eliminates the need for manual investigation and provides immediate access to relevant logs and system state information.

## How It Works

When a test with debug tags fails, the system automatically:
1. Detects the failure in the afterEach hook
2. Parses debug tags from the test title
3. Collects relevant diagnostic information based on the specified tags
4. Attaches the collected information to the Allure report

## Quick Setup Guide

### Step 1: Verify Debug Info Collection is Enabled

**For Local Development:**
Ensure this environment variable is set in your `.env` file:

```bash
COLLECT_DEBUG_INFO_ENABLED='true'
```

**For Pipeline/CI:**
The debug info collection feature is enabled by default in pipeline environments.

**Note**: When disabled, the system will skip all debug info collection, even if debug tags are present in test titles.

### Step 2: Update Import Statement

Replace your existing Playwright import:

```typescript
// Before
import { test, expect } from '@playwright/test';

// After
import { test, expect, DEBUG_TAGS } from '../../helpers/DebugInfoCollector';
```

### Step 3: Add Debug Tags to Test Titles

Add `debugTags` to your test titles. Here are common usage examples:

**Pattern 1: Single Tag at Test Level**
```typescript
test(`My test, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS}`, async function () {
    // Your test implementation
});
```

**Pattern 2: Multiple Tags at Test Level**
```typescript
test(`My test, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS} ${DEBUG_TAGS.PODS}`, async function () {
    // Your test implementation
});
```

**Pattern 3: Describe Level (applies to all tests in the suite)**
```typescript
test.describe(`My test suite, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS}`, function () {
    test('Test case 1', async function () {
        // Will collect SYSTEM_LOGS automatically
    });
    
    test('Test case 2', async function () {
        // Will collect SYSTEM_LOGS automatically
    });
});
```

**Pattern 4: Combined Usage (Describe + Test Level)**
```typescript
test.describe(`Workload tests, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS}`, function () {
    test('Create workload', async function () {
        // Will collect: SYSTEM_LOGS (from describe)
    });
    
    test(`Delete workload, debugTags: ${DEBUG_TAGS.PODS}`, async function () {
        // Will collect: SYSTEM_LOGS (from describe) + PODS (from test)
    });
    
    test(`Suspend workload, debugTags: ${DEBUG_TAGS.PODS} ${DEBUG_TAGS.TRAINING_WORKLOAD_CRDS}`, async function () {
        // Will collect: SYSTEM_LOGS (from describe) + PODS + TRAINING_WORKLOAD_CRDS (from test)
    });
});
```

That's it! When your test fails, debug information will be automatically collected and attached to the Allure report.

## How Debug Tags Work

### Describe Level vs Test Case Level

**Describe Level Tags:**
- Applied to all test cases within the describe block
- Useful for setting common debug requirements for a group of related tests
- Syntax: `test.describe('Suite name, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS}', function () {`

**Test Case Level Tags:**
- Applied only to the specific test case
- Useful for targeted debugging of individual tests
- Syntax: `test('Test name, debugTags: ${DEBUG_TAGS.PODS_LOGS}', async function () {`

### How They Work Together

When both describe-level and test-level tags are present, they are **combined** (not overridden). The system will collect debug information for all specified tags.

**Processing Steps:**
1. **Collection**: All debug tags from describe and test levels are collected
2. **Expansion**: Composite tags (like `SYSTEM_LOGS`) are expanded into their individual components
3. **Deduplication**: Duplicate tags are automatically removed
4. **Execution**: Debug information is collected for all unique tags

**Example:**
```typescript
test.describe(`My test suite, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS}`, function () {
    test('Test case 1', async function () {
        // Will collect: SYSTEM_LOGS (from describe)
        // Expanded to: WORKLOAD_CONTROLLER_LOGS, AGENT_LOGS, CLUSTER_SYNC_LOGS, etc.
    });
    
    test(`Specific test, debugTags: ${DEBUG_TAGS.PODS_LOGS}`, async function () {
        // Will collect: SYSTEM_LOGS (from describe) + PODS_LOGS (from test)
        // Expanded to: All system component logs + Pod logs
        // Result: Both system logs AND pod information
    });
});
```

### Usage Examples

**Basic Test Case Level:**
```typescript
test(`Test description, debugTags: ${DEBUG_TAGS.CORE_SYSTEM_LOGS}`, async function () {
    // Your test implementation
});
```

**Multiple Tags at Test Level:**
```typescript
test(`Complex test, debugTags: ${DEBUG_TAGS.CORE_SYSTEM_LOGS} ${DEBUG_TAGS.PODS_LOGS}`, async function () {
    // Your test implementation
});
```

**Describe Level (applies to all tests):**
```typescript
test.describe(`My test suite, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS}`, function () {
    test('Test case 1', async function () {
        // Will collect SYSTEM_LOGS automatically
    });
    
    test('Test case 2', async function () {
        // Will collect SYSTEM_LOGS automatically
    });
});
```

**Combined Usage (recommended for complex scenarios):**
```typescript
test.describe(`Workload management tests, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS}`, function () {
    test('Create workload', async function () {
        // Will collect: SYSTEM_LOGS
    });
    
    test(`Delete workload, debugTags: ${DEBUG_TAGS.PODS}`, async function () {
        // Will collect: SYSTEM_LOGS + PODS
        // Perfect for debugging workload deletion issues
    });
    
    test(`Suspend workload, debugTags: ${DEBUG_TAGS.TRAINING_WORKLOAD_CRDS}`, async function () {
        // Will collect: SYSTEM_LOGS + TRAINING_WORKLOAD_CRDS
        // Perfect for debugging suspension issues
    });
});
```

### Available Debug Tags

The Debug Info Collection framework provides a comprehensive set of debug tags organized into different categories:

#### System Component Logs
**Individual Components:**
- `DEBUG_TAGS.WORKLOAD_CONTROLLER_LOGS` - Workload Controller logs
- `DEBUG_TAGS.AGENT_LOGS` - Agent logs  
- `DEBUG_TAGS.CLUSTER_SYNC_LOGS` - Cluster Sync logs
- `DEBUG_TAGS.WORKLOAD_OVERSEER_LOGS` - Workload Overseer logs
- `DEBUG_TAGS.RUNAI_JOB_CONTROLLER_LOGS` - RunAI Job Controller logs
- `DEBUG_TAGS.MPI_OPERATOR_LOGS` - MPI Operator logs
- `DEBUG_TAGS.TRAINING_OPERATOR_LOGS` - Training Operator logs
- `DEBUG_TAGS.WORKLOADS_SERVICE_LOGS` - Workloads Service logs
- `DEBUG_TAGS.INFERENCE_WORKLOAD_CONTROLLER_LOGS` - Inference Workload Controller logs
- `DEBUG_TAGS.POD_GROUP_ASSIGNER_LOGS` - Pod Group Assigner logs

**Composite System Tags:**
- `DEBUG_TAGS.CORE_SYSTEM_LOGS` - Core system components (Workload Controller, Agent, Cluster Sync)
- `DEBUG_TAGS.EXTENDED_SYSTEM_LOGS` - Extended system components (includes Workload Overseer, RunAI Job Controller, Pod Group Assigner)
- `DEBUG_TAGS.SYSTEM_LOGS` - All system component logs (recommended for general debugging)

#### Pod Information
- `DEBUG_TAGS.PODS_YAML` - Pod YAML configurations
- `DEBUG_TAGS.PODS_LOGS` - Pod logs
- `DEBUG_TAGS.PODS` - Both pod YAML and logs (composite tag)

#### Workload CRDs (Custom Resource Definitions)
**Individual Workload Types:**
- `DEBUG_TAGS.TRAINING_WORKLOADS` - Training Workload CRDs
- `DEBUG_TAGS.INTERACTIVE_WORKLOADS` - Interactive Workload CRDs  
- `DEBUG_TAGS.DISTRIBUTED_WORKLOADS` - Distributed Workload CRDs
- `DEBUG_TAGS.INFERENCE_WORKLOADS` - Inference Workload CRDs

**Job Types:**
- `DEBUG_TAGS.RUNAI_JOBS` - RunAI Jobs
- `DEBUG_TAGS.XGBOOST_JOBS` - XGBoost Jobs
- `DEBUG_TAGS.TF_JOBS` - TensorFlow Jobs
- `DEBUG_TAGS.MPI_JOBS` - MPI Jobs
- `DEBUG_TAGS.PYTORCH_JOBS` - PyTorch Jobs
- `DEBUG_TAGS.JAX_JOBS` - JAX Jobs
- `DEBUG_TAGS.KNATIVE_SERVICES` - Knative Services

**Composite Workload Tags:**
- `DEBUG_TAGS.TRAINING_WORKLOAD_CRDS` - Training workloads + RunAI Jobs + Pods
- `DEBUG_TAGS.INTERACTIVE_WORKLOAD_CRDS` - Interactive workloads + RunAI Jobs + Pods
- `DEBUG_TAGS.DISTRIBUTED_WORKLOAD_CRDS` - All distributed job types + Pods
- `DEBUG_TAGS.INFERENCE_WORKLOAD_CRDS` - Inference workloads + Knative Services
- `DEBUG_TAGS.ALL_KUBEFLOW_JOBS` - All Kubeflow job types (XGBoost, TF, MPI, PyTorch, JAX)
- `DEBUG_TAGS.ALL_CRDS` - All CRD types

#### Comprehensive Debug Tags
- `DEBUG_TAGS.FULL_DEBUG` - Everything (all system logs + all CRDs + pods)

#### Quick Reference by Use Case

| Use Case | Recommended Tags |
|----------|------------------|
| General system debugging | `DEBUG_TAGS.SYSTEM_LOGS` |
| Training workload issues | `DEBUG_TAGS.TRAINING_WORKLOAD_CRDS` |
| Interactive workload issues | `DEBUG_TAGS.INTERACTIVE_WORKLOAD_CRDS` |
| Distributed training issues | `DEBUG_TAGS.DISTRIBUTED_WORKLOAD_CRDS` |
| Pod-related issues | `DEBUG_TAGS.PODS` |
| Comprehensive debugging | `DEBUG_TAGS.FULL_DEBUG` |

**Note**: Composite tags (like `SYSTEM_LOGS`, `PODS`, etc.) are automatically expanded into their individual components and deduplicated. For example, `DEBUG_TAGS.SYSTEM_LOGS` expands to all system component logs, and if you specify both `SYSTEM_LOGS` and `WORKLOAD_CONTROLLER_LOGS`, the system will only collect the Workload Controller logs once.

## Viewing Debug Information

1. **Run your tests** with debug tags
2. **Navigate to failed tests** in the Allure report
3. **Check the "After Hooks" section** for debug information

The debug information will appear as a "Debug Info Collection" step within the "After Hooks" section of the test execution details, organized by the type of information collected.

### Example: Debug Info Collection in Allure Report

Expand **Debug Info Collection** under **After Hooks** to open the collected component logs, pod YAML, and workload resource details. Each attachment shows its file size.

**Real Example from Codebase:**
Combine the following debug tags to collect this information:
- `DEBUG_TAGS.SYSTEM_LOGS` (describe level) - Collects all system component logs
- `DEBUG_TAGS.PODS` (test level) - Collects pod YAML configurations and logs  
- `DEBUG_TAGS.TRAINING_WORKLOAD_CRDS` (test level) - Collects Training Workload CRD information

This demonstrates how multiple debug tags can be combined to collect comprehensive diagnostic information for complex test scenarios. The collection is triggered automatically when tests fail, providing immediate access to diagnostic information without manual intervention.

## Best Practices

### 1. Choose Appropriate Tags
- Use specific tags for targeted debugging
- Use composite tags for comprehensive debugging
- Avoid using `FULL_DEBUG` unless necessary, as it collects a large amount of data

### 2. Tag Placement
- Place debug tags at the test level for specific test debugging
- Place debug tags at the describe level for group-level debugging

### 3. Performance Considerations
- Debug collection only occurs on test failures
- Consider storage implications when using comprehensive tags
- Use the global switch to disable debug collection when not needed

### 4. Troubleshooting
- Ensure the import path is correct
- Verify debug tags are properly formatted in test titles
- Check that the test is actually failing (debug collection only triggers on failures)

## Appendix: Technical Details

### Performance Impact

- **Zero impact on passing tests**: Debug collection only triggers on test failures
- **Storage consideration**: Large debug collections may increase report size
- **Global control**: Can be completely disabled via `COLLECT_DEBUG_INFO_ENABLED` environment variable

### Integration Points

- **Playwright Test Lifecycle**: Integrates with afterEach hooks
- **Allure Reporting**: Debug information appears as a "Debug Info Collection" step within the "After Hooks" section
- **Existing Test Code**: Fully backward compatible, no changes required to existing tests 

## Support

For questions or issues with the Debug Info Collection feature, please contact `loong` team
