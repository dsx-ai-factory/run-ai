import pytest
import yaml
import tempfile
import os
from unittest.mock import patch, mock_open
from src.openapi_scanner import get_schema_type, main


class TestGetSchemaType:
    """Test cases for the get_schema_type function."""
    
    def test_get_schema_type_with_200_response(self):
        """Test extracting schema type from 200 response."""
        method_data = {
            'responses': {
                '200': {
                    'content': {
                        'application/json': {
                            'schema': {
                                'type': 'object'
                            }
                        }
                    }
                }
            }
        }
        result = get_schema_type(method_data)
        assert result == 'object'
    
    def test_get_schema_type_with_ref(self):
        """Test extracting schema type from $ref."""
        method_data = {
            'responses': {
                '200': {
                    'content': {
                        'application/json': {
                            'schema': {
                                '$ref': '#/components/schemas/User'
                            }
                        }
                    }
                }
            }
        }
        result = get_schema_type(method_data)
        assert result == 'User'
    
    def test_get_schema_type_with_default_response(self):
        """Test extracting schema type from default response when 200 is not available."""
        method_data = {
            'responses': {
                'default': {
                    'content': {
                        'application/json': {
                            'schema': {
                                'type': 'array'
                            }
                        }
                    }
                }
            }
        }
        result = get_schema_type(method_data)
        assert result == 'array'
    
    def test_get_schema_type_with_top_level_schema(self):
        """Test extracting schema type from top-level schema."""
        method_data = {
            'responses': {
                '200': {
                    'schema': {
                        'type': 'string'
                    }
                }
            }
        }
        result = get_schema_type(method_data)
        assert result == 'string'
    
    def test_get_schema_type_with_no_schema(self):
        """Test when no schema is available."""
        method_data = {
            'responses': {
                '200': {
                    'content': {
                        'application/json': {}
                    }
                }
            }
        }
        result = get_schema_type(method_data)
        assert result == ''
    
    def test_get_schema_type_with_no_responses(self):
        """Test when no responses are available."""
        method_data = {}
        result = get_schema_type(method_data)
        assert result == ''
    
    def test_get_schema_type_with_non_json_content(self):
        """Test when content is not application/json."""
        method_data = {
            'responses': {
                '200': {
                    'content': {
                        'text/plain': {
                            'schema': {
                                'type': 'string'
                            }
                        }
                    }
                }
            }
        }
        result = get_schema_type(method_data)
        assert result == ''


class TestOpenAPIScannerMain:
    """Test cases for the main function."""
    
    def test_main_with_valid_yaml(self):
        """Test main function with valid OpenAPI YAML."""
        yaml_content = """
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: array
  /users/{id}:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
"""
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            f.write(yaml_content)
            temp_file = f.name
        
        try:
            with patch('src.openapi_scanner.console') as mock_console:
                # Handle SystemExit from Click
                try:
                    main([temp_file])
                except SystemExit:
                    pass  # Expected for Click commands
                # Verify that console.print was called (indicating the table was displayed)
                assert mock_console.print.called
        finally:
            os.unlink(temp_file)
    
    def test_main_with_no_paths(self):
        """Test main function with YAML that has no paths."""
        yaml_content = """
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
"""
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            f.write(yaml_content)
            temp_file = f.name
        
        try:
            with patch('src.openapi_scanner.console') as mock_console:
                # Handle SystemExit from Click
                try:
                    main([temp_file])
                except SystemExit:
                    pass  # Expected for Click commands
                # Verify that the error message was printed
                mock_console.print.assert_called_with('[red]No paths found in the OpenAPI spec.[/red]')
        finally:
            os.unlink(temp_file)
    
    def test_main_with_invalid_yaml(self):
        """Test main function with invalid YAML."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            f.write("invalid: yaml: content: [")
            temp_file = f.name
        
        try:
            with pytest.raises((yaml.YAMLError, SystemExit)):
                main([temp_file])
        finally:
            os.unlink(temp_file)
    
    def test_main_with_nonexistent_file(self):
        """Test main function with nonexistent file."""
        with pytest.raises((FileNotFoundError, SystemExit)):
            main(['nonexistent.yaml']) 