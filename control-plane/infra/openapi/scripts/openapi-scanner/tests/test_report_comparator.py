import pytest
import json
import tempfile
import os
from unittest.mock import patch, mock_open
from src.report_comparator import (
    ReportComparator, display_comparison, display_summary_changes,
    display_pattern_changes, display_unconstrained_strings_changes,
    generate_diff_report
)


class TestReportComparator:
    """Test cases for the ReportComparator class."""
    
    def setup_method(self):
        """Set up test data."""
        self.old_report_data = {
            'summary': {
                'total_patterns': 5,
                'schema_patterns': 3,
                'parameter_patterns': 1,
                'request_patterns': 1
            },
            'schema_patterns': [
                {
                    'schema_name': 'email',
                    'pattern': '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
                    'description': 'Email address',
                    'format': 'email',
                    'min_length': None,
                    'max_length': None,
                    'enum': None,
                    'example': 'user@example.com',
                    'nullable': False,
                    'required': True
                },
                {
                    'schema_name': 'user_id',
                    'pattern': '^[a-zA-Z][a-zA-Z0-9_-]*$',
                    'description': 'User identifier',
                    'format': None,
                    'min_length': 3,
                    'max_length': 20,
                    'enum': None,
                    'example': 'user123',
                    'nullable': False,
                    'required': True
                }
            ],
            'unconstrained_strings': [
                {
                    'schema_name': 'description',
                    'path': 'Product -> description',
                    'description': 'Product description'
                }
            ]
        }
        
        self.new_report_data = {
            'summary': {
                'total_patterns': 6,
                'schema_patterns': 4,
                'parameter_patterns': 1,
                'request_patterns': 1
            },
            'schema_patterns': [
                {
                    'schema_name': 'email',
                    'pattern': '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
                    'description': 'Email address',
                    'format': 'email',
                    'min_length': None,
                    'max_length': None,
                    'enum': None,
                    'example': 'user@example.com',
                    'nullable': False,
                    'required': True
                },
                {
                    'schema_name': 'user_id',
                    'pattern': '^[a-zA-Z][a-zA-Z0-9_-]*$',
                    'description': 'User identifier',
                    'format': None,
                    'min_length': 3,
                    'max_length': 25,  # Changed from 20
                    'enum': None,
                    'example': 'user123',
                    'nullable': False,
                    'required': True
                },
                {
                    'schema_name': 'password',
                    'pattern': '^.{8,}$',
                    'description': 'Password validation',
                    'format': None,
                    'min_length': 8,
                    'max_length': None,
                    'enum': None,
                    'example': 'password123',
                    'nullable': False,
                    'required': True
                }
            ],
            'unconstrained_strings': [
                {
                    'schema_name': 'description',
                    'path': 'Product -> description',
                    'description': 'Product description'
                },
                {
                    'schema_name': 'notes',
                    'path': 'Order -> notes',
                    'description': 'Order notes'
                }
            ]
        }
    
    def test_load_reports(self):
        """Test loading reports from files."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as old_file:
            json.dump(self.old_report_data, old_file)
            old_path = old_file.name
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as new_file:
            json.dump(self.new_report_data, new_file)
            new_path = new_file.name
        
        try:
            comparator = ReportComparator(old_path, new_path)
            comparator.load_reports()
            
            assert comparator.old_data == self.old_report_data
            assert comparator.new_data == self.new_report_data
        finally:
            os.unlink(old_path)
            os.unlink(new_path)
    
    def test_load_reports_file_not_found(self):
        """Test loading reports with nonexistent file."""
        comparator = ReportComparator('nonexistent_old.json', 'nonexistent_new.json')
        with pytest.raises(FileNotFoundError):
            comparator.load_reports()
    
    def test_load_reports_invalid_json(self):
        """Test loading reports with invalid JSON."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as old_file:
            old_file.write("invalid json content")
            old_path = old_file.name
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as new_file:
            json.dump(self.new_report_data, new_file)
            new_path = new_file.name
        
        try:
            comparator = ReportComparator(old_path, new_path)
            with pytest.raises(json.JSONDecodeError):
                comparator.load_reports()
        finally:
            os.unlink(old_path)
            os.unlink(new_path)
    
    def test_compare_reports(self):
        """Test comparing reports."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as old_file:
            json.dump(self.old_report_data, old_file)
            old_path = old_file.name
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as new_file:
            json.dump(self.new_report_data, new_file)
            new_path = new_file.name
        
        try:
            comparator = ReportComparator(old_path, new_path)
            comparator.load_reports()
            comparison = comparator.compare_reports()
            
            # Check summary changes
            assert comparison['summary']['changes']['total_patterns'] == 1
            assert comparison['summary']['changes']['schema_patterns'] == 1
            
            # Check added patterns
            assert len(comparison['added_patterns']) == 1
            assert comparison['added_patterns'][0]['schema_name'] == 'password'
            
            # Check modified patterns
            assert len(comparison['modified_patterns']) == 1
            assert comparison['modified_patterns'][0]['schema_name'] == 'user_id'
            
            # Check unchanged patterns
            assert len(comparison['unchanged_patterns']) == 1
            assert comparison['unchanged_patterns'][0]['schema_name'] == 'email'
            
            # Check unconstrained strings changes
            assert len(comparison['unconstrained_strings_changes']['added']) == 1
            assert len(comparison['unconstrained_strings_changes']['removed']) == 0
            assert len(comparison['unconstrained_strings_changes']['unchanged']) == 1
        finally:
            os.unlink(old_path)
            os.unlink(new_path)
    
    def test_patterns_are_equal(self):
        """Test pattern equality comparison."""
        pattern1 = {
            'pattern': '^[a-zA-Z]+$',
            'description': 'Test pattern',
            'format': None,
            'min_length': 1,
            'max_length': 10,
            'enum': None,
            'example': 'test',
            'nullable': False,
            'required': True
        }
        
        pattern2 = {
            'pattern': '^[a-zA-Z]+$',
            'description': 'Test pattern',
            'format': None,
            'min_length': 1,
            'max_length': 10,
            'enum': None,
            'example': 'test',
            'nullable': False,
            'required': True
        }
        
        pattern3 = {
            'pattern': '^[a-zA-Z]+$',
            'description': 'Different description',
            'format': None,
            'min_length': 1,
            'max_length': 10,
            'enum': None,
            'example': 'test',
            'nullable': False,
            'required': True
        }
        
        comparator = ReportComparator('old.json', 'new.json')
        
        assert comparator._patterns_are_equal(pattern1, pattern2) is True
        assert comparator._patterns_are_equal(pattern1, pattern3) is False
    
    def test_get_pattern_changes(self):
        """Test getting detailed pattern changes."""
        old_pattern = {
            'pattern': '^[a-zA-Z]+$',
            'description': 'Old description',
            'format': None,
            'min_length': 1,
            'max_length': 10,
            'enum': None,
            'example': 'old',
            'nullable': False,
            'required': True
        }
        
        new_pattern = {
            'pattern': '^[a-zA-Z]+$',
            'description': 'New description',
            'format': None,
            'min_length': 1,
            'max_length': 15,
            'enum': None,
            'example': 'new',
            'nullable': False,
            'required': True
        }
        
        comparator = ReportComparator('old.json', 'new.json')
        changes = comparator._get_pattern_changes(old_pattern, new_pattern)
        
        # Should detect 3 changes: description, max_length, and example
        assert len(changes) == 3
        change_fields = [change['field'] for change in changes]
        assert 'description' in change_fields
        assert 'max_length' in change_fields
        assert 'example' in change_fields


class TestDisplayFunctions:
    """Test cases for display functions."""
    
    def test_display_comparison(self):
        """Test display_comparison function."""
        comparison = {
            'summary': {
                'old': {'total_patterns': 5},
                'new': {'total_patterns': 6},
                'changes': {'total_patterns': 1}
            },
            'added_patterns': [],
            'removed_patterns': [],
            'modified_patterns': [],
            'unchanged_patterns': [],
            'unconstrained_strings_changes': {
                'added': [],
                'removed': [],
                'unchanged': []
            }
        }
        
        with patch('src.report_comparator.console') as mock_console:
            display_comparison(comparison)
            assert mock_console.print.called
    
    def test_display_summary_changes(self):
        """Test display_summary_changes function."""
        summary_comparison = {
            'old': {'total_patterns': 5},
            'new': {'total_patterns': 6},
            'changes': {'total_patterns': 1}
        }
        
        with patch('src.report_comparator.console') as mock_console:
            display_summary_changes(summary_comparison)
            assert mock_console.print.called
    
    def test_display_pattern_changes(self):
        """Test display_pattern_changes function."""
        comparison = {
            'added_patterns': [],
            'removed_patterns': [],
            'modified_patterns': [],
            'unchanged_patterns': []
        }
        
        with patch('src.report_comparator.console') as mock_console:
            display_pattern_changes(comparison)
            assert mock_console.print.called
    
    def test_display_unconstrained_strings_changes(self):
        """Test display_unconstrained_strings_changes function."""
        changes = {
            'added': [],
            'removed': [],
            'unchanged': []
        }
        
        with patch('src.report_comparator.console') as mock_console:
            display_unconstrained_strings_changes(changes)
            assert mock_console.print.called
    
    def test_generate_diff_report(self):
        """Test generate_diff_report function."""
        comparison = {
            'summary': {
                'old': {'total_patterns': 5},
                'new': {'total_patterns': 6},
                'changes': {'total_patterns': 1}
            },
            'added_patterns': [],
            'removed_patterns': [],
            'modified_patterns': [],
            'unchanged_patterns': [],
            'unconstrained_strings_changes': {
                'added': [],
                'removed': [],
                'unchanged': []
            }
        }
        
        report = generate_diff_report(comparison)
        assert isinstance(report, str)
        assert 'OpenAPI Report Comparison' in report
    
    def test_generate_diff_report_with_output_file(self):
        """Test generate_diff_report function with output file."""
        comparison = {
            'summary': {
                'old': {'total_patterns': 5},
                'new': {'total_patterns': 6},
                'changes': {'total_patterns': 1}
            },
            'added_patterns': [],
            'removed_patterns': [],
            'modified_patterns': [],
            'unchanged_patterns': [],
            'unconstrained_strings_changes': {
                'added': [],
                'removed': [],
                'unchanged': []
            }
        }
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.md', delete=False) as f:
            output_file = f.name
        
        try:
            report = generate_diff_report(comparison, output_file)
            assert isinstance(report, str)
            
            # Check that file was written
            with open(output_file, 'r') as f:
                content = f.read()
                assert 'OpenAPI Report Comparison' in content
        finally:
            os.unlink(output_file) 