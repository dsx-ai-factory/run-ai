# OpenAPI String Pattern Scanner

A comprehensive toolset for analyzing and reporting on string schemas with regex pattern validation in OpenAPI specifications.

## Overview

This project provides tools to scan OpenAPI YAML files and generate detailed reports about all string schemas that validate string values against regex patterns. It helps identify:

- All string schemas with pattern validation
- **String schemas without any constraints (unconstrained strings)**
- Pattern categories and complexity
- Security concerns and validation issues
- Common pattern types and usage statistics
- **Validation coverage analysis and improvement recommendations**
- **Report comparison between different versions**

## Features

### 🔍 String Pattern Scanner (`src/string_pattern_scanner.py`)
- Scans OpenAPI YAML files for string schemas with regex patterns
- **Identifies unconstrained string schemas (no patterns, enums, or other validations)**
- Identifies patterns in schema definitions, parameters, and request bodies
- Validates regex patterns for correctness
- Generates comprehensive, formatted reports using Rich
- Exports results to JSON for further analysis
- **Provides validation coverage statistics**

### 📊 Pattern Analyzer (`src/pattern_analyzer.py`)
- Provides detailed analysis of the found patterns
- **Analyzes unconstrained strings and provides improvement recommendations**
- Categorizes patterns by purpose (URLs, emails, resource quantities, etc.)
- Analyzes pattern complexity and security implications
- Identifies common pattern templates and validation issues
- **Prioritizes validation recommendations by security impact**
- **Categorizes unconstrained strings by validation needs**

### 🔄 Report Comparator (`src/report_comparator.py`)
- **Compares two OpenAPI pattern analysis reports to identify changes**
- **Shows added, removed, and modified patterns between versions**
- **Tracks changes in unconstrained strings**
- **Generates detailed diff reports in markdown and JSON formats**
- **Provides summary statistics of changes**
- **Exports comparison data programmatically via JSON**
- **Useful for API version migration and change tracking**

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd openapi-scanner
```

2. Install dependencies using Poetry:
```bash
poetry install
```

3. Activate the virtual environment:
```bash
poetry shell
```

## Usage

### Basic Pattern Scanning

Scan an OpenAPI YAML file for string patterns:

```bash
python src/string_pattern_scanner.py runai-public.yaml
```

This will display a comprehensive report showing:
- Summary of found patterns
- Schema definitions with pattern validation
- Parameter patterns
- Request body patterns
- Pattern analysis and statistics

### Save Report to JSON

Generate a JSON report for further analysis:

```bash
python src/string_pattern_scanner.py runai-public.yaml -o report.json
```

### Detailed Pattern Analysis

Analyze the generated JSON report for detailed insights:

```bash
python src/pattern_analyzer.py report.json
```

This provides:
- Pattern categorization by purpose
- Complexity analysis
- Security analysis
- Validation issue detection
- Statistical summaries

### Report Comparison

Compare two reports to identify changes between versions:

```bash
# Using the command-line tool
python src/report_comparator.py old_report.json new_report.json

# Or using the poetry script
poetry run report-comparator old_report.json new_report.json

# Generate a markdown diff report
python src/report_comparator.py old_report.json new_report.json -o comparison.md

# Generate a JSON diff report
python src/report_comparator.py old_report.json new_report.json -j comparison.json

# Generate both markdown and JSON reports
python src/report_comparator.py old_report.json new_report.json -o comparison.md -j comparison.json

# Suppress console output and only generate the diff report
python src/report_comparator.py old_report.json new_report.json -o comparison.md -q
```

The comparison tool will show:
- **Summary changes**: Total pattern counts and differences
- **Added patterns**: New patterns in the newer report
- **Removed patterns**: Patterns that were in the old report but not in the new
- **Modified patterns**: Patterns that changed between versions
- **Unconstrained strings changes**: Changes in unconstrained string schemas

### Example Comparison Workflow

```bash
# 1. Generate a baseline report
python src/string_pattern_scanner.py v1.0-api.yaml -o v1.0-report.json

# 2. Generate a new report after API changes
python src/string_pattern_scanner.py v1.1-api.yaml -o v1.1-report.json

# 3. Compare the reports (markdown output)
python src/report_comparator.py v1.0-report.json v1.1-report.json -o api-changes.md

# 4. Compare the reports (JSON output for programmatic processing)
python src/report_comparator.py v1.0-report.json v1.1-report.json -j api-changes.json

# 5. Generate both formats
python src/report_comparator.py v1.0-report.json v1.1-report.json -o api-changes.md -j api-changes.json
```

### Automated Pattern Change Detection (CI/CD)

#### Updating the Baseline Report (`update_branch.sh`)

The `update_branch.sh` script provides a convenient way to update the baseline report:

```bash
# Update the branch baseline report
./update_branch.sh
```

**What it does:**
1. Runs the string pattern scanner on `../../specs/runai-public.yaml`
2. Generates/updates `branch_report.json` with the current patterns
3. This file serves as the baseline for pattern change detection

**When to use:**
- After making intentional changes to OpenAPI patterns
- When creating a new baseline for your branch
- Before committing to update the baseline report in version control

**Usage:**
```bash
# Update the baseline report after making OpenAPI changes
./update_branch.sh

# Commit the updated baseline
git add branch_report.json
git commit -m "Update baseline pattern report"
```

#### Pattern Change Detection (`check_patterns.sh`)

The `check_patterns.sh` script provides an automated workflow for detecting pattern changes in CI/CD pipelines:

```bash
# Run the automated pattern change checker
./check_patterns.sh
```

**What it does:**
1. Checks that `branch_report.json` exists (baseline report from your branch)
2. Generates a current report from `../../specs/runai-public.yaml`
3. Creates a comparison report between the baseline and current reports
4. Exits with code 1 if added or modified patterns are detected, code 0 otherwise

**Usage in CI/CD:**
```bash
# Before running the script, generate a baseline report from your branch
./update_branch.sh
# OR manually:
# poetry run string-pattern-scanner ../../specs/runai-public.yaml -o branch_report.json

# Commit branch_report.json to your repository
git add branch_report.json
git commit -m "Add baseline pattern report"

# In your CI pipeline, run the check
./check_patterns.sh

# Exit code 0 = no breaking changes
# Exit code 1 = pattern changes detected (review required)
```

**Output example:**
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
  • newFieldPattern: ^[A-Z0-9]+$
  • anotherPattern: ^[a-z-]+$

Modified Patterns:
  • existingPattern
```

## Report Structure

### String Pattern Scanner Output

The scanner generates reports with the following sections:

1. **Summary**: Total count of patterns found
2. **Schema Definitions**: String schemas with pattern validation
3. **Parameters**: API parameters with pattern validation
4. **Request Bodies**: Request body schemas with pattern validation
5. **Pattern Analysis**: Statistics and categorization

### Pattern Categories

The analyzer categorizes patterns into:

- **Resource Quantities**: Memory, CPU, and storage specifications
- **URLs and URIs**: Web addresses and resource identifiers
- **Identifiers and Names**: Names, keys, and identifiers
- **Email Addresses**: Email validation patterns
- **File Paths**: File system path validation
- **Filter Expressions**: Query and filter patterns
- **Version Numbers**: Version string validation
- **UUIDs**: Unique identifier patterns
- **Other**: Uncategorized patterns

### Security Analysis

The tool identifies potential security concerns:

- Overly permissive patterns
- Weak validation patterns
- Potential injection vulnerabilities
- Missing anchors in patterns

## Example Output

### Pattern Scanner Report

```
🔍 Scanning OpenAPI file: runai-public.yaml
Scanning schema definitions...
Scanning parameters...
Scanning request bodies...

╭─────────────────────────────────────────────╮
│ 🔍 OpenAPI String Pattern Validation Report │
╰─────────────────────────────────────────────╯

📊 Summary: Found 37 string schemas with regex pattern validation
   • Schema definitions: 37
   • Parameters: 0
   • Request bodies: 0
   • Unconstrained strings: 492

📊 String validation coverage:
   • Constrained strings: 37 (7.0%)
   • Unconstrained strings: 492 (93.0%)
   ⚠️  Low validation coverage - consider adding more constraints

🎯 Priority Recommendations:
🔴 HIGH Priority:
   • key: NodeTaint -> key
   • Reason: Security-sensitive field needs validation

================================================================================
📋 SCHEMA DEFINITIONS WITH PATTERN VALIDATION
================================================================================
┏━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━┳━━━━━━━━━━┓
┃ Schema Name            ┃ Pattern                          ┃ Description             ┃ Format  ┃ Validat… ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━╇━━━━━━━━━━┩
│ reason                 │ ^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z │ reason contains a       │ -       │ ✅ Valid │
│                        │ 0-9_])?$                         │ program...              │         │          │
│ cpuSwapMemorySize      │ ^([+-]?[0-9.]+)(*[-+]?[0-9]*)$   │ Size of the swap cpu    │ -       │ ✅ Valid │
│                        │                                  │ memo...                 │         │          │
└────────────────────────┴──────────────────────────────────┴─────────────────────────┴─────────┴──────────┘
```

### Pattern Analyzer Report

```
╭──────────────────────────────╮
│ 🔍 Detailed Pattern Analysis │
╰──────────────────────────────╯

================================================================================
📊 PATTERN CATEGORIES
================================================================================

🔹 Resource Quantities (7 patterns):
   • cpuSwapMemorySize: ^([+-]?[0-9.]+)(*[-+]?[0-9]*)$
     Size of the swap cpu memory. Only allowed when swapEnabled is true...

🔹 URLs and URIs (4 patterns):
   • metadataXmlUrl: /^(https?://)?(+\.{2,})(:[0-9]{1,5})?(/[^\s]*)?$|^$/

🔹 Identifiers and Names (5 patterns):
   • reason: ^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$
     reason contains a programmatic identifier indicating the reason for the condition's last transition...

================================================================================
🧮 PATTERN COMPLEXITY ANALYSIS
================================================================================
                                             Pattern Complexity

┏━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ Schema                ┃ Pattern                              ┃ Complexity  ┃ Description                 ┃
┃                       ┃                                      ┃ Score       ┃                             ┃
┡━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩
│ metadataXmlUrl        │ /^(https?://)?(+\.{2,})(:[0-9]{1,5}) │ 53 (High)   │                             │
│                       │ ?(/[^\s]*)?$|^$/                     │             │                             │
└───────────────────────┴──────────────────────────────────────┴─────────────┴─────────────────────────────┘
```

## JSON Report Format

### Pattern Scanner JSON Report

The JSON report contains:

```json
{
  "summary": {
    "total_patterns": 37,
    "schema_patterns": 37,
    "parameter_patterns": 0,
    "request_patterns": 0
  },
  "schema_patterns": [
    {
      "schema_name": "reason",
      "path": "ClusterReportedStatus -> conditions -> items -> reason",
      "pattern": "^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$",
      "description": "reason contains a programmatic identifier...",
      "format": "",
      "min_length": 1,
      "max_length": 1024,
      "enum": null,
      "example": "",
      "nullable": false,
      "required": false
    }
  ],
  "parameter_patterns": [],
  "request_patterns": []
}
```

### Report Comparison JSON Report

The comparison JSON report contains comprehensive diff information:

```json
{
  "summary": {
    "old": {
      "total_patterns": 35,
      "schema_patterns": 35,
      "parameter_patterns": 0,
      "request_patterns": 0
    },
    "new": {
      "total_patterns": 37,
      "schema_patterns": 37,
      "parameter_patterns": 0,
      "request_patterns": 0
    },
    "changes": {
      "total_patterns": 2,
      "schema_patterns": 2,
      "parameter_patterns": 0,
      "request_patterns": 0
    }
  },
  "added_patterns": [
    {
      "schema_name": "newPattern",
      "pattern": "^[A-Z]+$",
      "description": "A new pattern added in v1.1"
    }
  ],
  "removed_patterns": [],
  "modified_patterns": [
    {
      "schema_name": "existingPattern",
      "old": {
        "pattern": "^[a-z]+$",
        "min_length": 1
      },
      "new": {
        "pattern": "^[a-z]{2,}$",
        "min_length": 2
      },
      "changes": [
        {
          "field": "pattern",
          "old": "^[a-z]+$",
          "new": "^[a-z]{2,}$"
        },
        {
          "field": "min_length",
          "old": 1,
          "new": 2
        }
      ]
    }
  ],
  "unchanged_patterns": [],
  "unconstrained_strings_changes": {
    "added": [],
    "removed": [],
    "unchanged": []
  },
  "statistics": {
    "total_changes": 2,
    "added_count": 1,
    "removed_count": 0,
    "modified_count": 1,
    "unchanged_count": 35
  }
}
```

## Dependencies

- **Python**: ^3.10
- **PyYAML**: ^6.0.2
- **Click**: ^8.2.1
- **Rich**: ^14.0.0
- **Pydantic**: ^2.11.7

## Development

### Running Tests

```bash
poetry run pytest
```

### Code Formatting

```bash
poetry run black src/
poetry run isort src/
```

### Type Checking

```bash
poetry run mypy src/
```

## Use Cases

This tool is particularly useful for:

1. **API Documentation Review**: Ensuring consistent pattern validation across APIs
2. **Security Audits**: Identifying weak or overly permissive validation patterns
3. **Code Generation**: Understanding validation requirements for client code generation
4. **Compliance**: Verifying that APIs follow consistent validation standards
5. **Migration**: Analyzing patterns when migrating between API versions
6. **CI/CD Integration**: 
   - Using `check_patterns.sh` for automated pattern change detection in pipelines
   - Failing builds when breaking pattern changes are detected
   - Using JSON comparison reports for automated validation
7. **Change Management**: Tracking and reviewing API validation changes between releases
8. **Programmatic Analysis**: Processing comparison results with custom tools using JSON format
9. **Pull Request Reviews**: Automatically detecting and flagging pattern changes in PRs

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 