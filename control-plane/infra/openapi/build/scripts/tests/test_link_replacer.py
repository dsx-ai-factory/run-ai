#!/usr/bin/env python3
"""
Test suite for core.link_replacer module
"""

import unittest
import sys
from pathlib import Path

# Add parent directory to path to import core modules
sys.path.insert(0, str(Path(__file__).parent.parent))

from core.link_replacer import LinkReplacer
from core.version_manager import VersionInfo
from core.config import Config


class TestLinkReplacer(unittest.TestCase):
    """Test cases for LinkReplacer class"""

    def setUp(self):
        """Set up test fixtures"""
        self.config = Config.get_default()
        self.link_replacer = LinkReplacer(self.config)

    def test_get_docs_url_build_mode(self):
        """Test documentation URL generation for build mode"""
        # Latest version (>= 2.20)
        version_info = VersionInfo("2.22.0", "2.22", 222, False, True)
        url = self.link_replacer.get_docs_url(version_info)
        self.assertEqual(url, "https://run-ai-docs.nvidia.com/self-hosted")

        # Previous version (>= 2.20)
        version_info = VersionInfo("2.21.0", "2.21", 221, False, False)
        url = self.link_replacer.get_docs_url(version_info)
        self.assertEqual(url, "https://run-ai-docs.nvidia.com/self-hosted/2.21")

        # Legacy version (< 2.20)
        version_info = VersionInfo("2.19.0", "2.19", 219, False, False)
        url = self.link_replacer.get_docs_url(version_info)
        self.assertEqual(url, "https://docs.run.ai/v2.19")

        # Master version
        version_info = VersionInfo("2.22.0.master", "master", None, True, False)
        url = self.link_replacer.get_docs_url(version_info)
        self.assertEqual(url, "https://run-ai-docs.nvidia.com/self-hosted")

    def test_replace_docs_links_saas(self):
        """Test documentation link replacement for SAAS URLs"""
        content = """
        See [documentation](https://run-ai-docs.nvidia.com/self-hosted/latest/guide/)
        and [another link](https://run-ai-docs.nvidia.com/self-hosted/latest/reference/)
        """

        docs_base_url = "https://run-ai-docs.nvidia.com/saas"
        result = self.link_replacer.replace_docs_links(content, docs_base_url)

        self.assertIn("https://run-ai-docs.nvidia.com/saas/guide/", result)
        self.assertIn("https://run-ai-docs.nvidia.com/saas/reference/", result)
        self.assertNotIn("/self-hosted/latest/", result)

    def test_replace_docs_links_self_hosted(self):
        """Test documentation link replacement for self-hosted URLs"""
        content = """
        See [documentation](https://run-ai-docs.nvidia.com/self-hosted/latest/guide/)
        and [another link](https://run-ai-docs.nvidia.com/self-hosted/latest/reference/)
        """

        docs_base_url = "https://run-ai-docs.nvidia.com/self-hosted/2.21"
        result = self.link_replacer.replace_docs_links(content, docs_base_url)

        self.assertIn("https://run-ai-docs.nvidia.com/self-hosted/2.21/guide/", result)
        self.assertIn("https://run-ai-docs.nvidia.com/self-hosted/2.21/reference/", result)
        self.assertNotIn("/latest/", result)

    def test_replace_docs_links_legacy(self):
        """Test documentation link replacement for legacy URLs"""
        content = """
        See [documentation](https://run-ai-docs.nvidia.com/self-hosted/latest/guide/)
        and [another link](https://run-ai-docs.nvidia.com/self-hosted/latest/reference/)
        """

        docs_base_url = "https://docs.run.ai/v2.19"
        result = self.link_replacer.replace_docs_links(content, docs_base_url)

        # Legacy docs should skip automatic replacement
        self.assertEqual(result, content)  # Should be unchanged

    def test_replace_docs_links_no_changes(self):
        """Test that content without matching links is unchanged"""
        content = """
        See [external link](https://example.com/guide/)
        and [another external](https://other.com/reference/)
        """

        docs_base_url = "https://run-ai-docs.nvidia.com/saas"
        result = self.link_replacer.replace_docs_links(content, docs_base_url)

        self.assertEqual(result, content)  # Should be unchanged

    def test_get_build_mode_url_internal(self):
        """Test the internal _get_build_mode_url method"""
        # Latest version (>= 2.20)
        version_info = VersionInfo("2.22.0", "2.22", 222, False, True)
        url = self.link_replacer._get_build_mode_url(version_info)
        self.assertEqual(url, "https://run-ai-docs.nvidia.com/self-hosted")

        # Previous version (>= 2.20)
        version_info = VersionInfo("2.21.0", "2.21", 221, False, False)
        url = self.link_replacer._get_build_mode_url(version_info)
        self.assertEqual(url, "https://run-ai-docs.nvidia.com/self-hosted/2.21")

        # Legacy version (< 2.20)
        version_info = VersionInfo("2.19.0", "2.19", 219, False, False)
        url = self.link_replacer._get_build_mode_url(version_info)
        self.assertEqual(url, "https://docs.run.ai/v2.19")

        # Master version (no numeric)
        version_info = VersionInfo("master", "master", None, True, False)
        url = self.link_replacer._get_build_mode_url(version_info)
        self.assertEqual(url, "https://run-ai-docs.nvidia.com/self-hosted")


if __name__ == '__main__':
    unittest.main(verbosity=2)
