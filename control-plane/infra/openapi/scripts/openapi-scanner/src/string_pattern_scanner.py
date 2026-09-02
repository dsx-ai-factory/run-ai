#!/usr/bin/env python3
"""
OpenAPI String Pattern Scanner

This script scans OpenAPI YAML files and generates a report of all string schemas
that validate string values against regex patterns.
"""

import yaml
import click
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.text import Text
import os
import re
from typing import Dict, List, Any, Optional, Tuple

try:
    console = Console(width=os.get_terminal_size().columns)
except OSError:
    console = Console(width=120)


class StringPatternScanner:
    """Scanner for finding string schemas with regex pattern validation."""
    
    def __init__(self, yaml_data: Dict[str, Any]):
        self.yaml_data = yaml_data
        self.string_patterns = []
        self.unconstrained_strings = []
        
    def scan_schemas(self) -> List[Dict[str, Any]]:
        """Scan all schemas in the OpenAPI spec for string patterns."""
        schemas = self.yaml_data.get('components', {}).get('schemas', {})
        
        for schema_name, schema_def in schemas.items():
            self._scan_schema(schema_name, schema_def, [schema_name])
            
        return self.string_patterns
    
    def scan_unconstrained_strings(self) -> List[Dict[str, Any]]:
        """Scan all schemas for string types without constraints."""
        schemas = self.yaml_data.get('components', {}).get('schemas', {})
        
        for schema_name, schema_def in schemas.items():
            self._scan_unconstrained_schema(schema_name, schema_def, [schema_name])
            
        return self.unconstrained_strings
    
    def _scan_schema(self, schema_name: str, schema_def: Dict[str, Any], path: List[str]):
        """Recursively scan a schema definition for string patterns."""
        if not isinstance(schema_def, dict):
            return
            
        # Check if this is a string schema with pattern
        if schema_def.get('type') == 'string' and 'pattern' in schema_def:
            self.string_patterns.append({
                'schema_name': schema_name,
                'path': ' -> '.join(path),
                'pattern': schema_def['pattern'],
                'description': schema_def.get('description', ''),
                'format': schema_def.get('format', ''),
                'min_length': schema_def.get('minLength'),
                'max_length': schema_def.get('maxLength'),
                'enum': schema_def.get('enum'),
                'example': schema_def.get('example', ''),
                'nullable': schema_def.get('nullable', False),
                'required': schema_def.get('required', False)
            })
        
        # Handle allOf, anyOf, oneOf
        for key in ['allOf', 'anyOf', 'oneOf']:
            if key in schema_def:
                for i, sub_schema in enumerate(schema_def[key]):
                    self._scan_schema(f"{schema_name}_{key}_{i}", sub_schema, path + [f"{key}[{i}]"])
        
        # Handle properties
        if 'properties' in schema_def:
            for prop_name, prop_def in schema_def['properties'].items():
                self._scan_schema(prop_name, prop_def, path + [prop_name])
        
        # Handle items (for arrays)
        if 'items' in schema_def:
            self._scan_schema(f"{schema_name}_item", schema_def['items'], path + ['items'])
        
        # Handle additionalProperties
        if 'additionalProperties' in schema_def:
            if isinstance(schema_def['additionalProperties'], dict):
                self._scan_schema(f"{schema_name}_additional", schema_def['additionalProperties'], path + ['additionalProperties'])
    
    def _scan_unconstrained_schema(self, schema_name: str, schema_def: Dict[str, Any], path: List[str]):
        """Recursively scan a schema definition for unconstrained strings."""
        if not isinstance(schema_def, dict):
            return
            
        # Check if this is a string schema without constraints
        if schema_def.get('type') == 'string':
            has_constraints = any(key in schema_def for key in [
                'pattern', 'enum', 'minLength', 'maxLength', 'format'
            ])
            
            if not has_constraints:
                self.unconstrained_strings.append({
                    'schema_name': schema_name,
                    'path': ' -> '.join(path),
                    'description': schema_def.get('description', ''),
                    'format': schema_def.get('format', ''),
                    'example': schema_def.get('example', ''),
                    'nullable': schema_def.get('nullable', False),
                    'required': schema_def.get('required', False)
                })
        
        # Handle allOf, anyOf, oneOf
        for key in ['allOf', 'anyOf', 'oneOf']:
            if key in schema_def:
                for i, sub_schema in enumerate(schema_def[key]):
                    self._scan_unconstrained_schema(f"{schema_name}_{key}_{i}", sub_schema, path + [f"{key}[{i}]"])
        
        # Handle properties
        if 'properties' in schema_def:
            for prop_name, prop_def in schema_def['properties'].items():
                self._scan_unconstrained_schema(prop_name, prop_def, path + [prop_name])
        
        # Handle items (for arrays)
        if 'items' in schema_def:
            self._scan_unconstrained_schema(f"{schema_name}_item", schema_def['items'], path + ['items'])
        
        # Handle additionalProperties
        if 'additionalProperties' in schema_def:
            if isinstance(schema_def['additionalProperties'], dict):
                self._scan_unconstrained_schema(f"{schema_name}_additional", schema_def['additionalProperties'], path + ['additionalProperties'])
    
    def scan_parameters(self) -> List[Dict[str, Any]]:
        """Scan parameters for string patterns."""
        paths = self.yaml_data.get('paths', {})
        parameter_patterns = []
        
        for path, methods in paths.items():
            for method, method_def in methods.items():
                if not isinstance(method_def, dict):
                    continue
                
                # Check parameters
                parameters = method_def.get('parameters', [])
                for param in parameters:
                    if isinstance(param, dict) and param.get('schema'):
                        schema = param['schema']
                        if schema.get('type') == 'string' and 'pattern' in schema:
                            parameter_patterns.append({
                                'path': path,
                                'method': method.upper(),
                                'parameter_name': param.get('name', ''),
                                'parameter_in': param.get('in', ''),
                                'pattern': schema['pattern'],
                                'description': param.get('description', ''),
                                'required': param.get('required', False)
                            })
        
        return parameter_patterns
    
    def scan_request_bodies(self) -> List[Dict[str, Any]]:
        """Scan request bodies for string patterns."""
        paths = self.yaml_data.get('paths', {})
        request_patterns = []
        
        for path, methods in paths.items():
            for method, method_def in methods.items():
                if not isinstance(method_def, dict):
                    continue
                
                request_body = method_def.get('requestBody', {})
                if request_body and 'content' in request_body:
                    for content_type, content_def in request_body['content'].items():
                        if 'schema' in content_def:
                            schema = content_def['schema']
                            # Create a temporary scanner for the request body schema
                            temp_scanner = StringPatternScanner({'components': {'schemas': {'request_body': schema}}})
                            temp_patterns = temp_scanner.scan_schemas()
                            
                            for pattern in temp_patterns:
                                pattern['path'] = path
                                pattern['method'] = method.upper()
                                pattern['content_type'] = content_type
                                request_patterns.append(pattern)
        
        return request_patterns


def validate_regex_pattern(pattern: str) -> Tuple[bool, str]:
    """Validate if a regex pattern is valid."""
    try:
        re.compile(pattern)
        return True, ""
    except re.error as e:
        return False, str(e)


def generate_report(string_patterns: List[Dict[str, Any]], 
                   parameter_patterns: List[Dict[str, Any]], 
                   request_patterns: List[Dict[str, Any]],
                   unconstrained_strings: List[Dict[str, Any]]) -> None:
    """Generate a comprehensive report of string patterns."""
    
    console.print(Panel.fit("🔍 OpenAPI String Pattern Validation Report", style="bold blue"))
    
    # Summary
    total_patterns = len(string_patterns) + len(parameter_patterns) + len(request_patterns)
    console.print(f"\n📊 Summary: Found {total_patterns} string schemas with regex pattern validation")
    console.print(f"   • Schema definitions: {len(string_patterns)}")
    console.print(f"   • Parameters: {len(parameter_patterns)}")
    console.print(f"   • Request bodies: {len(request_patterns)}")
    console.print(f"   • Unconstrained strings: {len(unconstrained_strings)}")
    
    # Schema Patterns
    if string_patterns:
        console.print("\n" + "="*80)
        console.print("📋 SCHEMA DEFINITIONS WITH PATTERN VALIDATION", style="bold green")
        console.print("="*80)
        
        table = Table(title="String Schema Patterns", show_header=True, header_style="bold magenta")
        table.add_column("Schema Name", style="cyan", width=30, overflow="fold")
        table.add_column("Pattern", style="yellow", width=40, overflow="fold")
        table.add_column("Description", style="green", width=30, overflow="fold")
        table.add_column("Format", style="blue", width=15)
        table.add_column("Validation", style="red", width=15)
        
        for pattern in string_patterns:
            is_valid, error = validate_regex_pattern(pattern['pattern'])
            validation_status = "✅ Valid" if is_valid else f"❌ {error[:10]}..."
            
            table.add_row(
                pattern['schema_name'],
                pattern['pattern'],
                pattern['description'][:25] + "..." if len(pattern['description']) > 25 else pattern['description'],
                pattern['format'] or "-",
                validation_status
            )
        
        console.print(table)
    
    # Unconstrained Strings
    if unconstrained_strings:
        console.print("\n" + "="*80)
        console.print("⚠️  UNCONSTRAINED STRING SCHEMAS", style="bold yellow")
        console.print("="*80)
        
        table = Table(title="Unconstrained String Schemas", show_header=True, header_style="bold magenta")
        table.add_column("Schema Name", style="cyan", width=30, overflow="fold")
        table.add_column("Path", style="yellow", width=40, overflow="fold")
        table.add_column("Description", style="green", width=30, overflow="fold")
        table.add_column("Format", style="blue", width=15)
        table.add_column("Nullable", style="red", width=10)
        
        for string_schema in unconstrained_strings:
            table.add_row(
                string_schema['schema_name'],
                string_schema['path'],
                string_schema['description'][:25] + "..." if len(string_schema['description']) > 25 else string_schema['description'],
                string_schema['format'] or "-",
                "Yes" if string_schema['nullable'] else "No"
            )
        
        console.print(table)
        
        # Recommendations for unconstrained strings
        console.print("\n💡 Recommendations for unconstrained strings:")
        console.print("   • Consider adding pattern validation for data consistency")
        console.print("   • Add minLength/maxLength constraints where appropriate")
        console.print("   • Use enum for fields with predefined values")
        console.print("   • Add format validation for common types (email, uri, etc.)")
    
    # Parameter Patterns
    if parameter_patterns:
        console.print("\n" + "="*80)
        console.print("🔗 PARAMETER PATTERNS", style="bold green")
        console.print("="*80)
        
        table = Table(title="Parameter Patterns", show_header=True, header_style="bold magenta")
        table.add_column("Path", style="cyan", width=30, overflow="fold")
        table.add_column("Method", style="blue", width=10)
        table.add_column("Parameter", style="yellow", width=20)
        table.add_column("Pattern", style="green", width=40, overflow="fold")
        table.add_column("Validation", style="red", width=15)
        
        for pattern in parameter_patterns:
            is_valid, error = validate_regex_pattern(pattern['pattern'])
            validation_status = "✅ Valid" if is_valid else f"❌ {error[:10]}..."
            
            table.add_row(
                pattern['path'],
                pattern['method'],
                pattern['parameter_name'],
                pattern['pattern'],
                validation_status
            )
        
        console.print(table)
    
    # Request Body Patterns
    if request_patterns:
        console.print("\n" + "="*80)
        console.print("📤 REQUEST BODY PATTERNS", style="bold green")
        console.print("="*80)
        
        table = Table(title="Request Body Patterns", show_header=True, header_style="bold magenta")
        table.add_column("Path", style="cyan", width=30, overflow="fold")
        table.add_column("Method", style="blue", width=10)
        table.add_column("Schema", style="yellow", width=20)
        table.add_column("Pattern", style="green", width=40, overflow="fold")
        table.add_column("Validation", style="red", width=15)
        
        for pattern in request_patterns:
            is_valid, error = validate_regex_pattern(pattern['pattern'])
            validation_status = "✅ Valid" if is_valid else f"❌ {error[:10]}..."
            
            table.add_row(
                pattern['path'],
                pattern['method'],
                pattern['schema_name'],
                pattern['pattern'],
                validation_status
            )
        
        console.print(table)
    
    # Pattern Analysis
    console.print("\n" + "="*80)
    console.print("📈 PATTERN ANALYSIS", style="bold green")
    console.print("="*80)
    
    all_patterns = []
    all_patterns.extend([p['pattern'] for p in string_patterns])
    all_patterns.extend([p['pattern'] for p in parameter_patterns])
    all_patterns.extend([p['pattern'] for p in request_patterns])
    
    # Count unique patterns
    unique_patterns = list(set(all_patterns))
    console.print(f"Unique patterns found: {len(unique_patterns)}")
    
    # Most common patterns
    pattern_counts = {}
    for pattern in all_patterns:
        pattern_counts[pattern] = pattern_counts.get(pattern, 0) + 1
    
    if pattern_counts:
        console.print("\nMost common patterns:")
        sorted_patterns = sorted(pattern_counts.items(), key=lambda x: x[1], reverse=True)
        for pattern, count in sorted_patterns[:5]:
            console.print(f"  • {pattern} (used {count} times)")
    
    # Pattern categories
    email_patterns = [p for p in unique_patterns if 'email' in p.lower() or '@' in p]
    url_patterns = [p for p in unique_patterns if 'http' in p.lower() or 'url' in p.lower()]
    uuid_patterns = [p for p in unique_patterns if 'uuid' in p.lower() or '[0-9a-fA-F]{8}' in p]
    version_patterns = [p for p in unique_patterns if 'version' in p.lower() or 'v\\d' in p]
    
    console.print(f"\nPattern categories:")
    console.print(f"  • Email patterns: {len(email_patterns)}")
    console.print(f"  • URL patterns: {len(url_patterns)}")
    console.print(f"  • UUID patterns: {len(uuid_patterns)}")
    console.print(f"  • Version patterns: {len(version_patterns)}")
    
    # String validation coverage
    total_strings = len(string_patterns) + len(unconstrained_strings)
    if total_strings > 0:
        coverage_percentage = (len(string_patterns) / total_strings) * 100
        console.print(f"\n📊 String validation coverage:")
        console.print(f"  • Constrained strings: {len(string_patterns)} ({coverage_percentage:.1f}%)")
        console.print(f"  • Unconstrained strings: {len(unconstrained_strings)} ({100 - coverage_percentage:.1f}%)")


@click.command()
@click.argument('yaml_file', type=click.Path(exists=True))
@click.option('--output', '-o', type=click.Path(), help='Output file for the report (JSON format)')
def main(yaml_file: str, output: Optional[str]):
    """Scan an OpenAPI YAML file for string schemas with regex pattern validation."""
    
    console.print(f"🔍 Scanning OpenAPI file: {yaml_file}")
    
    try:
        with open(yaml_file, 'r') as f:
            data = yaml.safe_load(f)
    except Exception as e:
        console.print(f"[red]Error reading YAML file: {e}[/red]")
        return
    
    # Create scanner and scan for patterns
    scanner = StringPatternScanner(data)
    
    console.print("Scanning schema definitions...")
    string_patterns = scanner.scan_schemas()
    
    console.print("Scanning for unconstrained strings...")
    unconstrained_strings = scanner.scan_unconstrained_strings()
    
    console.print("Scanning parameters...")
    parameter_patterns = scanner.scan_parameters()
    
    console.print("Scanning request bodies...")
    request_patterns = scanner.scan_request_bodies()
    
    # Generate report
    generate_report(string_patterns, parameter_patterns, request_patterns, unconstrained_strings)
    
    # Save to file if requested
    if output:
        import json
        report_data = {
            'summary': {
                'total_patterns': len(string_patterns) + len(parameter_patterns) + len(request_patterns),
                'schema_patterns': len(string_patterns),
                'parameter_patterns': len(parameter_patterns),
                'request_patterns': len(request_patterns),
                'unconstrained_strings': len(unconstrained_strings)
            },
            'schema_patterns': string_patterns,
            'parameter_patterns': parameter_patterns,
            'request_patterns': request_patterns,
            'unconstrained_strings': unconstrained_strings
        }
        
        try:
            with open(output, 'w') as f:
                json.dump(report_data, f, indent=2)
            console.print(f"\n[green]Report saved to: {output}[/green]")
        except Exception as e:
            console.print(f"[red]Error saving report: {e}[/red]")


if __name__ == '__main__':
    main() 