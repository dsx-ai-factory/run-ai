import pytest
import json
import tempfile
import os
import yaml
import copy
from unittest.mock import patch
from src.string_pattern_scanner import StringPatternScanner
from src.pattern_analyzer import analyze_patterns
from src.report_comparator import ReportComparator


class TestIntegrationWorkflow:
    """Integration tests for the complete OpenAPI scanner workflow."""
    
    @pytest.mark.integration
    def test_full_workflow(self, temp_yaml_file, temp_yaml_data):
        """Test the complete workflow from OpenAPI YAML to pattern analysis."""
        
        # Step 1: Test OpenAPI scanner functionality
        with open(temp_yaml_file, 'r') as f:
            data = yaml.safe_load(f)
        
        paths = data.get('paths', {})
        assert len(paths) > 0
        
        # Step 2: Test string pattern scanner
        scanner = StringPatternScanner(temp_yaml_data)
        
        # Scan for patterns
        string_patterns = scanner.scan_schemas()
        parameter_patterns = scanner.scan_parameters()
        request_patterns = scanner.scan_request_bodies()
        unconstrained_strings = scanner.scan_unconstrained_strings()
        
        # Verify we found expected patterns
        assert len(string_patterns) >= 4  # id, name, email, product_id
        assert len(parameter_patterns) >= 2  # search, id parameters
        assert len(unconstrained_strings) >= 1  # description
        
        # Step 3: Test pattern analysis
        report_data = {
            'schema_patterns': string_patterns,
            'parameter_patterns': parameter_patterns,
            'request_patterns': request_patterns,
            'unconstrained_strings': unconstrained_strings,
            'summary': {
                'total_patterns': len(string_patterns) + len(parameter_patterns) + len(request_patterns),
                'schema_patterns': len(string_patterns),
                'parameter_patterns': len(parameter_patterns),
                'request_patterns': len(request_patterns)
            }
        }
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            json.dump(report_data, f)
            report_file = f.name
        
        try:
            with patch('src.pattern_analyzer.console') as mock_console:
                analyze_patterns(report_file)
                assert mock_console.print.called
        finally:
            os.unlink(report_file)
    
    @pytest.mark.integration
    def test_report_comparison_workflow(self, sample_pattern_report):
        """Test the report comparison workflow."""
        
        # Create two slightly different reports
        old_report = copy.deepcopy(sample_pattern_report)
        new_report = copy.deepcopy(sample_pattern_report)
        
        # Add a new pattern to the new report
        new_pattern = {
            'schema_name': 'new_field',
            'path': 'User -> new_field',
            'pattern': '^[0-9]+$',
            'description': 'New numeric field',
            'format': None,
            'min_length': None,
            'max_length': None,
            'enum': None,
            'example': '123',
            'nullable': False,
            'required': False
        }
        new_report['schema_patterns'].append(new_pattern)
        
        # Update the summary to reflect the new pattern
        new_report['summary']['total_patterns'] = old_report['summary']['total_patterns'] + 1
        new_report['summary']['schema_patterns'] = old_report['summary']['schema_patterns'] + 1
        
        # Modify an existing pattern
        new_report['schema_patterns'][0]['max_length'] = 50  # Changed from None
        
        # Create temporary files
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as old_file:
            json.dump(old_report, old_file)
            old_path = old_file.name
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as new_file:
            json.dump(new_report, new_file)
            new_path = new_file.name
        
        try:
            # Test report comparison
            comparator = ReportComparator(old_path, new_path)
            comparator.load_reports()
            comparison = comparator.compare_reports()
            
            # Debug: Print the comparison results
            print(f"Old report has {len(old_report['schema_patterns'])} patterns")
            print(f"New report has {len(new_report['schema_patterns'])} patterns")
            print(f"Added patterns: {len(comparison['added_patterns'])}")
            print(f"Modified patterns: {len(comparison['modified_patterns'])}")
            print(f"Unchanged patterns: {len(comparison['unchanged_patterns'])}")
            
            # Verify that we have the expected number of patterns
            assert len(old_report['schema_patterns']) == 4
            assert len(new_report['schema_patterns']) == 5
            
            # Verify comparison results
            # The comparison should detect 1 added pattern and 1 modified pattern
            assert len(comparison['added_patterns']) == 1
            assert len(comparison['modified_patterns']) == 1
            assert comparison['added_patterns'][0]['schema_name'] == 'new_field'
            assert comparison['modified_patterns'][0]['schema_name'] == 'id'
            
            # Test display functions
            with patch('src.report_comparator.console') as mock_console:
                from src.report_comparator import display_comparison
                display_comparison(comparison)
                assert mock_console.print.called
            
            # Test diff report generation
            with tempfile.NamedTemporaryFile(mode='w', suffix='.md', delete=False) as diff_file:
                diff_path = diff_file.name
            
            try:
                from src.report_comparator import generate_diff_report
                report_content = generate_diff_report(comparison, diff_path)
                assert isinstance(report_content, str)
                assert 'OpenAPI Report Comparison' in report_content
                
                # Verify file was written
                with open(diff_path, 'r') as f:
                    content = f.read()
                    assert 'OpenAPI Report Comparison' in content
            finally:
                if os.path.exists(diff_path):
                    os.unlink(diff_path)
        finally:
            os.unlink(old_path)
            os.unlink(new_path)
    
    @pytest.mark.integration
    def test_string_pattern_scanner_with_complex_schema(self):
        """Test string pattern scanner with complex nested schemas."""
        
        complex_yaml_data = {
            'components': {
                'schemas': {
                    'ComplexObject': {
                        'type': 'object',
                        'properties': {
                            'simple_string': {
                                'type': 'string',
                                'pattern': '^[a-zA-Z]+$'
                            },
                            'nested_object': {
                                'type': 'object',
                                'properties': {
                                    'nested_string': {
                                        'type': 'string',
                                        'pattern': '^[0-9]+$'
                                    }
                                }
                            },
                            'array_of_objects': {
                                'type': 'array',
                                'items': {
                                    'type': 'object',
                                    'properties': {
                                        'array_string': {
                                            'type': 'string',
                                            'pattern': '^[a-zA-Z0-9]+$'
                                        }
                                    }
                                }
                            },
                            'all_of_composition': {
                                'allOf': [
                                    {
                                        'type': 'object',
                                        'properties': {
                                            'base_string': {
                                                'type': 'string',
                                                'pattern': '^base_'
                                            }
                                        }
                                    },
                                    {
                                        'type': 'object',
                                        'properties': {
                                            'extended_string': {
                                                'type': 'string',
                                                'pattern': '^extended_'
                                            }
                                        }
                                    }
                                ]
                            }
                        }
                    }
                }
            }
        }
        
        scanner = StringPatternScanner(complex_yaml_data)
        patterns = scanner.scan_schemas()
        
        # Should find 5 patterns: simple_string, nested_string, array_string, base_string, extended_string
        assert len(patterns) == 5
        
        pattern_names = [p['schema_name'] for p in patterns]
        expected_names = ['simple_string', 'nested_string', 'array_string', 'base_string', 'extended_string']
        
        for expected_name in expected_names:
            assert expected_name in pattern_names
    
    @pytest.mark.integration
    def test_pattern_validation_workflow(self):
        """Test the pattern validation workflow."""
        
        patterns_with_issues = [
            {
                'schema_name': 'valid_pattern',
                'pattern': '^[a-zA-Z]+$',
                'description': 'Valid pattern'
            },
            {
                'schema_name': 'invalid_pattern',
                'pattern': '^[a-zA-Z+$',  # Invalid - missing closing bracket
                'description': 'Invalid pattern'
            },
            {
                'schema_name': 'complex_valid',
                'pattern': '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
                'description': 'Complex but valid email pattern'
            }
        ]
        
        from src.string_pattern_scanner import validate_regex_pattern
        
        # Test validation
        for pattern in patterns_with_issues:
            is_valid, error = validate_regex_pattern(pattern['pattern'])
            
            if pattern['schema_name'] == 'invalid_pattern':
                assert is_valid is False
                assert len(error) > 0
            else:
                assert is_valid is True
                assert error == ""
    
    @pytest.mark.integration
    def test_error_handling_workflow(self):
        """Test error handling in the workflow."""
        
        # Test with invalid YAML
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            f.write("invalid: yaml: content: [")
            invalid_yaml = f.name
        
        try:
            with pytest.raises(yaml.YAMLError):
                with open(invalid_yaml, 'r') as f:
                    yaml.safe_load(f)
        finally:
            os.unlink(invalid_yaml)
        
        # Test with nonexistent file
        with pytest.raises(FileNotFoundError):
            with open('nonexistent.yaml', 'r') as f:
                pass
        
        # Test with invalid JSON report
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            f.write("invalid json content")
            invalid_json = f.name
        
        try:
            with pytest.raises(json.JSONDecodeError):
                with open(invalid_json, 'r') as f:
                    json.load(f)
        finally:
            os.unlink(invalid_json) 