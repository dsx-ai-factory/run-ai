#!/usr/bin/env python3
"""
Example script demonstrating how to use the report comparator.

This script shows how to compare two OpenAPI pattern analysis reports
to identify changes between versions.
"""

import sys
import os
from pathlib import Path

# Add the src directory to the Python path
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from report_comparator import ReportComparator, display_comparison, generate_diff_report


def main():
    """Example usage of the report comparator."""
    
    # Example file paths (you would replace these with your actual report files)
    old_report = "report.json"  # Your old report file
    new_report = "enhanced_report.json"  # Your new report file
    
    # Check if the files exist
    if not os.path.exists(old_report):
        print(f"Error: Old report file '{old_report}' not found.")
        print("Please ensure you have generated reports using the string pattern scanner.")
        return 1
    
    if not os.path.exists(new_report):
        print(f"Error: New report file '{new_report}' not found.")
        print("Please ensure you have generated reports using the string pattern scanner.")
        return 1
    
    try:
        # Create comparator and load reports
        print("Loading reports...")
        comparator = ReportComparator(old_report, new_report)
        comparator.load_reports()
        
        # Perform comparison
        print("Comparing reports...")
        comparison = comparator.compare_reports()
        
        # Display results
        display_comparison(comparison)
        
        # Generate a markdown diff report
        diff_report_file = "report_comparison.md"
        print(f"\nGenerating diff report to {diff_report_file}...")
        generate_diff_report(comparison, diff_report_file)
        
        # Summary
        total_changes = (
            len(comparison['added_patterns']) + 
            len(comparison['removed_patterns']) + 
            len(comparison['modified_patterns'])
        )
        
        if total_changes > 0:
            print(f"\n✅ Comparison complete! Found {total_changes} changes.")
            print(f"📄 Detailed report saved to: {diff_report_file}")
        else:
            print("\n✅ Comparison complete! No changes detected between reports.")
        
        return 0
        
    except Exception as e:
        print(f"Error during comparison: {e}")
        return 1


if __name__ == "__main__":
    exit(main()) 