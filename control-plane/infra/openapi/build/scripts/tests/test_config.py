#!/usr/bin/env python3

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Test suite for core.config module
"""

import unittest
import tempfile
import sys
from pathlib import Path

# Add parent directory to path to import core modules
sys.path.insert(0, str(Path(__file__).parent.parent))

from core.config import Config, ProcessingMode, Colors, RegexPatterns


class TestConfig(unittest.TestCase):
    """Test cases for Config class"""
    
    def test_get_default(self):
        """Test default configuration creation"""
        config = Config.get_default()
        
        self.assertEqual(config.FILE_ENCODING, "utf-8")


class TestProcessingMode(unittest.TestCase):
    """Test cases for ProcessingMode enum"""
    
    def test_processing_mode_values(self):
        """Test processing mode values"""
        self.assertEqual(ProcessingMode.BUILD.value, "build")

    def test_processing_mode_count(self):
        """Test that we have the expected number of processing modes"""
        modes = list(ProcessingMode)
        self.assertEqual(len(modes), 1)


class TestColors(unittest.TestCase):
    """Test cases for Colors class"""
    
    def test_color_constants(self):
        """Test that color constants are defined"""
        self.assertIsInstance(Colors.RED, str)
        self.assertIsInstance(Colors.GREEN, str)
        self.assertIsInstance(Colors.YELLOW, str)
        self.assertIsInstance(Colors.BLUE, str)
        self.assertIsInstance(Colors.NC, str)
        
        # Test that colors contain ANSI escape sequences
        self.assertTrue(Colors.RED.startswith('\033['))
        self.assertTrue(Colors.GREEN.startswith('\033['))
        self.assertTrue(Colors.YELLOW.startswith('\033['))
        self.assertTrue(Colors.BLUE.startswith('\033['))
        self.assertEqual(Colors.NC, '\033[0m')


class TestRegexPatterns(unittest.TestCase):
    """Test cases for RegexPatterns class"""
    
    def test_version_pattern(self):
        """Test version regex pattern"""
        pattern = RegexPatterns.VERSION
        
        # Should match version patterns
        self.assertIsNotNone(pattern.search("2.20"))
        self.assertIsNotNone(pattern.search("2.21.0"))
        self.assertIsNotNone(pattern.search("v2.20"))  # Now accepts v prefix
        self.assertIsNotNone(pattern.search("v2.21.0"))  # Now accepts v prefix
        self.assertIsNotNone(pattern.search("2.22.1-saas.10"))
        
        # Should not match invalid patterns
        self.assertIsNone(pattern.search("abc"))
        self.assertIsNone(pattern.search(""))
        self.assertIsNone(pattern.search("latest"))
    
    def test_master_pattern(self):
        """Test master version pattern"""
        pattern = RegexPatterns.MASTER
        
        # Should match master versions
        self.assertIsNotNone(pattern.search("2.22.0.master"))
        self.assertIsNotNone(pattern.search("2.21.1.master"))
        
        # Should not match non-master versions
        self.assertIsNone(pattern.search("2.22.0"))
        self.assertIsNone(pattern.search("master"))
    
    def test_content_replacement_patterns(self):
        """Test content replacement patterns"""
        # Version replacement
        version_pattern = RegexPatterns.VERSION_REPLACE
        self.assertIsNotNone(version_pattern.search("version: 2.0.0"))
        self.assertIsNone(version_pattern.search("version: 2.22.0"))


if __name__ == '__main__':
    unittest.main(verbosity=2) 