# Cluster Config Schema Validation

This document describes the automated validation system for `values.schema.json` in the RunAI Helm chart.

## Overview

The RunAI cluster uses a JSON Schema to validate Helm chart values based on the Go struct `RunaiConfigSpec`. To ensure this schema stays synchronized with API changes, we have implemented both local validation via Make targets and automated CI validation via GitHub Actions.

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   API Changes   │    │   Make Targets   │    │ GitHub Actions  │
│                 │    │                  │    │                 │
│ apis/run/v1/    │───▶│ make validate    │    │ PR Validation   │
│ *_types.go      │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │ Local Validation │    │ CI Validation   │
                       │                  │    │                 │
                       │ - Generate       │    │ - Generate temp │
                       │ - Compare        │    │ - Compare files │
                       │ - Git check      │    │ - PR comments   │
                       └──────────────────┘    └─────────────────┘
```

## Components

### 1. Schema Generation Script
- **Location**: `hack/generate-values-schema.go`
- **Purpose**: Generates JSON Schema from Go structs
- **Input**: `RunaiConfigSpec` and related types
- **Output**: `values.schema.json` file

### 2. Make Target
- **Target**: `generate-cluster-config-values-schema`
- **Purpose**: Regenerate schema locally
- **Usage**: `make generate-cluster-config-values-schema`

### 3. Validation Make Target
- **Target**: `validate`
- **Purpose**: Ensure all generated files are up-to-date
- **Dependencies**: `generate`, `manifests`, `copy-crds`, `generate-cluster-config-values-schema`

### 4. GitHub Action
- **File**: `.github/workflows/validate-cluster-config-schema.yml`
- **Trigger**: PRs affecting API types or schema files
- **Purpose**: Automated validation in CI/CD

## Workflow

### Local Development

1. **Developer modifies API types**:
   ```bash
   # Edit files in apis/run/v1/
   vim apis/run/v1/runaiconfig_types.go
   ```

2. **Run validation**:
   ```bash
   make validate
   ```

3. **If validation fails**:
   ```bash
   make generate-cluster-config-values-schema
   git add ../runai-chart/values.schema.json
   git commit -m "Update values.schema.json after API changes"
   ```

### CI/CD Validation

1. **PR Creation**: Developer creates PR with API changes
2. **GitHub Action Triggered**: On paths matching:
   - `runai-cluster/cluster/sdk/apis/**`
   - `runai-cluster/cluster/runai-chart/values.schema.json`
   - `runai-cluster/cluster/sdk/hack/generate-values-schema.go`

3. **Validation Process**:
   - Generate schema to temporary file
   - Compare with committed schema
   - Report results via PR comments
   - Fail CI if schemas don't match

4. **Developer Fix**:
   - Run `make generate-cluster-config-values-schema`
   - Commit updated schema
   - Push changes

## GitHub Action Details

### Triggers
- **Pull Requests** to `master` or version branches
- **Path Filters** to only run when relevant files change
- **Draft PR Exclusion** to avoid unnecessary runs

### Steps
1. **Setup**: Checkout code, setup Go
2. **Generate**: Create schema in temporary file
3. **Compare**: Diff against committed schema
4. **Report**: Comment on PR with results
5. **Fail**: Exit with error if validation fails

### Error Scenarios

#### Missing Schema File
```
❌ Missing Schema File

The committed schema file was not found.

How to Fix:
cd runai-cluster/cluster/sdk
make generate-cluster-config-values-schema
git add ../runai-chart/values.schema.json
git commit -m "Add generated values.schema.json"
```

#### Schema Out of Sync
```
❌ Schema Out of Sync

The committed values.schema.json doesn't match what would be 
generated from the current API types.

How to Fix:
cd runai-cluster/cluster/sdk
make generate-cluster-config-values-schema
git add ../runai-chart/values.schema.json
git commit -m "Update values.schema.json after API changes"
```

## Benefits

### For Developers
- **Early Detection**: Catch schema inconsistencies before merge
- **Clear Instructions**: Automated comments explain exactly how to fix issues
- **No Manual Checks**: Validation happens automatically

### For CI/CD
- **Prevents Regressions**: Ensures schema always matches API
- **Consistent State**: Avoids deployments with stale schemas
- **Documentation**: Schema serves as contract for Helm values

### For Operations
- **Helm Validation**: Chart values validated against current API
- **IDE Support**: Autocomplete and validation in IDEs
- **Documentation**: Schema documents available configuration options

## File Locations

```
runai-cluster/cluster/
├── sdk/
│   ├── Makefile                              # Contains validation targets
│   ├── apis/run/v1/runaiconfig_types.go     # Source of truth for schema
│   ├── hack/generate-values-schema.go       # Schema generation script
│   └── docs/schema-validation.md            # This documentation
└── runai-chart/
    └── values.schema.json                    # Generated schema file
```

## Troubleshooting

### Common Issues

1. **"Schema files differ"**
   - **Cause**: API types changed but schema not regenerated
   - **Fix**: `make generate-cluster-config-values-schema`

2. **"Schema file not found"**
   - **Cause**: Schema never generated or accidentally deleted
   - **Fix**: `make generate-cluster-config-values-schema`

3. **"Go module errors"**
   - **Cause**: Missing dependencies or Go version mismatch
   - **Fix**: `go mod tidy && go mod download`

### Manual Schema Generation

```bash
# Navigate to SDK directory
cd runai-cluster/cluster/sdk

# Generate schema to specific location
go run hack/generate-values-schema.go ../runai-chart/values.schema.json

# Or use make target
make generate-cluster-config-values-schema
```

### Bypass Validation (Emergency Only)

```bash
# Skip validation (NOT RECOMMENDED)
git commit --no-verify -m "Emergency commit - bypass validation"

# Or disable GitHub Action temporarily
# Edit .github/workflows/validate-cluster-config-schema.yml
# Comment out the trigger paths
```

## Integration with Existing Workflows

The schema validation integrates seamlessly with existing development workflows:

- **Make System**: Extends existing `validate` target
- **GitHub Actions**: Follows same patterns as other CI workflows  
- **Git Hooks**: Can be added to pre-commit hooks if desired
- **IDE Integration**: Schema enables autocomplete in VS Code/IntelliJ
