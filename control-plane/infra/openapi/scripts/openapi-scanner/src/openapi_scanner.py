import yaml
import click
from rich.console import Console
from rich.table import Table
import os

try:
    console = Console(width=os.get_terminal_size().columns)
except OSError:
    console = Console(width=120)

def get_schema_type(method_data):
    """Extract the schema type from the 200/default response if available."""
    responses = method_data.get('responses', {})
    # Prefer 200, then default, then any available
    for code in ['200', 'default'] + list(responses.keys()):
        resp = responses.get(code, {})
        content = resp.get('content', {})
        if 'application/json' in content:
            schema = content['application/json'].get('schema', {})
            if '$ref' in schema:
                return schema['$ref'].split('/')[-1]  # Show the referenced schema name
            if 'type' in schema:
                return schema['type']
        # fallback: check for schema at the top level
        if 'schema' in resp:
            schema = resp['schema']
            if '$ref' in schema:
                return schema['$ref'].split('/')[-1]
            if 'type' in schema:
                return schema['type']
    return ''

@click.command()
@click.argument('yaml_file', type=click.Path(exists=True))
def main(yaml_file):
    """Parse an OpenAPI YAML file and print all paths."""
    with open(yaml_file, 'r') as f:
        data = yaml.safe_load(f)

    paths = data.get('paths', {})
    if not paths:
        console.print('[red]No paths found in the OpenAPI spec.[/red]')
        return

    table = Table(title="OpenAPI Paths", show_header=True, header_style="bold magenta")
    table.add_column("Path", style="cyan", width=50, overflow="fold")
    table.add_column("Methods", style="magenta", width=15)
    table.add_column("Schema Type", style="green", width=20)

    for path, methods in paths.items():
        method_names = []
        schema_types = []
        for method, method_data in methods.items():
            if not isinstance(method_data, dict):
                continue
            method_names.append(method)
            schema_types.append(get_schema_type(method_data) or '-')
        table.add_row(path, ', '.join(method_names), ', '.join(schema_types))

    console.print(table)

if __name__ == '__main__':
    main()
