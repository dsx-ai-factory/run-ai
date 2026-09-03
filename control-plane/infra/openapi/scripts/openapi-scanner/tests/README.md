# OpenAPI Scanner Test Suite

This directory contains comprehensive unit and integration tests for the OpenAPI Scanner project.

## Test Structure

### Unit Tests

- **`test_openapi_scanner.py`**: Tests for the basic OpenAPI YAML parsing and path extraction functionality
- **`test_string_pattern_scanner.py`**: Tests for the string pattern scanning and validation functionality
- **`test_pattern_analyzer.py`**: Tests for pattern analysis and categorization functionality
- **`test_report_comparator.py`**: Tests for report comparison and diff generation functionality

### Integration Tests

- **`test_integration.py`**: End-to-end workflow tests that verify the complete pipeline from OpenAPI YAML to pattern analysis and comparison

### Test Configuration

- **`conftest.py`**: Shared test fixtures and configuration
- **`pytest.ini`**: Pytest configuration and test discovery settings

## Running Tests

### Prerequisites

Make sure you have the project dependencies installed:

```bash
poetry install
```

### Run All Tests

```bash
# Run all tests
poetry run pytest

# Run with verbose output
poetry run pytest -v

# Run with coverage report
poetry run pytest --cov=src --cov-report=html
```

### Run Specific Test Categories

```bash
# Run only unit tests
poetry run pytest -m "not integration"

# Run only integration tests
poetry run pytest -m integration

# Run tests for a specific module
poetry run pytest tests/test_openapi_scanner.py

# Run a specific test function
poetry run pytest tests/test_string_pattern_scanner.py::TestStringPatternScanner::test_scan_schemas
```

### Test Options

```bash
# Run tests in parallel (requires pytest-xdist)
poetry run pytest -n auto

# Stop on first failure
poetry run pytest -x

# Show local variables in tracebacks
poetry run pytest -l

# Run tests with detailed output
poetry run pytest -s
```

## Test Coverage

The test suite covers:

### Core Functionality
- OpenAPI YAML parsing and validation
- String pattern extraction from schemas, parameters, and request bodies
- Regex pattern validation
- Pattern categorization and analysis
- Report generation and comparison

### Edge Cases
- Invalid YAML/JSON files
- Missing or malformed schemas
- Complex nested schema structures
- Error handling and recovery

### Integration Scenarios
- Complete workflow from YAML to analysis
- Report comparison between different versions
- Complex schema compositions (allOf, anyOf, oneOf)
- Array and additional properties handling

## Test Fixtures

The test suite includes several reusable fixtures:

- **`sample_openapi_yaml`**: Sample OpenAPI 3.0 specification for testing
- **`sample_pattern_report`**: Sample pattern analysis report
- **`temp_yaml_file`**: Temporary YAML file with sample data
- **`temp_json_file`**: Temporary JSON file with sample report data
- **`temp_yaml_data`**: Parsed YAML data for direct testing

## Adding New Tests

When adding new tests:

1. Follow the existing naming convention: `test_*.py` for test files
2. Use descriptive test function names: `test_function_name_scenario`
3. Add appropriate docstrings explaining what each test verifies
4. Use the provided fixtures when possible
5. Mock external dependencies (console output, file operations)
6. Test both success and failure scenarios
7. Add integration tests for new workflows

## Test Data

Test data is designed to be realistic but minimal, covering:
- Common OpenAPI patterns (email, UUID, version numbers)
- Various schema compositions
- Different parameter types and locations
- Edge cases and error conditions

## Continuous Integration

The test suite is designed to run in CI/CD environments:
- No external dependencies beyond the project requirements
- Fast execution (under 30 seconds for full suite)
- Deterministic results
- Proper cleanup of temporary files 