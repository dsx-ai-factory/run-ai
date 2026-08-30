#!/usr/bin/env python3
"""
Test runner for all modular OpenAPI build tests

This script runs all test files and provides a comprehensive test report.
"""

import sys
import unittest
from pathlib import Path


def discover_and_run_tests():
    """Discover and run all tests in the current directory"""
    # Add current directory to Python path for imports
    current_dir = Path(__file__).parent
    sys.path.insert(0, str(current_dir))
    
    # Test discovery
    loader = unittest.TestLoader()
    suite = loader.discover(
        start_dir=str(current_dir),
        pattern='test_*.py',
        top_level_dir=str(current_dir)
    )
    
    # Run tests with detailed output
    runner = unittest.TextTestRunner(
        verbosity=2,
        stream=sys.stdout,
        buffer=True
    )
    
    print("=" * 80)
    print("RUNNING MODULAR OPENAPI BUILD TEST SUITE")
    print("=" * 80)
    
    result = runner.run(suite)
    
    # Print summary
    print("\n" + "=" * 80)
    print("TEST SUMMARY")
    print("=" * 80)
    print(f"Tests run: {result.testsRun}")
    print(f"Failures: {len(result.failures)}")
    print(f"Errors: {len(result.errors)}")
    print(f"Skipped: {len(result.skipped)}")
    
    if result.failures:
        print(f"\nFAILURES ({len(result.failures)}):")
        for test, traceback in result.failures:
            print(f"  - {test}")
    
    if result.errors:
        print(f"\nERRORS ({len(result.errors)}):")
        for test, traceback in result.errors:
            print(f"  - {test}")
    
    if result.skipped:
        print(f"\nSKIPPED ({len(result.skipped)}):")
        for test, reason in result.skipped:
            print(f"  - {test}: {reason}")
    
    # Return success/failure
    success = len(result.failures) == 0 and len(result.errors) == 0
    
    if success:
        print(f"\n🎉 ALL TESTS PASSED!")
    else:
        print(f"\n❌ SOME TESTS FAILED")
    
    return success


def main():
    """Main entry point"""
    print("Starting comprehensive test suite for modular OpenAPI build...")
    
    # Run unittest-based tests
    unittest_success = discover_and_run_tests()
    
    print("\n" + "=" * 80)
    print("FINAL RESULTS")
    print("=" * 80)
    
    if unittest_success:
        print("🎉 ALL TEST SUITES PASSED!")
        sys.exit(0)
    else:
        print("❌ SOME TESTS FAILED")
        sys.exit(1)


if __name__ == "__main__":
    main() 