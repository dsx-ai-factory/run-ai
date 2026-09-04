# Schema Validation GitHub Action

## Overview

The `validate-cluster-config-schema.yml` workflow ensures that the Helm chart's `values.schema.json` file stays synchronized with the Go API types in `RunaiConfigSpec`.

## Purpose

When developers modify API types in `runai-cluster/cluster/sdk/apis/`, the corresponding JSON schema for Helm chart validation must be regenerated. This action automates the detection of when the schema is out of sync.

## Workflow Details

### Triggers
- **Pull Requests** to `master`, `main`, `v*.*`, or `release-*.*` branches
- **Path Filters**: Only runs when these paths change:
  - `runai-cluster/cluster/sdk/apis/**`
  - `runai-cluster/cluster/runai-chart/values.schema.json`  
  - `runai-cluster/cluster/sdk/hack/generate-values-schema.go`

### Environment
- **Runner**: `ubuntu-latest`
- **Go Version**: 1.24.6
- **Permissions**: `contents: read`, `pull-requests: write`

### Steps

1. **Setup**
   - Checkout repository
   - Setup Go environment

2. **Generate Schema**
   - Run schema generation script
   - Create temporary schema file
   - Log file sizes and paths

3. **Compare Schemas**
   - Compare generated vs committed schema
   - Detect missing files
   - Calculate diff statistics

4. **Report Results**
   - Post automated PR comment with results
   - Update existing comments instead of creating duplicates
   - Provide specific fix instructions

5. **Cleanup & Exit**
   - Remove temporary files
   - Exit with appropriate status code

### Error Scenarios

#### Missing Schema File
**Cause**: Schema file doesn't exist in the repository
**Action**: Posts comment with creation instructions
**Fix**: Run `make generate-cluster-config-values-schema`

#### Schema Out of Sync  
**Cause**: Generated schema differs from committed version
**Action**: Posts diff summary and fix instructions
**Fix**: Run `make generate-cluster-config-values-schema` and commit

### Success Scenario
**Condition**: Generated schema matches committed version
**Action**: Posts success confirmation
**Result**: CI check passes

## PR Comment Examples

### ❌ Schema Missing
```markdown
## 🔍 Schema Validation Failed

❌ **Missing Schema File**

The committed schema file `runai-cluster/cluster/runai-chart/values.schema.json` was not found.

### 🔧 How to Fix
Run the following commands:
```bash
cd runai-cluster/cluster/sdk
make generate-cluster-config-values-schema
git add ../runai-chart/values.schema.json
git commit -m "Add generated values.schema.json"
```

### ❌ Schema Out of Sync
```markdown  
## 🔍 Schema Validation Failed

❌ **Schema Out of Sync**

The committed `values.schema.json` doesn't match what would be generated from the current API types.

**Diff Summary:** 150 lines changed

### 🔧 How to Fix
```bash
cd runai-cluster/cluster/sdk
make generate-cluster-config-values-schema
git add ../runai-chart/values.schema.json
git commit -m "Update values.schema.json after API changes"
```

<details>
<summary>📋 View Diff (first 50 lines)</summary>
[diff content...]
</details>
```

### ✅ Success
```markdown
## 🔍 Schema Validation Passed

✅ **Schema Up to Date**

The committed `values.schema.json` matches the generated schema.
```

## Integration Points

### Local Development
- Complements the `make validate` target
- Provides same validation logic as CI
- Helps catch issues before pushing

### Other GitHub Actions
- Runs independently of other workflows
- Can be combined with other validation steps
- Uses concurrency groups to avoid conflicts

### Make System
- Uses same generation script as `make generate-cluster-config-values-schema`
- Consistent validation logic between local and CI
- Leverages existing Go toolchain

## Configuration

### Environment Variables
```yaml
GO_VERSION: "1.21"
SCHEMA_PATH: "runai-cluster/cluster/runai-chart/values.schema.json"
SDK_PATH: "runai-cluster/cluster/sdk"
```

### Permissions Required
- `contents: read` - Checkout repository
- `pull-requests: write` - Post comments on PRs

### Concurrency
```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

## Monitoring & Debugging

### Logs Structure
- Clear step names with emojis
- File size reporting
- Path logging for troubleshooting
- Diff line counts

### Common Issues

1. **Go Module Issues**
   - Check Go version compatibility
   - Verify module dependencies

2. **Path Issues**  
   - Verify relative paths from SDK directory
   - Check file permissions

3. **Git Issues**
   - Ensure proper checkout depth
   - Verify file tracking in git

### Testing Locally

To test the validation logic locally:

```bash
# Navigate to repository root
cd /path/to/runai

# Generate schema to temp file
cd runai-cluster/cluster/sdk
TEMP_SCHEMA=$(mktemp)
go run hack/generate-values-schema.go "$TEMP_SCHEMA"

# Compare with committed version
diff ../runai-chart/values.schema.json "$TEMP_SCHEMA"

# Cleanup
rm "$TEMP_SCHEMA"
```

## Maintenance

### Updating the Action
- Modify workflow file: `.github/workflows/validate-cluster-config-schema.yml`
- Test changes in feature branches
- Consider backward compatibility

### Adding New Paths
Update the `paths` filter to include new relevant file patterns:
```yaml
paths:
  - 'runai-cluster/cluster/sdk/apis/**'
  - 'runai-cluster/cluster/runai-chart/values.schema.json'  
  - 'runai-cluster/cluster/sdk/hack/generate-values-schema.go'
  # Add new paths here
```

### Disabling Temporarily
Comment out the `paths` section to disable path-based triggering:
```yaml
# paths:
#   - 'runai-cluster/cluster/sdk/apis/**'
```
