#!/usr/bin/env python3
"""
OpenAPI Pattern Analyzer

This script provides detailed analysis of regex patterns found in OpenAPI specifications.
"""

import json
import re
from typing import Dict, List, Any, Set, Optional
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.syntax import Syntax
from collections import Counter

console = Console()


def analyze_patterns(report_file: str):
    """Analyze patterns from the JSON report file."""
    
    with open(report_file, 'r') as f:
        data = json.load(f)
    
    patterns = data['schema_patterns']
    unconstrained_strings = data.get('unconstrained_strings', [])
    
    console.print(Panel.fit("🔍 Detailed Pattern Analysis", style="bold blue"))
    
    # Pattern Categories
    categorize_patterns(patterns)
    
    # Pattern Complexity Analysis
    analyze_pattern_complexity(patterns)
    
    # Common Pattern Types
    identify_common_patterns(patterns)
    
    # Security Analysis
    analyze_security_patterns(patterns)
    
    # Validation Issues
    check_validation_issues(patterns)
    
    # Unconstrained Strings Analysis
    if unconstrained_strings:
        analyze_unconstrained_strings(unconstrained_strings)


def categorize_patterns(patterns: List[Dict[str, Any]]):
    """Categorize patterns by their purpose."""
    
    console.print("\n" + "="*80)
    console.print("📊 PATTERN CATEGORIES", style="bold green")
    console.print("="*80)
    
    categories = {
        'Resource Quantities': [],
        'URLs and URIs': [],
        'Identifiers and Names': [],
        'Email Addresses': [],
        'File Paths': [],
        'Filter Expressions': [],
        'Version Numbers': [],
        'UUIDs': [],
        'Other': []
    }
    
    for pattern in patterns:
        p = pattern['pattern']
        desc = pattern['description'].lower()
        
        # Resource quantities (memory, CPU, etc.)
        if re.search(r'[+-]?\d+\.?\d*[eEinumkKMGTP]*', p) and any(word in desc for word in ['memory', 'cpu', 'size', 'limit', 'request']):
            categories['Resource Quantities'].append(pattern)
        
        # URLs and URIs
        elif 'http' in p or 'url' in desc or 'uri' in desc:
            categories['URLs and URIs'].append(pattern)
        
        # Identifiers and names
        elif re.search(r'[a-zA-Z][a-zA-Z0-9_-]*', p) and any(word in desc for word in ['name', 'id', 'key', 'alias']):
            categories['Identifiers and Names'].append(pattern)
        
        # Email addresses
        elif '@' in p or 'email' in desc:
            categories['Email Addresses'].append(pattern)
        
        # File paths
        elif '\\x2F' in p or 'path' in desc:
            categories['File Paths'].append(pattern)
        
        # Filter expressions
        elif '==' in p or 'filter' in desc:
            categories['Filter Expressions'].append(pattern)
        
        # Version numbers
        elif re.search(r'v\d+', p) or 'version' in desc:
            categories['Version Numbers'].append(pattern)
        
        # UUIDs
        elif re.search(r'[0-9a-fA-F]{8}', p) or 'uuid' in desc:
            categories['UUIDs'].append(pattern)
        
        else:
            categories['Other'].append(pattern)
    
    # Display categories
    for category, pattern_list in categories.items():
        if pattern_list:
            console.print(f"\n🔹 {category} ({len(pattern_list)} patterns):")
            for pattern in pattern_list:
                console.print(f"   • {pattern['schema_name']}: {pattern['pattern']}")
                if pattern['description']:
                    console.print(f"     {pattern['description'][:80]}...")


def analyze_pattern_complexity(patterns: List[Dict[str, Any]]):
    """Analyze the complexity of patterns."""
    
    console.print("\n" + "="*80)
    console.print("🧮 PATTERN COMPLEXITY ANALYSIS", style="bold green")
    console.print("="*80)
    
    complexity_scores = []
    
    for pattern in patterns:
        p = pattern['pattern']
        score = 0
        
        # Count special characters
        score += len(re.findall(r'[\[\](){}|+*?^$\\]', p))
        
        # Count character classes
        score += len(re.findall(r'\[[^\]]*\]', p))
        
        # Count quantifiers
        score += len(re.findall(r'[+*?]\??', p))
        
        # Count groups
        score += len(re.findall(r'\([^)]*\)', p))
        
        # Count anchors
        score += len(re.findall(r'[\^$]', p))
        
        complexity_scores.append((pattern, score))
    
    # Sort by complexity
    complexity_scores.sort(key=lambda x: x[1], reverse=True)
    
    table = Table(title="Pattern Complexity", show_header=True, header_style="bold magenta")
    table.add_column("Schema", style="cyan", width=25)
    table.add_column("Pattern", style="yellow", width=40, overflow="fold")
    table.add_column("Complexity Score", style="red", width=15)
    table.add_column("Description", style="green", width=30, overflow="fold")
    
    for pattern, score in complexity_scores[:10]:  # Top 10 most complex
        complexity_level = "High" if score > 10 else "Medium" if score > 5 else "Low"
        table.add_row(
            pattern['schema_name'],
            pattern['pattern'],
            f"{score} ({complexity_level})",
            pattern['description'][:25] + "..." if len(pattern['description']) > 25 else pattern['description']
        )
    
    console.print(table)


def identify_common_patterns(patterns: List[Dict[str, Any]]):
    """Identify and analyze common pattern types."""
    
    console.print("\n" + "="*80)
    console.print("🔄 COMMON PATTERN TYPES", style="bold green")
    console.print("="*80)
    
    # Extract pattern templates
    pattern_templates = []
    
    for pattern in patterns:
        p = pattern['pattern']
        
        # Replace specific values with placeholders
        template = p
        
        # Replace numbers with N
        template = re.sub(r'\d+', 'N', template)
        
        # Replace common units with U
        template = re.sub(r'[eEinumkKMGTP]+', 'U', template)
        
        # Replace common field names with F
        template = re.sub(r'(projectName|departmentName|clusterId)', 'F', template)
        
        pattern_templates.append((template, pattern))
    
    # Count template occurrences
    template_counts = Counter(template for template, _ in pattern_templates)
    
    console.print("\nMost common pattern templates:")
    for template, count in template_counts.most_common(5):
        console.print(f"\n🔸 Template: {template} (used {count} times)")
        
        # Show examples
        examples = [p for t, p in pattern_templates if t == template]
        for example in examples[:3]:  # Show first 3 examples
            console.print(f"   • {example['schema_name']}: {example['pattern']}")


def analyze_security_patterns(patterns: List[Dict[str, Any]]):
    """Analyze patterns for potential security implications."""
    
    console.print("\n" + "="*80)
    console.print("🔒 SECURITY PATTERN ANALYSIS", style="bold green")
    console.print("="*80)
    
    security_concerns = []
    
    for pattern in patterns:
        p = pattern['pattern']
        concerns = []
        
        # Check for overly permissive patterns
        if p == '.*' or p == '.*.*':
            concerns.append("Overly permissive - matches anything")
        
        # Check for potential injection patterns
        if re.search(r'[<>"\']', p):
            concerns.append("Contains special characters that might allow injection")
        
        # Check for weak email validation
        if '@' in p and not re.search(r'@[^@]+\.[^@]+', p):
            concerns.append("Weak email validation pattern")
        
        # Check for weak URL validation
        if 'http' in p and not re.search(r'https?://[^\s]+', p):
            concerns.append("Weak URL validation pattern")
        
        if concerns:
            security_concerns.append((pattern, concerns))
    
    if security_concerns:
        console.print("\n⚠️  Potential security concerns found:")
        for pattern, concerns in security_concerns:
            console.print(f"\n🔸 {pattern['schema_name']}:")
            console.print(f"   Pattern: {pattern['pattern']}")
            for concern in concerns:
                console.print(f"   ⚠️  {concern}")
    else:
        console.print("\n✅ No obvious security concerns found in patterns")


def check_validation_issues(patterns: List[Dict[str, Any]]):
    """Check for potential validation issues."""
    
    console.print("\n" + "="*80)
    console.print("🔍 VALIDATION ISSUES", style="bold green")
    console.print("="*80)
    
    issues = []
    
    for pattern in patterns:
        p = pattern['pattern']
        
        # Check for invalid regex
        try:
            re.compile(p)
        except re.error as e:
            issues.append((pattern, f"Invalid regex: {e}"))
            continue
        
        # Check for potentially problematic patterns
        if p.startswith('.*') and p.endswith('.*'):
            issues.append((pattern, "Pattern starts and ends with .* - might be too permissive"))
        
        if len(p) < 3:
            issues.append((pattern, "Very short pattern - might be too simple"))
        
        # Check for missing anchors
        if not p.startswith('^') and not p.startswith('\\A'):
            issues.append((pattern, "Pattern doesn't start with anchor - might match substrings"))
        
        if not p.endswith('$') and not p.endswith('\\Z'):
            issues.append((pattern, "Pattern doesn't end with anchor - might match substrings"))
    
    if issues:
        console.print("\n⚠️  Potential validation issues found:")
        for pattern, issue in issues:
            console.print(f"\n🔸 {pattern['schema_name']}:")
            console.print(f"   Pattern: {pattern['pattern']}")
            console.print(f"   Issue: {issue}")
    else:
        console.print("\n✅ No obvious validation issues found")


def analyze_unconstrained_strings(unconstrained_strings: List[Dict[str, Any]]):
    """Analyze unconstrained strings and provide recommendations."""
    
    console.print("\n" + "="*80)
    console.print("⚠️  UNCONSTRAINED STRINGS ANALYSIS", style="bold yellow")
    console.print("="*80)
    
    # Categorize unconstrained strings by potential validation needs
    categories = {
        'Email Fields': [],
        'URL/URI Fields': [],
        'Name/Identifier Fields': [],
        'Description Fields': [],
        'Message Fields': [],
        'Path Fields': [],
        'Version Fields': [],
        'Other': []
    }
    
    for string_schema in unconstrained_strings:
        name = string_schema['schema_name'].lower()
        desc = string_schema['description'].lower()
        path = string_schema['path'].lower()
        
        # Email fields
        if 'email' in name or 'email' in desc or 'email' in path:
            categories['Email Fields'].append(string_schema)
        
        # URL/URI fields
        elif any(word in name or word in desc or word in path for word in ['url', 'uri', 'link', 'endpoint']):
            categories['URL/URI Fields'].append(string_schema)
        
        # Name/Identifier fields
        elif any(word in name or word in desc or word in path for word in ['name', 'id', 'key', 'alias', 'identifier']):
            categories['Name/Identifier Fields'].append(string_schema)
        
        # Description fields
        elif 'description' in name or 'description' in desc or 'description' in path:
            categories['Description Fields'].append(string_schema)
        
        # Message fields
        elif 'message' in name or 'message' in desc or 'message' in path:
            categories['Message Fields'].append(string_schema)
        
        # Path fields
        elif 'path' in name or 'path' in desc or 'path' in path:
            categories['Path Fields'].append(string_schema)
        
        # Version fields
        elif 'version' in name or 'version' in desc or 'version' in path:
            categories['Version Fields'].append(string_schema)
        
        else:
            categories['Other'].append(string_schema)
    
    # Display categories with recommendations
    for category, string_list in categories.items():
        if string_list:
            console.print(f"\n🔹 {category} ({len(string_list)} fields):")
            
            # Show first few examples
            for string_schema in string_list[:5]:
                console.print(f"   • {string_schema['schema_name']}: {string_schema['path']}")
                if string_schema['description']:
                    console.print(f"     {string_schema['description'][:60]}...")
            
            if len(string_list) > 5:
                console.print(f"   ... and {len(string_list) - 5} more")
            
            # Provide specific recommendations
            if category == 'Email Fields':
                console.print("   💡 Recommendation: Add email format validation")
            elif category == 'URL/URI Fields':
                console.print("   💡 Recommendation: Add URL format validation")
            elif category == 'Name/Identifier Fields':
                console.print("   💡 Recommendation: Add pattern validation for consistent naming")
            elif category == 'Description Fields':
                console.print("   💡 Recommendation: Consider maxLength constraint")
            elif category == 'Message Fields':
                console.print("   💡 Recommendation: Consider maxLength and content validation")
            elif category == 'Path Fields':
                console.print("   💡 Recommendation: Add path format validation")
            elif category == 'Version Fields':
                console.print("   💡 Recommendation: Add version format validation")
    
    # Priority recommendations
    console.print("\n" + "="*80)
    console.print("🎯 PRIORITY RECOMMENDATIONS", style="bold yellow")
    console.print("="*80)
    
    # Find high-priority fields
    high_priority = []
    for string_schema in unconstrained_strings:
        name = string_schema['schema_name'].lower()
        desc = string_schema['description'].lower()
        
        # High priority: authentication, security, or critical fields
        if any(word in name or word in desc for word in ['password', 'secret', 'token', 'key', 'auth', 'credential']):
            high_priority.append(("🔴 HIGH", string_schema, "Security-sensitive field needs validation"))
        elif any(word in name or word in desc for word in ['email', 'url', 'uri']):
            high_priority.append(("🟡 MEDIUM", string_schema, "Common format field needs validation"))
        elif any(word in name or word in desc for word in ['name', 'id', 'identifier']):
            high_priority.append(("🟢 LOW", string_schema, "Identifier field should have consistent format"))
    
    if high_priority:
        console.print("\nPrioritized validation recommendations:")
        for priority, string_schema, reason in high_priority[:10]:  # Show top 10
            console.print(f"\n{priority} Priority:")
            console.print(f"   • {string_schema['schema_name']}: {string_schema['path']}")
            console.print(f"   • Reason: {reason}")
            if string_schema['description']:
                console.print(f"   • Description: {string_schema['description'][:50]}...")
    else:
        console.print("\nNo high-priority validation needs identified.")


def generate_pattern_summary(patterns: List[Dict[str, Any]], unconstrained_strings: Optional[List[Dict[str, Any]]] = None):
    """Generate a summary of all patterns."""
    
    console.print("\n" + "="*80)
    console.print("📋 PATTERN SUMMARY", style="bold green")
    console.print("="*80)
    
    # Basic statistics
    total_patterns = len(patterns)
    unique_patterns = len(set(p['pattern'] for p in patterns))
    
    # Pattern length statistics
    lengths = [len(p['pattern']) for p in patterns]
    if lengths:
        avg_length = sum(lengths) / len(lengths)
        min_length = min(lengths)
        max_length = max(lengths)
        
        console.print(f"\n📊 Basic Statistics:")
        console.print(f"   • Total patterns: {total_patterns}")
        console.print(f"   • Unique patterns: {unique_patterns}")
        console.print(f"   • Average pattern length: {avg_length:.1f} characters")
        console.print(f"   • Shortest pattern: {min_length} characters")
        console.print(f"   • Longest pattern: {max_length} characters")
    
    # Most used patterns
    pattern_counts = Counter(p['pattern'] for p in patterns)
    console.print(f"\n🔄 Most frequently used patterns:")
    for pattern, count in pattern_counts.most_common(3):
        console.print(f"   • {pattern} (used {count} times)")
    
    # Validation coverage analysis
    if unconstrained_strings:
        total_strings = len(patterns) + len(unconstrained_strings)
        coverage_percentage = (len(patterns) / total_strings) * 100
        
        console.print(f"\n📊 Validation Coverage Analysis:")
        console.print(f"   • Constrained strings: {len(patterns)} ({coverage_percentage:.1f}%)")
        console.print(f"   • Unconstrained strings: {len(unconstrained_strings)} ({100 - coverage_percentage:.1f}%)")
        
        if coverage_percentage < 20:
            console.print(f"   ⚠️  Low validation coverage - consider adding more constraints")
        elif coverage_percentage < 50:
            console.print(f"   🟡 Moderate validation coverage - room for improvement")
        else:
            console.print(f"   ✅ Good validation coverage")


if __name__ == '__main__':
    import sys
    
    if len(sys.argv) != 2:
        console.print("[red]Usage: python pattern_analyzer.py <report_file.json>[/red]")
        sys.exit(1)
    
    report_file = sys.argv[1]
    
    try:
        analyze_patterns(report_file)
        with open(report_file, 'r') as f:
            data = json.load(f)
        generate_pattern_summary(data['schema_patterns'], data.get('unconstrained_strings', []))
    except FileNotFoundError:
        console.print(f"[red]Report file not found: {report_file}[/red]")
        sys.exit(1)
    except Exception as e:
        console.print(f"[red]Error analyzing patterns: {e}[/red]")
        sys.exit(1) 