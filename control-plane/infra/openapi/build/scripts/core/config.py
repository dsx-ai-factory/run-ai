"""
Configuration module for OpenAPI spec build processing.
"""

import re
from enum import Enum
from dataclasses import dataclass


class Colors:
    RED = '\033[31m'
    GREEN = '\033[32m'
    YELLOW = '\033[33m'
    BLUE = '\033[34m'
    NC = '\033[0m'


class ProcessingMode(Enum):
    BUILD = "build"  # Local development


class Environment(Enum):
    LAB = "lab"       # Never pushes
    TEST = "test"         
    STAGING = "staging"   
    SAAS = "saas"


class FileConstants:
    DEFAULT_ENCODING = 'utf-8'
    SPEC_FILE_BACKUP_SUFFIX = '.backup'
    

class VersionConstants:
    DEFAULT_API_VERSION = "2.0.0"
    MIN_NEW_DOCS_VERSION = 220  # 2.20 minimum for new docs
    LATEST_VERSION_KEYWORD = "latest"
    MASTER_VERSION_SUFFIX = ".master"
    MASTER_VERSION_KEYWORD = "master"
    VERSION_MULTIPLIER = 100  # For converting "2.20" to 220


class BrandingConstants:
    OLD_BRAND = "Run:AI"
    NEW_BRAND = "Run:ai"


class DocumentationConstants:
    NVIDIA_DOCS_DOMAIN = "https://run-ai-docs.nvidia.com"
    LEGACY_DOCS_DOMAIN = "https://docs.run.ai"





@dataclass
class Config:
    NVIDIA_DOCS_DOMAIN: str = DocumentationConstants.NVIDIA_DOCS_DOMAIN
    LEGACY_DOCS_DOMAIN: str = DocumentationConstants.LEGACY_DOCS_DOMAIN
    MIN_NEW_DOCS_VERSION: int = VersionConstants.MIN_NEW_DOCS_VERSION
    FILE_ENCODING: str = FileConstants.DEFAULT_ENCODING
    
    @classmethod
    def get_default(cls) -> 'Config':
        return cls()


class RegexPatterns:
    # Version parsing
    VERSION = re.compile(r'v?(\d+\.\d+)')
    SEMVER = re.compile(r'^\d+\.\d+\.\d+$')
    
    # Environment detection
    MASTER = re.compile(r'\.master')
    STAGING = re.compile(r'-staging')
    SAAS = re.compile(r'-saas')
    TEST = re.compile(r'-test')
    
    # Content replacement
    VERSION_REPLACE = re.compile(r'version: 2\.0\.0')
    BRANDING = re.compile(r'Run:AI')
    LOGO = re.compile(r'url: \./runai-logo-api\.png')
    
    @classmethod
    def get_docs_pattern(cls, docs_domain: str) -> str:
        docs_domain_escaped = docs_domain.replace('.', r'\.')
        return f"{docs_domain_escaped}/self-hosted/latest/" 