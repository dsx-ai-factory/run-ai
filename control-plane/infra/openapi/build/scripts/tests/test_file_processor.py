#!/usr/bin/env python3
"""
Test suite for core.file_processor module
"""

import unittest
import tempfile
import os
import sys
from pathlib import Path

# Add parent directory to path to import core modules
sys.path.insert(0, str(Path(__file__).parent.parent))

from core.file_processor import FileProcessor
from core.config import Config


class TestFileProcessor(unittest.TestCase):
    """Test cases for FileProcessor class"""
    
    def setUp(self):
        """Set up test fixtures"""
        self.config = Config.get_default()
        
        # Create temporary test files
        self.test_dir = tempfile.mkdtemp()
        self.spec_file_path = Path(self.test_dir) / "test_spec.yaml"
        self.spec_file_src = Path(self.test_dir) / "test_spec_src.yaml"
        
        # Sample OpenAPI content
        self.sample_content = """openapi: 3.0.0
info:
  title: Test API
  description: |
    To assign access rules, see [Create access rules](https://run-ai-docs.nvidia.com/self-hosted/latest/infrastructure-setup/authentication/accessrules#create-or-delete-rules).
    Use the [Roles](https://run-ai-docs.nvidia.com/self-hosted/latest/infrastructure-setup/authentication/roles) to assign rules.
    Get your token from [Request a token](https://run-ai-docs.nvidia.com/self-hosted/latest/reference/api/rest-auth/#request-an-api-token).
  version: 2.0.0
  x-logo:
    url: ./runai-logo-api.png
    altText: 'NVIDIA Run:AI'
paths:
  /test:
    get:
      summary: Test endpoint
      responses:
        '200':
          description: Success
"""
        
        # Write initial content to source file
        self.spec_file_src.write_text(self.sample_content, encoding='utf-8')
        self.spec_file_path.write_text(self.sample_content, encoding='utf-8')
        
        # Create FileProcessor instance
        self.file_processor = FileProcessor(self.spec_file_path, self.spec_file_src, self.config)
    
    def tearDown(self):
        """Clean up test fixtures"""
        import shutil
        shutil.rmtree(self.test_dir, ignore_errors=True)
    
    def test_init(self):
        """Test FileProcessor initialization"""
        self.assertIsInstance(self.file_processor.spec_file_path, Path)
        self.assertIsInstance(self.file_processor.spec_file_src, Path)
        self.assertIsNotNone(self.file_processor.config)
        self.assertIsNotNone(self.file_processor.link_replacer)
        # Cache should be None until file is read
        self.assertIsNone(self.file_processor._cached_content)
    
    def test_validate_files(self):
        """Test file validation"""
        # Should not raise for existing files
        self.file_processor.validate_files()
        
        # Should raise for non-existent spec file
        non_existent_spec = Path(self.test_dir) / "non_existent.yaml"
        processor = FileProcessor(non_existent_spec, self.spec_file_src, self.config)
        with self.assertRaises(FileNotFoundError):
            processor.validate_files()
        
        # Should raise for non-existent source file
        non_existent_src = Path(self.test_dir) / "non_existent_src.yaml"
        processor = FileProcessor(self.spec_file_path, non_existent_src, self.config)
        with self.assertRaises(FileNotFoundError):
            processor.validate_files()
    
    def test_read_spec_file(self):
        """Test reading spec file content"""
        content = self.file_processor.read_spec_file()
        self.assertEqual(content, self.sample_content)
        
        # Test caching
        content2 = self.file_processor.read_spec_file()
        self.assertEqual(content, content2)
    
    def test_write_spec_file(self):
        """Test writing spec file content"""
        new_content = "new test content"
        self.file_processor.write_spec_file(new_content)
        
        # Verify content was written
        actual_content = self.spec_file_path.read_text(encoding='utf-8')
        self.assertEqual(actual_content, new_content)
        
        # Verify cache was updated
        cached_content = self.file_processor.read_spec_file()
        self.assertEqual(cached_content, new_content)
    
    def test_reset_spec_file(self):
        """Test resetting spec file to source content"""
        # Modify the spec file
        modified_content = "modified content"
        self.file_processor.write_spec_file(modified_content)
        
        # Reset to source (method now returns None, no exception means success)
        self.file_processor.reset_spec_file()
        
        # Verify content was reset
        content = self.file_processor.read_spec_file()
        self.assertEqual(content, self.sample_content)
    
    def test_apply_version_transformation(self):
        """Test version transformation"""
        # Read the file first to populate cache
        self.file_processor.read_spec_file()
        content = self.file_processor._cached_content
        result = self.file_processor._replace_version(content, "2.22.0")
        self.assertIn('version: "2.22.0"', result)
        self.assertNotIn('version: 2.0.0', result)
    
    def test_apply_branding_transformation(self):
        """Test branding transformation"""
        # Read the file first to populate cache
        self.file_processor.read_spec_file()
        content = self.file_processor._cached_content
        result = self.file_processor._replace_branding(content)
        self.assertIn('Run:ai', result)
        self.assertNotIn('Run:AI', result)
    
    def test_apply_docs_link_transformation(self):
        """Test documentation link transformation"""
        docs_base_url = "https://run-ai-docs.nvidia.com/self-hosted"
        # Read the file first
        content = self.file_processor.read_spec_file()
        result = self.file_processor.link_replacer.replace_docs_links(content, docs_base_url)
        self.assertIn(docs_base_url, result)
    
    def test_apply_build_transformations(self):
        """Test applying all build transformations"""
        version = "2.22.0"
        docs_base_url = "https://run-ai-docs.nvidia.com/self-hosted"
        
        self.file_processor.apply_build_transformations(version, docs_base_url)
        content = self.file_processor.read_spec_file()
        
        # Check that transformations were applied
        self.assertIn('"2.22.0"', content)  # Version should be quoted
        self.assertIn(docs_base_url, content)
        # Logo should remain local for build mode
        self.assertIn('./runai-logo-api.png', content)
    
    def test_error_handling_invalid_file(self):
        """Test error handling for invalid files"""
        # Create a file path that points to a non-existent directory
        # This will reliably cause a write error on all platforms
        invalid_dir = Path(self.test_dir) / "non_existent_dir"
        invalid_file = invalid_dir / "test.yaml"
        
        processor = FileProcessor(invalid_file, self.spec_file_src, Config.get_default())
        
        # Should fail to write to a file in a non-existent directory
        with self.assertRaises(RuntimeError) as context:
            processor.write_spec_file("new content")
        
        # Verify the error message mentions the expected file path and context
        self.assertIn(str(invalid_file), str(context.exception))
        self.assertIn("Failed to write spec file", str(context.exception))
    
    def test_encoding_handling(self):
        """Test proper encoding handling"""
        # Test with unicode content
        unicode_content = "Test with unicode: 中文 日本語 한국어"
        self.file_processor.write_spec_file(unicode_content)
        
        # Read back and verify
        content = self.file_processor.read_spec_file()
        self.assertEqual(content, unicode_content)
        
        # Verify file on disk has correct encoding
        disk_content = self.spec_file_path.read_text(encoding='utf-8')
        self.assertEqual(disk_content, unicode_content)
    
    def test_large_file_handling(self):
        """Test handling of larger files"""
        # Create a larger spec file
        large_content = self.sample_content + "\n" + "  " * 10000 + "# Large comment"
        self.spec_file_path.write_text(large_content, encoding='utf-8')
        
        content = self.file_processor.read_spec_file()
        result = self.file_processor._replace_version(content, "2.23.0")
        
        self.assertIn('version: "2.23.0"', result)
        self.assertTrue(len(result) > len(self.sample_content))


class TestFileProcessorPathValidation(unittest.TestCase):
    """Test path validation in FileProcessor (S2083 path injection fix)"""

    def setUp(self):
        self.config = Config.get_default()
        self.test_dir = tempfile.mkdtemp()
        self.valid_yaml = Path(self.test_dir) / "spec.yaml"
        self.valid_yaml.write_text("openapi: 3.0.0\n", encoding="utf-8")

    def tearDown(self):
        import shutil
        shutil.rmtree(self.test_dir, ignore_errors=True)

    def test_disallowed_extension_raises(self):
        """Paths with non-spec extensions must be rejected"""
        for bad_ext in [".sh", ".txt", ".py", ".exe", ""]:
            bad_path = Path(self.test_dir) / f"malicious{bad_ext}"
            with self.assertRaises(ValueError, msg=f"Extension {bad_ext!r} should be rejected"):
                FileProcessor(bad_path, self.valid_yaml, self.config)

    def test_allowed_extensions_accepted(self):
        """Paths with .yaml, .yml, and .json extensions must be accepted"""
        for ext in [".yaml", ".yml", ".json"]:
            path = Path(self.test_dir) / f"spec{ext}"
            path.write_text("{}\n", encoding="utf-8")
            # Should not raise
            fp = FileProcessor(path, self.valid_yaml, self.config)
            self.assertEqual(fp.spec_file_path, path.resolve())

    def test_path_traversal_canonicalized(self):
        """Paths with '..' components within trusted root are canonicalized, not rejected"""
        subdir = Path(self.test_dir) / "subdir"
        subdir.mkdir()
        traversal = subdir / ".." / "spec.yaml"
        fp = FileProcessor(traversal, self.valid_yaml, self.config,
                           trusted_root=Path(self.test_dir))
        self.assertEqual(fp.spec_file_path, traversal.resolve())
        self.assertNotIn("..", str(fp.spec_file_path))

    def test_path_traversal_outside_root_rejected(self):
        """Paths that traverse outside the trusted root must raise ValueError"""
        outside_dir = Path(self.test_dir).parent
        outside_spec = outside_dir / "escape.yaml"
        # spec_file_path escapes the trusted root
        with self.assertRaises(ValueError):
            FileProcessor(outside_spec, self.valid_yaml, self.config,
                          trusted_root=Path(self.test_dir))
        # spec_file_src escapes the trusted root
        with self.assertRaises(ValueError):
            FileProcessor(self.valid_yaml, outside_spec, self.config,
                          trusted_root=Path(self.test_dir))

    def test_absolute_path_escaping_root_is_rejected(self):
        """Absolute paths with allowed extensions that escape root are rejected"""
        absolute_outside = Path("/tmp/spec.yaml")
        with self.assertRaises(ValueError):
            FileProcessor(absolute_outside, self.valid_yaml, self.config,
                          trusted_root=Path(self.test_dir))

    def test_absolute_path_stored(self):
        """Stored paths are always absolute after validation"""
        fp = FileProcessor(self.valid_yaml, self.valid_yaml, self.config)
        self.assertTrue(fp.spec_file_path.is_absolute())
        self.assertTrue(fp.spec_file_src.is_absolute())

    def test_no_trusted_root_skips_boundary_check(self):
        """Without a trusted_root the boundary check is skipped (backward compat)"""
        import tempfile
        import shutil
        outside_dir = tempfile.mkdtemp()
        try:
            outside = Path(outside_dir) / "spec.yaml"
            outside.write_text("{}\n", encoding="utf-8")
            # Should not raise — no trusted_root configured
            fp = FileProcessor(outside, self.valid_yaml, self.config)
            self.assertEqual(fp.spec_file_path, outside.resolve())
        finally:
            shutil.rmtree(outside_dir, ignore_errors=True)


if __name__ == '__main__':
    unittest.main(verbosity=2)