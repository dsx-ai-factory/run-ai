#!/usr/bin/env python3
"""
Test suite for core.config module
"""

import unittest
import tempfile
import sys
from pathlib import Path

# Add parent directory to path to import core modules
sys.path.insert(0, str(Path(__file__).parent.parent))

from core.config import Config, Environment, ProcessingMode, Colors, RegexPatterns


class TestConfig(unittest.TestCase):
    """Test cases for Config class"""
    
    def test_get_default(self):
        """Test default configuration creation"""
        config = Config.get_default()
        
        self.assertEqual(config.NVIDIA_DOCS_DOMAIN, "https://run-ai-docs.nvidia.com")
        self.assertEqual(config.LEGACY_DOCS_DOMAIN, "https://docs.run.ai") 
        self.assertEqual(config.MIN_NEW_DOCS_VERSION, 220)
        self.assertEqual(config.FILE_ENCODING, "utf-8")


class TestEnvironment(unittest.TestCase):
    """Test cases for Environment enum"""
    
    def test_environment_values(self):
        """Test all environment values"""
        self.assertEqual(Environment.LAB.value, "lab")
        self.assertEqual(Environment.TEST.value, "test")
        self.assertEqual(Environment.STAGING.value, "staging")
        self.assertEqual(Environment.SAAS.value, "saas")
    
    def test_environment_count(self):
        """Test that we have the expected number of environments"""
        environments = list(Environment)
        self.assertEqual(len(environments), 4)


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
    
    def test_environment_patterns(self):
        """Test environment detection patterns"""
        # Staging pattern
        self.assertIsNotNone(RegexPatterns.STAGING.search("2.22.0-staging"))
        self.assertIsNone(RegexPatterns.STAGING.search("2.22.0"))
        
        # SAAS pattern
        self.assertIsNotNone(RegexPatterns.SAAS.search("2.22.0-saas.1"))
        self.assertIsNone(RegexPatterns.SAAS.search("2.22.0"))
        
        # Test pattern
        self.assertIsNotNone(RegexPatterns.TEST.search("2.22.0-test"))
        self.assertIsNone(RegexPatterns.TEST.search("2.22.0"))
    
    def test_semver_pattern(self):
        """Test semantic version pattern"""
        pattern = RegexPatterns.SEMVER
        
        # Should match semantic versions
        self.assertIsNotNone(pattern.match("2.22.0"))
        self.assertIsNotNone(pattern.match("1.0.0"))
        self.assertIsNotNone(pattern.match("10.5.3"))
        
        # Should not match partial versions
        self.assertIsNone(pattern.match("2.22"))
        self.assertIsNone(pattern.match("2.22.0.master"))
        self.assertIsNone(pattern.match("v2.22.0"))
    
    def test_content_replacement_patterns(self):
        """Test content replacement patterns"""
        # Version replacement
        version_pattern = RegexPatterns.VERSION_REPLACE
        self.assertIsNotNone(version_pattern.search("version: 2.0.0"))
        self.assertIsNone(version_pattern.search("version: 2.22.0"))
        
        # Branding replacement
        branding_pattern = RegexPatterns.BRANDING
        self.assertIsNotNone(branding_pattern.search("Run:AI platform"))
        self.assertIsNone(branding_pattern.search("Run:ai platform"))
        
        # Logo replacement
        logo_pattern = RegexPatterns.LOGO
        self.assertIsNotNone(logo_pattern.search("url: ./runai-logo-api.png"))
        self.assertIsNone(logo_pattern.search("url: https://example.com/logo.png"))
    
    def test_get_docs_pattern(self):
        """Test documentation pattern generation"""
        docs_domain = "https://run-ai-docs.nvidia.com"
        pattern = RegexPatterns.get_docs_pattern(docs_domain)
        
        # Should escape dots properly
        self.assertIn(r'run-ai-docs\.nvidia\.com', pattern)
        self.assertIn('/self-hosted/latest/', pattern)


if __name__ == '__main__':
    unittest.main(verbosity=2) 