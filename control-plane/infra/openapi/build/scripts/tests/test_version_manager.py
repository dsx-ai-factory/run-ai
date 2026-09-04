#!/usr/bin/env python3
"""
Test suite for core.version_manager module
"""

import unittest
import sys
from pathlib import Path

# Add parent directory to path to import core modules
sys.path.insert(0, str(Path(__file__).parent.parent))

from core.version_manager import VersionManager, VersionInfo
from core.config import VersionConstants


class TestVersionInfo(unittest.TestCase):
    """Test cases for VersionInfo dataclass"""
    
    def test_version_info_creation(self):
        """Test VersionInfo creation"""
        version_info = VersionInfo(
            original="2.22.0",
            major_minor="2.22",
            numeric=222,
            is_master=False,
            is_latest=True
        )
        
        self.assertEqual(version_info.original, "2.22.0")
        self.assertEqual(version_info.major_minor, "2.22")
        self.assertEqual(version_info.numeric, 222)
        self.assertFalse(version_info.is_master)
        self.assertTrue(version_info.is_latest)
    
    def test_version_info_master(self):
        """Test VersionInfo for master version"""
        version_info = VersionInfo(
            original="2.22.0.master",
            major_minor=VersionConstants.MASTER_VERSION_KEYWORD,
            numeric=None,
            is_master=True,
            is_latest=False
        )
        
        self.assertEqual(version_info.original, "2.22.0.master")
        self.assertEqual(version_info.major_minor, VersionConstants.MASTER_VERSION_KEYWORD)
        self.assertIsNone(version_info.numeric)
        self.assertTrue(version_info.is_master)
        self.assertFalse(version_info.is_latest)


class TestVersionManager(unittest.TestCase):
    """Test cases for VersionManager class"""
    
    def test_extract_major_minor(self):
        """Test extracting major.minor version"""
        # Normal versions
        self.assertEqual(VersionManager.extract_major_minor("2.22.0"), "2.22")
        self.assertEqual(VersionManager.extract_major_minor("2.21"), "2.21")
        self.assertEqual(VersionManager.extract_major_minor("1.0.5"), "1.0")
        
        # Master versions
        self.assertEqual(VersionManager.extract_major_minor("2.22.0.master"), VersionConstants.MASTER_VERSION_KEYWORD)
        self.assertEqual(VersionManager.extract_major_minor("2.21.1.master"), VersionConstants.MASTER_VERSION_KEYWORD)
        
        # Edge cases
        self.assertEqual(VersionManager.extract_major_minor("latest"), "")
        self.assertEqual(VersionManager.extract_major_minor("invalid"), "")
    
    def test_is_latest(self):
        """Test latest version detection"""
        # Matching latest
        self.assertTrue(VersionManager.is_latest("2.22.0", "v2.22.1-saas.10"))
        self.assertTrue(VersionManager.is_latest("2.21", "v2.21.0-staging"))
        
        # Not matching latest
        self.assertFalse(VersionManager.is_latest("2.20.0", "v2.22.1-saas.10"))
        self.assertFalse(VersionManager.is_latest("2.19", "v2.21.0-staging"))
        
        # Master versions - should compare major.minor, not be automatically false
        self.assertTrue(VersionManager.is_latest("2.22.0.master", "v2.22.1-saas.10"))  # Same major.minor
        
        # Edge cases
        self.assertFalse(VersionManager.is_latest("invalid", "v2.22.1-saas.10"))
        self.assertFalse(VersionManager.is_latest("2.22.0", "invalid"))
    

    def test_validate_version(self):
        """Test version validation"""
        # Valid versions
        self.assertTrue(VersionManager.validate_version("2.22.0"))
        self.assertTrue(VersionManager.validate_version("2.21"))
        self.assertTrue(VersionManager.validate_version("2.22.0.master"))
        self.assertTrue(VersionManager.validate_version("latest"))
        self.assertTrue(VersionManager.validate_version("v2.22.0"))  # Now accepts v prefix
        
        # Invalid versions
        self.assertFalse(VersionManager.validate_version(""))
        self.assertFalse(VersionManager.validate_version("invalid"))
        self.assertFalse(VersionManager.validate_version("master"))  # Only .master suffix is valid
    
    def test_parse_version_info(self):
        """Test version info parsing"""
        # Regular version (latest)
        version_info = VersionManager.parse_version_info("2.22.0", "v2.22.1-saas.10")
        self.assertEqual(version_info.original, "2.22.0")
        self.assertEqual(version_info.major_minor, "2.22")
        self.assertEqual(version_info.numeric, 222)
        self.assertFalse(version_info.is_master)
        self.assertTrue(version_info.is_latest)
        
        # Regular version (not latest)
        version_info = VersionManager.parse_version_info("2.21.0", "v2.22.1-saas.10")
        self.assertEqual(version_info.original, "2.21.0")
        self.assertEqual(version_info.major_minor, "2.21")
        self.assertEqual(version_info.numeric, 221)
        self.assertFalse(version_info.is_master)
        self.assertFalse(version_info.is_latest)
        
        # Master version
        version_info = VersionManager.parse_version_info("2.22.0.master", "v2.21.1-saas.10")
        self.assertEqual(version_info.original, "2.22.0.master")
        self.assertEqual(version_info.major_minor, VersionConstants.MASTER_VERSION_KEYWORD)
        self.assertIsNone(version_info.numeric)
        self.assertTrue(version_info.is_master)
        self.assertFalse(version_info.is_latest)
        
        # "latest" version
        version_info = VersionManager.parse_version_info("latest", "v2.22.1-saas.10")
        self.assertEqual(version_info.original, "latest")
        self.assertEqual(version_info.major_minor, "2.22")  # Extracted from latest_tag
        self.assertEqual(version_info.numeric, 222)
        self.assertFalse(version_info.is_master)
        self.assertTrue(version_info.is_latest)
    
    def test_parse_version_number(self):
        """Test internal version number parsing"""
        # Valid versions
        self.assertEqual(VersionManager._convert_version_to_numeric("2.22"), 222)
        self.assertEqual(VersionManager._convert_version_to_numeric("2.20"), 220)
        self.assertEqual(VersionManager._convert_version_to_numeric("1.5"), 105)
        self.assertEqual(VersionManager._convert_version_to_numeric("10.15"), 1015)
        self.assertEqual(VersionManager._convert_version_to_numeric("v2.22"), 222)  # Now accepts v prefix
        
        # Master version
        self.assertIsNone(VersionManager._convert_version_to_numeric(VersionConstants.MASTER_VERSION_KEYWORD))
        
        # Invalid versions
        self.assertIsNone(VersionManager._convert_version_to_numeric("invalid"))
        self.assertIsNone(VersionManager._convert_version_to_numeric(""))
    
    def test_is_latest_version(self):
        """Test internal latest version checking"""
        # Matching versions
        self.assertTrue(VersionManager._does_version_match_latest_tag("2.22", "v2.22.1-saas.10"))
        self.assertTrue(VersionManager._does_version_match_latest_tag("2.21.0", "v2.21.5-staging"))
        
        # Non-matching versions
        self.assertFalse(VersionManager._does_version_match_latest_tag("2.20", "v2.22.1-saas.10"))
        self.assertFalse(VersionManager._does_version_match_latest_tag("2.19.0", "v2.21.5-staging"))
        
        # Edge cases
        self.assertFalse(VersionManager._does_version_match_latest_tag("invalid", "v2.22.1-saas.10"))
        self.assertFalse(VersionManager._does_version_match_latest_tag("2.22", "invalid"))


if __name__ == '__main__':
    unittest.main(verbosity=2) 