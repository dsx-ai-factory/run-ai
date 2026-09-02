import pytest
import json
import tempfile
import os
from unittest.mock import patch, mock_open
from src.pattern_analyzer import (
    analyze_patterns, categorize_patterns, analyze_pattern_complexity,
    identify_common_patterns, analyze_security_patterns, check_validation_issues,
    analyze_unconstrained_strings, generate_pattern_summary
)


class TestPatternAnalyzer:
    """Test cases for the pattern analyzer functions."""
    
    def setup_method(self):
        """Set up test data."""
        self.sample_patterns = [
            {
                'schema_name': 'email_field',
                'pattern': '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
                'description': 'Email address validation'
            },
            {
                'schema_name': 'memory_limit',
                'pattern': '^[+-]?\\d+(\\.\\d+)?[eEinumkKMGTP]*$',
                'description': 'Memory limit with units'
            },
            {
                'schema_name': 'user_id',
                'pattern': '^[a-zA-Z][a-zA-Z0-9_-]*$',
                'description': 'User identifier'
            },
            {
                'schema_name': 'file_path',
                'pattern': '^[a-zA-Z0-9\\x2F._-]+$',
                'description': 'File path validation'
            },
            {
                'schema_name': 'version',
                'pattern': '^v\\d+\\.\\d+\\.\\d+$',
                'description': 'Version number'
            }
        ]
        
        self.sample_unconstrained_strings = [
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
    
    def test_categorize_patterns(self):
        """Test pattern categorization."""
        with patch('src.pattern_analyzer.console') as mock_console:
            categorize_patterns(self.sample_patterns)
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_analyze_pattern_complexity(self):
        """Test pattern complexity analysis."""
        with patch('src.pattern_analyzer.console') as mock_console:
            analyze_pattern_complexity(self.sample_patterns)
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_identify_common_patterns(self):
        """Test common pattern identification."""
        with patch('src.pattern_analyzer.console') as mock_console:
            identify_common_patterns(self.sample_patterns)
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_analyze_security_patterns(self):
        """Test security pattern analysis."""
        with patch('src.pattern_analyzer.console') as mock_console:
            analyze_security_patterns(self.sample_patterns)
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_check_validation_issues(self):
        """Test validation issues checking."""
        with patch('src.pattern_analyzer.console') as mock_console:
            check_validation_issues(self.sample_patterns)
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_analyze_unconstrained_strings(self):
        """Test unconstrained strings analysis."""
        with patch('src.pattern_analyzer.console') as mock_console:
            analyze_unconstrained_strings(self.sample_unconstrained_strings)
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_generate_pattern_summary(self):
        """Test pattern summary generation."""
        with patch('src.pattern_analyzer.console') as mock_console:
            generate_pattern_summary(self.sample_patterns, self.sample_unconstrained_strings)
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_generate_pattern_summary_no_unconstrained(self):
        """Test pattern summary generation without unconstrained strings."""
        with patch('src.pattern_analyzer.console') as mock_console:
            generate_pattern_summary(self.sample_patterns)
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_analyze_patterns_with_report_file(self):
        """Test analyze_patterns function with a report file."""
        report_data = {
            'schema_patterns': self.sample_patterns,
            'unconstrained_strings': self.sample_unconstrained_strings
        }
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            json.dump(report_data, f)
            temp_file = f.name
        
        try:
            with patch('src.pattern_analyzer.console') as mock_console:
                analyze_patterns(temp_file)
                # Verify that console.print was called
                assert mock_console.print.called
        finally:
            os.unlink(temp_file)
    
    def test_analyze_patterns_with_invalid_file(self):
        """Test analyze_patterns function with invalid file."""
        with pytest.raises(FileNotFoundError):
            analyze_patterns('nonexistent.json')
    
    def test_analyze_patterns_with_invalid_json(self):
        """Test analyze_patterns function with invalid JSON."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            f.write("invalid json content")
            temp_file = f.name
        
        try:
            with pytest.raises(json.JSONDecodeError):
                analyze_patterns(temp_file)
        finally:
            os.unlink(temp_file)
    
    def test_pattern_categorization_logic(self):
        """Test the logic of pattern categorization."""
        # Test email pattern categorization
        email_pattern = {
            'schema_name': 'email',
            'pattern': '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
            'description': 'Email address'
        }
        
        with patch('src.pattern_analyzer.console') as mock_console:
            categorize_patterns([email_pattern])
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_complexity_scoring(self):
        """Test pattern complexity scoring."""
        simple_pattern = {
            'schema_name': 'simple',
            'pattern': '^[a-z]+$',
            'description': 'Simple pattern'
        }
        
        complex_pattern = {
            'schema_name': 'complex',
            'pattern': '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
            'description': 'Complex email pattern'
        }
        
        with patch('src.pattern_analyzer.console') as mock_console:
            analyze_pattern_complexity([simple_pattern, complex_pattern])
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_security_pattern_detection(self):
        """Test security pattern detection."""
        security_patterns = [
            {
                'schema_name': 'password',
                'pattern': '^.{8,}$',
                'description': 'Password validation'
            },
            {
                'schema_name': 'token',
                'pattern': '^[a-zA-Z0-9]{32}$',
                'description': 'API token'
            }
        ]
        
        with patch('src.pattern_analyzer.console') as mock_console:
            analyze_security_patterns(security_patterns)
            # Verify that console.print was called
            assert mock_console.print.called
    
    def test_validation_issues_detection(self):
        """Test validation issues detection."""
        problematic_patterns = [
            {
                'schema_name': 'problematic',
                'pattern': '^[a-zA-Z+$',  # Invalid pattern
                'description': 'Problematic pattern'
            }
        ]
        
        with patch('src.pattern_analyzer.console') as mock_console:
            check_validation_issues(problematic_patterns)
            # Verify that console.print was called
            assert mock_console.print.called 