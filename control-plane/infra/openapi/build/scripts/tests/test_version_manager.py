#!/usr/bin/env python3

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Test suite for core.version_manager module
"""

import unittest
import sys
from pathlib import Path

# Add parent directory to path to import core modules
sys.path.insert(0, str(Path(__file__).parent.parent))

from core.version_manager import VersionManager


class TestVersionManager(unittest.TestCase):
    """Test cases for VersionManager class"""

    def test_validate_version_format(self):
        """Test version validation"""
        for version in ["2.25.0", "2.24", "2.25.0.master", "latest", "v2.25.0"]:
            with self.subTest(version=version):
                self.assertTrue(VersionManager.validate_version_format(version))

        for version in ["", "invalid", "master"]:  # Only the .master suffix is valid
            with self.subTest(version=version):
                self.assertFalse(VersionManager.validate_version_format(version))


if __name__ == '__main__':
    unittest.main(verbosity=2)
