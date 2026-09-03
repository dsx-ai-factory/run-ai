# Pattern Change Checker Script - Usage Guide

## Overview

The `check_patterns.sh` script automates the detection of OpenAPI pattern changes between a baseline (branch) report and the current state of your OpenAPI specification.

## Quick Start

```bash
# 1. Generate a baseline report (do this once per branch)
poetry run string-pattern-scanner ../../specs/runai-public.yaml -o branch_report.json

# 2. Run the checker
./check_patterns.sh
```

## What the Script Does

### Step 1: Validation
- ✅ Checks that `branch_report.json` exists
- 🚫 Exits with error if missing, with instructions on how to generate it

### Step 2: Current State Analysis
- 📊 Generates `current_report.json` from `../../specs/runai-public.yaml`
- 🔍 Scans all string patterns in your OpenAPI specification

### Step 3: Comparison
- 📈 Creates `comparison.json` comparing branch and current reports
- 🔄 Uses the `report-comparator` tool to identify differences

### Step 4: Change Detection
- 🎯 Analyzes the comparison for breaking changes
- ⚠️  Flags added or modified patterns as potential breaking changes
- ✅ Reports success if no breaking changes detected

## Exit Codes

- **0**: No pattern changes detected (safe to proceed)
- **1**: Pattern changes detected OR an error occurred (review required)

## Files Generated

- `current_report.json`: Report of current OpenAPI spec patterns
- `comparison.json`: Detailed diff between branch and current reports

## Requirements

### Optional: jq (JSON processor)
Install for detailed output:
```bash
brew install jq  # macOS
```

Without jq, the script will still work but with less detailed output.

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Check Pattern Changes

on: [pull_request]

jobs:
  check-patterns:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.10'
      
      - name: Install Poetry
        run: pip install poetry
      
      - name: Install Dependencies
        working-directory: control-plane/infra/openapi/scripts/openapi-scanner
        run: poetry install
      
      - name: Check Pattern Changes
        working-directory: control-plane/infra/openapi/scripts/openapi-scanner
        run: ./check_patterns.sh
```

### GitLab CI Example
```yaml
check-patterns:
  stage: test
  image: python:3.10
  script:
    - cd control-plane/infra/openapi/scripts/openapi-scanner
    - pip install poetry
    - poetry install
    - ./check_patterns.sh
  only:
    - merge_requests
```

## Typical Workflow

### For Developers

1. **Start working on a branch**
   ```bash
   git checkout -b feature/my-changes
   cd control-plane/infra/openapi/scripts/openapi-scanner
   ```

2. **Generate baseline report**
   ```bash
   poetry run string-pattern-scanner ../../specs/runai-public.yaml -o branch_report.json
   git add branch_report.json
   git commit -m "Add baseline pattern report"
   ```

3. **Make changes to OpenAPI spec**
   ```bash
   # Edit ../../specs/runai-public.yaml
   # Add/modify/remove patterns
   ```

4. **Check for pattern changes**
   ```bash
   ./check_patterns.sh
   ```

5. **Review changes if detected**
   - Check `comparison.json` for details
   - Review added/modified patterns
   - Update documentation if needed

### For CI/CD Pipelines

The script is designed to fail fast if breaking changes are detected:

```bash
#!/bin/bash
set -e  # Exit on any error

# Run pattern check
cd control-plane/infra/openapi/scripts/openapi-scanner
./check_patterns.sh

# If we get here, no breaking changes were detected
echo "Pattern check passed!"
```

## Output Examples

### Success (No Changes)
```
========================================
OpenAPI Pattern Change Checker
========================================

Step 1: Checking for branch_report.json...
✓ branch_report.json found

Step 2: Generating current report from OpenAPI spec...
✓ current_report.json generated

Step 3: Creating comparison report...
✓ comparison.json generated

Step 4: Checking for breaking changes...
Comparison Results:
  Added patterns: 0
  Modified patterns: 0
  Removed patterns: 0

✓ No breaking pattern changes detected
```

### Warning (Changes Detected)
```
========================================
OpenAPI Pattern Change Checker
========================================

Step 1: Checking for branch_report.json...
✓ branch_report.json found

Step 2: Generating current report from OpenAPI spec...
✓ current_report.json generated

Step 3: Creating comparison report...
✓ comparison.json generated

Step 4: Checking for breaking changes...
Comparison Results:
  Added patterns: 2
  Modified patterns: 1
  Removed patterns: 0

⚠️  WARNING: Pattern changes detected!
Added or modified patterns may indicate breaking changes.
Please review the comparison report: /path/to/comparison.json

Added Patterns:
  • emailPattern: ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$
  • phonePattern: ^\+?[1-9]\d{1,14}$

Modified Patterns:
  • usernamePattern
```

## Troubleshooting

### Error: branch_report.json not found
**Solution**: Generate the baseline report first
```bash
poetry run string-pattern-scanner ../../specs/runai-public.yaml -o branch_report.json
```

### Error: OpenAPI spec not found at ../../specs/runai-public.yaml
**Solution**: Check that you're running the script from the correct directory
```bash
cd control-plane/infra/openapi/scripts/openapi-scanner
./check_patterns.sh
```

### Script not executable
**Solution**: Make the script executable
```bash
chmod +x check_patterns.sh
```

## Advanced Usage

### Custom OpenAPI Spec Location
Edit the `OPENAPI_SPEC` variable in `check_patterns.sh`:
```bash
OPENAPI_SPEC="/path/to/your/spec.yaml"
```

### Review Detailed Comparison
```bash
# View JSON comparison
cat comparison.json | jq '.'

# View only added patterns
cat comparison.json | jq '.added_patterns'

# View only modified patterns
cat comparison.json | jq '.modified_patterns'
```

### Generate Markdown Report
```bash
poetry run report-comparator branch_report.json current_report.json -o changes.md
```

## Best Practices

1. **Always commit branch_report.json**: This ensures everyone on the team has the same baseline
2. **Run before committing**: Catch pattern changes early in development
3. **Review all changes**: Even if patterns are added/modified intentionally, review their impact
4. **Document breaking changes**: Update API documentation when patterns change
5. **Use in PR reviews**: Make pattern checking part of your PR review process

