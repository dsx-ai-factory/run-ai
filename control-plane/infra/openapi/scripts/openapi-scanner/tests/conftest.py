import pytest
import tempfile
import os
import json
import yaml


@pytest.fixture
def sample_openapi_yaml():
    """Sample OpenAPI YAML data for testing."""
    return """
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      parameters:
        - name: search
          in: query
          schema:
            type: string
            pattern: '^[a-zA-Z0-9\\s]+$'
      responses:
        '200':
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '201':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
  /users/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            pattern: '^[a-zA-Z0-9_-]+$'
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          pattern: '^[a-zA-Z0-9_-]+$'
          description: User identifier
        name:
          type: string
          pattern: '^[a-zA-Z\\s]+$'
          description: User name
          minLength: 1
          maxLength: 100
        email:
          type: string
          pattern: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$'
          format: email
          description: Email address
        age:
          type: integer
          minimum: 0
          maximum: 150
        description:
          type: string
          description: User description
      required:
        - name
        - email
    Product:
      type: object
      properties:
        id:
          type: string
          pattern: '^[A-Z]{2}\\d{6}$'
          description: Product ID
        name:
          type: string
          pattern: '^[a-zA-Z0-9\\s_-]+$'
          description: Product name
        category:
          type: string
          enum: [electronics, clothing, books, other]
        metadata:
          type: object
          additionalProperties:
            type: string
            pattern: '^[a-zA-Z0-9]+$'
"""


@pytest.fixture
def sample_pattern_report():
    """Sample pattern analysis report for testing."""
    return {
        'summary': {
            'total_patterns': 6,
            'schema_patterns': 4,
            'parameter_patterns': 2,
            'request_patterns': 0
        },
        'schema_patterns': [
            {
                'schema_name': 'id',
                'path': 'User -> id',
                'pattern': '^[a-zA-Z0-9_-]+$',
                'description': 'User identifier',
                'format': None,
                'min_length': None,
                'max_length': None,
                'enum': None,
                'example': 'user123',
                'nullable': False,
                'required': False
            },
            {
                'schema_name': 'name',
                'path': 'User -> name',
                'pattern': '^[a-zA-Z\\s]+$',
                'description': 'User name',
                'format': None,
                'min_length': 1,
                'max_length': 100,
                'enum': None,
                'example': 'John Doe',
                'nullable': False,
                'required': True
            },
            {
                'schema_name': 'email',
                'path': 'User -> email',
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
                'schema_name': 'product_id',
                'path': 'Product -> id',
                'pattern': '^[A-Z]{2}\\d{6}$',
                'description': 'Product ID',
                'format': None,
                'min_length': None,
                'max_length': None,
                'enum': None,
                'example': 'PR123456',
                'nullable': False,
                'required': False
            }
        ],
        'parameter_patterns': [
            {
                'path': '/users',
                'method': 'GET',
                'parameter_name': 'search',
                'parameter_in': 'query',
                'pattern': '^[a-zA-Z0-9\\s]+$',
                'description': '',
                'required': False
            },
            {
                'path': '/users/{id}',
                'method': 'GET',
                'parameter_name': 'id',
                'parameter_in': 'path',
                'pattern': '^[a-zA-Z0-9_-]+$',
                'description': '',
                'required': True
            }
        ],
        'request_patterns': [],
        'unconstrained_strings': [
            {
                'schema_name': 'description',
                'path': 'User -> description',
                'description': 'User description',
                'format': None,
                'example': '',
                'nullable': False,
                'required': False
            }
        ]
    }


@pytest.fixture
def temp_yaml_file(sample_openapi_yaml):
    """Create a temporary YAML file with sample OpenAPI data."""
    with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
        f.write(sample_openapi_yaml)
        temp_file = f.name
    
    yield temp_file
    
    # Cleanup
    if os.path.exists(temp_file):
        os.unlink(temp_file)


@pytest.fixture
def temp_json_file(sample_pattern_report):
    """Create a temporary JSON file with sample pattern report data."""
    with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
        json.dump(sample_pattern_report, f)
        temp_file = f.name
    
    yield temp_file
    
    # Cleanup
    if os.path.exists(temp_file):
        os.unlink(temp_file)


@pytest.fixture
def temp_yaml_data(sample_openapi_yaml):
    """Parse sample OpenAPI YAML data."""
    return yaml.safe_load(sample_openapi_yaml) 