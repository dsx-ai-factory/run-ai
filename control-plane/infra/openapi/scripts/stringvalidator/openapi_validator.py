import os
import sys

class OpenAPIValidator:
    def __init__(self, openapi_spec, valid_regexes):
        self.openapi_spec = openapi_spec
        self.valid_regexes = valid_regexes

    def resolve_reference(self, ref_path):
        path_parts = ref_path.lstrip('#/').split('/')
        current = self.openapi_spec
        for part in path_parts:
            current = current.get(part)
            if current is None:
                raise ValueError(f"Invalid reference path: {ref_path}")
        return current

    def validate_string_properties(self, schema):
        errors = []

        if '$ref' in schema:
            schema = self.resolve_reference(schema['$ref'])

        if schema.get('type') == 'object' and 'properties' in schema:
            for prop_name, prop_schema in schema['properties'].items():
                if prop_schema.get('type') == 'string':
                    if not any(key in prop_schema for key in ['enum', 'pattern', 'format']):
                        errors.append({
                            'name': prop_name,
                            'message': f"String property missing enum/pattern/format: {prop_schema}"
                        })
                    elif 'pattern' in prop_schema:
                        pattern = prop_schema['pattern']
                        if pattern not in self.valid_regexes:
                            errors.append({
                                'name': prop_name,
                                'message': f"Invalid pattern '{pattern}' for string property: {prop_schema}"
                            })
                else:
                    errors.extend(self.validate_string_properties(prop_schema))

        elif schema.get('type') == 'array' and 'items' in schema:
            errors.extend(self.validate_string_properties(schema['items']))

        elif schema.get('type') == 'string':
            if not any(key in schema for key in ['enum', 'pattern', 'format']):
                errors.append({
                    'name': 'Unknown Property',
                    'message': f"String property missing enum/pattern/format: {schema}"
                })
            elif 'pattern' in schema:
                pattern = schema['pattern']
                if pattern not in self.valid_regexes:
                    errors.append({
                        'name': 'Unknown Property',
                        'message': f"Invalid pattern '{pattern}' for string property: {schema}"
                    })

        return errors


    def validate_post_put_apis(self):
        errors = []

        for path, operations in self.openapi_spec.get('paths', {}).items():
            # Check path-level parameters
            if 'parameters' in operations:
                for param in operations['parameters']:
                    if param.get('schema', {}).get('type') == 'string':
                        if not any(key in param.get('schema', {}) for key in ['enum', 'pattern', 'format']):
                            errors.append(f"Path '{path}' parameter '{param.get('name', 'Unknown Parameter')}' missing enum/pattern/format: {param}")
                        elif 'pattern' in param.get('schema', {}):
                            pattern = param['schema']['pattern']
                            if pattern not in self.valid_regexes:
                                errors.append(f"Path '{path}' parameter '{param.get('name', 'Unknown Parameter')}' has invalid pattern '{pattern}': {param}")

            for method, operation in operations.items():
                if method.lower() in ['post', 'put']:
                    # Check operation-level parameters
                    if 'parameters' in operation:
                        for param in operation['parameters']:
                            if param.get('schema', {}).get('type') == 'string':
                                if not any(key in param.get('schema', {}) for key in ['enum', 'pattern', 'format']):
                                    errors.append(f"Path '{path}' operation '{method.upper()}' parameter '{param.get('name', 'Unknown Parameter')}' missing enum/pattern/format: {param}")
                                elif 'pattern' in param.get('schema', {}):
                                    pattern = param['schema']['pattern']
                                    if pattern not in self.valid_regexes:
                                        errors.append(f"Path '{path}' operation '{method.upper()}' parameter '{param.get('name', 'Unknown Parameter')}' has invalid pattern '{pattern}': {param}")

                    # Check requestBody
                    request_body = operation.get('requestBody')
                    if request_body and 'content' in request_body:
                        for media_type, media_schema in request_body['content'].items():
                            schema = media_schema.get('schema')
                            if schema:
                                for error in self.validate_string_properties(schema):
                                    # Try to extract more context if available
                                    if isinstance(error, dict):
                                        errors.append(f"Path '{path}' operation '{method.upper()}' request body: {error.get('name', 'Unknown Property')} - {error.get('message', error)}")
                                    else:
                                        errors.append(f"Path '{path}' operation '{method.upper()}' request body: {error}")

        return errors