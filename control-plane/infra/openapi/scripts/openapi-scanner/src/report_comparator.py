#!/usr/bin/env python3
"""
OpenAPI Report Comparator

This script compares two OpenAPI pattern analysis reports to identify changes
between versions, including added, removed, and modified patterns.
"""

import json
import click
import os
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.text import Text
from rich.syntax import Syntax
from typing import Dict, List, Any, Set, Optional, Tuple
from collections import defaultdict
import difflib

try:
    console = Console(width=os.get_terminal_size().columns)
except OSError:
    console = Console(width=120)


class ReportComparator:
    """Comparator for OpenAPI pattern analysis reports."""
    
    def __init__(self, old_report_path: str, new_report_path: str):
        self.old_report_path = old_report_path
        self.new_report_path = new_report_path
        self.old_data: Optional[Dict[str, Any]] = None
        self.new_data: Optional[Dict[str, Any]] = None
        
    def load_reports(self):
        """Load both reports from JSON files."""
        try:
            with open(self.old_report_path, 'r') as f:
                self.old_data = json.load(f)
            with open(self.new_report_path, 'r') as f:
                self.new_data = json.load(f)
        except FileNotFoundError as e:
            console.print(f"[red]Error: Could not find report file: {e.filename}[/red]")
            raise
        except json.JSONDecodeError as e:
            console.print(f"[red]Error: Invalid JSON in report file: {e}[/red]")
            raise
    
    def compare_reports(self) -> Dict[str, Any]:
        """Compare the two reports and return differences."""
        if not self.old_data or not self.new_data:
            raise ValueError("Reports must be loaded before comparison")
        
        comparison = {
            'summary': self._compare_summary(),
            'added_patterns': [],
            'removed_patterns': [],
            'modified_patterns': [],
            'unchanged_patterns': [],
            'unconstrained_strings_changes': self._compare_unconstrained_strings()
        }
        
        # Compare schema patterns
        old_patterns = {p['schema_name']: p for p in self.old_data.get('schema_patterns', [])}
        new_patterns = {p['schema_name']: p for p in self.new_data.get('schema_patterns', [])}
        
        old_schema_names = set(old_patterns.keys())
        new_schema_names = set(new_patterns.keys())
        
        # Find added patterns
        added_names = new_schema_names - old_schema_names
        comparison['added_patterns'] = [new_patterns[name] for name in added_names]
        
        # Find removed patterns
        removed_names = old_schema_names - new_schema_names
        comparison['removed_patterns'] = [old_patterns[name] for name in removed_names]
        
        # Find modified and unchanged patterns
        common_names = old_schema_names & new_schema_names
        for name in common_names:
            old_pattern = old_patterns[name]
            new_pattern = new_patterns[name]
            
            if self._patterns_are_equal(old_pattern, new_pattern):
                comparison['unchanged_patterns'].append(new_pattern)
            else:
                comparison['modified_patterns'].append({
                    'schema_name': name,
                    'old': old_pattern,
                    'new': new_pattern,
                    'changes': self._get_pattern_changes(old_pattern, new_pattern)
                })
        
        return comparison
    
    def _compare_summary(self) -> Dict[str, Any]:
        """Compare summary statistics between reports."""
        if not self.old_data or not self.new_data:
            raise ValueError("Reports must be loaded before comparison")
            
        old_summary = self.old_data.get('summary', {})
        new_summary = self.new_data.get('summary', {})
        
        return {
            'old': old_summary,
            'new': new_summary,
            'changes': {
                'total_patterns': new_summary.get('total_patterns', 0) - old_summary.get('total_patterns', 0),
                'schema_patterns': new_summary.get('schema_patterns', 0) - old_summary.get('schema_patterns', 0),
                'parameter_patterns': new_summary.get('parameter_patterns', 0) - old_summary.get('parameter_patterns', 0),
                'request_patterns': new_summary.get('request_patterns', 0) - old_summary.get('request_patterns', 0)
            }
        }
    
    def _compare_unconstrained_strings(self) -> Dict[str, Any]:
        """Compare unconstrained strings between reports."""
        if not self.old_data or not self.new_data:
            raise ValueError("Reports must be loaded before comparison")
            
        old_strings = {s['schema_name']: s for s in self.old_data.get('unconstrained_strings', [])}
        new_strings = {s['schema_name']: s for s in self.new_data.get('unconstrained_strings', [])}
        
        old_names = set(old_strings.keys())
        new_names = set(new_strings.keys())
        
        return {
            'added': [new_strings[name] for name in new_names - old_names],
            'removed': [old_strings[name] for name in old_names - new_names],
            'unchanged': [new_strings[name] for name in old_names & new_names]
        }
    
    def _patterns_are_equal(self, pattern1: Dict[str, Any], pattern2: Dict[str, Any]) -> bool:
        """Check if two patterns are equal."""
        key_fields = ['pattern', 'description', 'format', 'min_length', 'max_length', 'enum', 'example', 'nullable', 'required']
        
        for field in key_fields:
            if pattern1.get(field) != pattern2.get(field):
                return False
        return True
    
    def _get_pattern_changes(self, old_pattern: Dict[str, Any], new_pattern: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Get detailed changes between two patterns."""
        changes = []
        key_fields = ['pattern', 'description', 'format', 'min_length', 'max_length', 'enum', 'example', 'nullable', 'required']
        
        for field in key_fields:
            old_value = old_pattern.get(field)
            new_value = new_pattern.get(field)
            
            if old_value != new_value:
                changes.append({
                    'field': field,
                    'old': old_value,
                    'new': new_value
                })
        
        return changes


def display_comparison(comparison: Dict[str, Any]):
    """Display the comparison results in a formatted way."""
    
    console.print(Panel.fit("📊 OpenAPI Report Comparison", style="bold blue"))
    
    # Summary changes
    display_summary_changes(comparison['summary'])
    
    # Pattern changes
    display_pattern_changes(comparison)
    
    # Unconstrained strings changes
    display_unconstrained_strings_changes(comparison['unconstrained_strings_changes'])


def display_summary_changes(summary_comparison: Dict[str, Any]):
    """Display summary statistics changes."""
    
    console.print("\n" + "="*80)
    console.print("📈 SUMMARY CHANGES", style="bold green")
    console.print("="*80)
    
    old_summary = summary_comparison['old']
    new_summary = summary_comparison['new']
    changes = summary_comparison['changes']
    
    table = Table(title="Summary Statistics", show_header=True, header_style="bold magenta")
    table.add_column("Metric", style="cyan", width=20)
    table.add_column("Old", style="yellow", width=10)
    table.add_column("New", style="yellow", width=10)
    table.add_column("Change", style="green", width=15)
    
    for metric in ['total_patterns', 'schema_patterns', 'parameter_patterns', 'request_patterns']:
        old_val = old_summary.get(metric, 0)
        new_val = new_summary.get(metric, 0)
        change = changes.get(metric, 0)
        
        change_str = f"{change:+d}" if change != 0 else "0"
        change_style = "green" if change >= 0 else "red"
        
        table.add_row(
            metric.replace('_', ' ').title(),
            str(old_val),
            str(new_val),
            Text(change_str, style=change_style)
        )
    
    console.print(table)


def display_pattern_changes(comparison: Dict[str, Any]):
    """Display pattern changes."""
    
    console.print("\n" + "="*80)
    console.print("🔄 PATTERN CHANGES", style="bold green")
    console.print("="*80)
    
    # Added patterns
    if comparison['added_patterns']:
        console.print(f"\n✅ Added Patterns ({len(comparison['added_patterns'])}):")
        for pattern in comparison['added_patterns']:
            console.print(f"   • {pattern['schema_name']}: {pattern['pattern']}")
            if pattern['description']:
                console.print(f"     {pattern['description'][:80]}...")
    
    # Removed patterns
    if comparison['removed_patterns']:
        console.print(f"\n❌ Removed Patterns ({len(comparison['removed_patterns'])}):")
        for pattern in comparison['removed_patterns']:
            console.print(f"   • {pattern['schema_name']}: {pattern['pattern']}")
            if pattern['description']:
                console.print(f"     {pattern['description'][:80]}...")
    
    # Modified patterns
    if comparison['modified_patterns']:
        console.print(f"\n🔄 Modified Patterns ({len(comparison['modified_patterns'])}):")
        for change in comparison['modified_patterns']:
            console.print(f"\n   • {change['schema_name']}:")
            for field_change in change['changes']:
                field = field_change['field']
                old_val = field_change['old']
                new_val = field_change['new']
                
                console.print(f"     {field}: {old_val} → {new_val}")
    
    # Unchanged patterns
    if comparison['unchanged_patterns']:
        console.print(f"\n⏺️  Unchanged Patterns ({len(comparison['unchanged_patterns'])}):")
        for pattern in comparison['unchanged_patterns'][:10]:  # Show first 10
            console.print(f"   • {pattern['schema_name']}: {pattern['pattern']}")
        if len(comparison['unchanged_patterns']) > 10:
            console.print(f"   ... and {len(comparison['unchanged_patterns']) - 10} more")


def display_unconstrained_strings_changes(changes: Dict[str, Any]):
    """Display unconstrained strings changes."""
    
    console.print("\n" + "="*80)
    console.print("📝 UNCONSTRAINED STRINGS CHANGES", style="bold green")
    console.print("="*80)
    
    if changes['added']:
        console.print(f"\n✅ Added Unconstrained Strings ({len(changes['added'])}):")
        for string_def in changes['added']:
            console.print(f"   • {string_def['schema_name']}")
    
    if changes['removed']:
        console.print(f"\n❌ Removed Unconstrained Strings ({len(changes['removed'])}):")
        for string_def in changes['removed']:
            console.print(f"   • {string_def['schema_name']}")
    
    if not changes['added'] and not changes['removed']:
        console.print("\n⏺️  No changes to unconstrained strings")


def generate_diff_report(comparison: Dict[str, Any], output_file: Optional[str] = None) -> str:
    """Generate a detailed diff report in markdown format."""
    
    report_lines = []
    report_lines.append("# OpenAPI Report Comparison")
    report_lines.append("")
    
    # Summary
    summary = comparison['summary']
    report_lines.append("## Summary")
    report_lines.append("")
    report_lines.append("| Metric | Old | New | Change |")
    report_lines.append("|--------|-----|-----|--------|")
    
    for metric in ['total_patterns', 'schema_patterns', 'parameter_patterns', 'request_patterns']:
        old_val = summary['old'].get(metric, 0)
        new_val = summary['new'].get(metric, 0)
        change = summary['changes'].get(metric, 0)
        change_str = f"{change:+d}" if change != 0 else "0"
        
        report_lines.append(f"| {metric.replace('_', ' ').title()} | {old_val} | {new_val} | {change_str} |")
    
    report_lines.append("")
    
    # Added patterns
    if comparison['added_patterns']:
        report_lines.append("## Added Patterns")
        report_lines.append("")
        for pattern in comparison['added_patterns']:
            report_lines.append(f"### {pattern['schema_name']}")
            report_lines.append(f"- **Pattern:** `{pattern['pattern']}`")
            if pattern['description']:
                report_lines.append(f"- **Description:** {pattern['description']}")
            report_lines.append("")
    
    # Removed patterns
    if comparison['removed_patterns']:
        report_lines.append("## Removed Patterns")
        report_lines.append("")
        for pattern in comparison['removed_patterns']:
            report_lines.append(f"### {pattern['schema_name']}")
            report_lines.append(f"- **Pattern:** `{pattern['pattern']}`")
            if pattern['description']:
                report_lines.append(f"- **Description:** {pattern['description']}")
            report_lines.append("")
    
    # Modified patterns
    if comparison['modified_patterns']:
        report_lines.append("## Modified Patterns")
        report_lines.append("")
        for change in comparison['modified_patterns']:
            report_lines.append(f"### {change['schema_name']}")
            for field_change in change['changes']:
                field = field_change['field']
                old_val = field_change['old']
                new_val = field_change['new']
                report_lines.append(f"- **{field}:** `{old_val}` → `{new_val}`")
            report_lines.append("")
    
    report_content = "\n".join(report_lines)
    
    if output_file:
        with open(output_file, 'w') as f:
            f.write(report_content)
        console.print(f"[green]Diff report saved to: {output_file}[/green]")
    
    return report_content


def generate_json_report(comparison: Dict[str, Any], output_file: Optional[str] = None) -> str:
    """Generate a detailed diff report in JSON format."""
    
    report_data = {
        'summary': comparison['summary'],
        'added_patterns': comparison['added_patterns'],
        'removed_patterns': comparison['removed_patterns'],
        'modified_patterns': comparison['modified_patterns'],
        'unchanged_patterns': comparison['unchanged_patterns'],
        'unconstrained_strings_changes': comparison['unconstrained_strings_changes'],
        'statistics': {
            'total_changes': (
                len(comparison['added_patterns']) + 
                len(comparison['removed_patterns']) + 
                len(comparison['modified_patterns'])
            ),
            'added_count': len(comparison['added_patterns']),
            'removed_count': len(comparison['removed_patterns']),
            'modified_count': len(comparison['modified_patterns']),
            'unchanged_count': len(comparison['unchanged_patterns'])
        }
    }
    
    report_content = json.dumps(report_data, indent=2)
    
    if output_file:
        with open(output_file, 'w') as f:
            f.write(report_content)
        console.print(f"[green]JSON report saved to: {output_file}[/green]")
    
    return report_content


@click.command()
@click.argument('old_report', type=click.Path(exists=True))
@click.argument('new_report', type=click.Path(exists=True))
@click.option('--output', '-o', type=click.Path(), help='Output file for the diff report (markdown format)')
@click.option('--json-output', '-j', type=click.Path(), help='Output file for the diff report (JSON format)')
@click.option('--format', '-f', type=click.Choice(['markdown', 'json', 'both']), default='markdown', 
              help='Output format for the diff report (default: markdown)')
@click.option('--quiet', '-q', is_flag=True, help='Suppress console output')
def main(old_report: str, new_report: str, output: Optional[str], json_output: Optional[str], 
         format: str, quiet: bool):
    """Compare two OpenAPI pattern analysis reports.
    
    Examples:
        # Compare reports and display results
        python report_comparator.py old_report.json new_report.json
        
        # Generate markdown diff report
        python report_comparator.py old_report.json new_report.json -o diff.md
        
        # Generate JSON diff report
        python report_comparator.py old_report.json new_report.json -j diff.json
        
        # Generate both formats
        python report_comparator.py old_report.json new_report.json -o diff.md -j diff.json
    """
    
    try:
        # Create comparator and load reports
        comparator = ReportComparator(old_report, new_report)
        comparator.load_reports()
        
        # Perform comparison
        comparison = comparator.compare_reports()
        
        # Display results
        if not quiet:
            display_comparison(comparison)
        
        # Generate diff reports based on options
        if output:
            generate_diff_report(comparison, output)
        
        if json_output:
            generate_json_report(comparison, json_output)
        
        # If format is specified but no explicit output file, use default names
        if not output and not json_output:
            if format == 'markdown':
                pass  # No default output for markdown
            elif format == 'json':
                pass  # No default output for JSON
            elif format == 'both':
                pass  # No default output for both
        
        # Exit with appropriate code
        total_changes = (
            len(comparison['added_patterns']) + 
            len(comparison['removed_patterns']) + 
            len(comparison['modified_patterns'])
        )
        
        if total_changes > 0:
            console.print(f"\n[yellow]Total changes detected: {total_changes}[/yellow]")
            return 1  # Exit with code 1 if there are changes
        else:
            console.print("\n[green]No changes detected between reports[/green]")
            return 0
            
    except Exception as e:
        console.print(f"[red]Error: {e}[/red]")
        return 1


if __name__ == '__main__':
    main() 