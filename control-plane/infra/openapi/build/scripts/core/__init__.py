"""
Core functionality for OpenAPI spec build processing.

This package provides modular components for:
- Version management and parsing
- Documentation link replacement
- File processing and transformations
- Configuration management
"""

from .config import Config, Environment, ProcessingMode, Colors
from .version_manager import VersionManager, VersionInfo
from .link_replacer import LinkReplacer
from .file_processor import FileProcessor
from .reporter import Reporter
from .base_processor import BaseProcessor

__all__ = [
    'Config',
    'Environment',
    'ProcessingMode',
    'Colors',
    'VersionManager',
    'VersionInfo',
    'LinkReplacer',
    'FileProcessor',
    'Reporter',
    'BaseProcessor',
]