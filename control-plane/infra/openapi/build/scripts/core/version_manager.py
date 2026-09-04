"""
Version management for OpenAPI spec build processing.
"""

from dataclasses import dataclass
from typing import Optional, List
from .config import Environment, RegexPatterns, VersionConstants


@dataclass
class VersionInfo:
    original: str
    major_minor: str
    numeric: Optional[int]
    is_master: bool
    is_latest: bool


class VersionManager:
    
    @staticmethod
    def extract_major_minor_from_version(version: str) -> str:
        if VersionManager._is_master_version(version):
            return VersionConstants.MASTER_VERSION_KEYWORD
        
        match = RegexPatterns.VERSION.search(version)
        return match.group(1) if match else ""
    
    @staticmethod
    def is_latest_stable_version(version: str, latest_tag: str) -> bool:
        return VersionManager._does_version_match_latest_tag(version, latest_tag)
    
    @staticmethod
    def detect_environment_from_version_and_param(env: str, version: str) -> Environment:
        # Explicit env parameter takes priority over version-based detection
        if env == Environment.LAB.value:
            return Environment.LAB
        if env == Environment.TEST.value:
            return Environment.TEST
        if env == Environment.STAGING.value:
            return Environment.STAGING
        if env == Environment.SAAS.value:
            return Environment.SAAS
        
        # Fallback to version-based detection when env is not explicitly set
        if VersionManager._is_staging_version(version):
            return Environment.STAGING
        
        if (VersionManager._is_saas_version(version) or 
            VersionManager._is_semantic_version(version)):
            return Environment.SAAS
        
        if (VersionManager._is_master_version(version) or 
            VersionManager._is_test_version(version)):
            return Environment.TEST
        
        return Environment.TEST
    
    @staticmethod
    def build_version_string(version: str, latest_tag: str, env: Environment) -> str:
        version_info = VersionManager.parse_version_into_components(version, latest_tag)
        
        if version_info.is_master:
            return VersionConstants.MASTER_VERSION_KEYWORD
        
        if version_info.is_latest and env in [Environment.SAAS, Environment.TEST]:
            return f"{version_info.major_minor},latest"
        
        return version_info.major_minor
    
    @staticmethod
    def parse_version_into_components(version: str, latest_tag: str) -> VersionInfo:
        is_master = VersionManager._is_master_version(version)
        
        if is_master:
            return VersionInfo(
                original=version,
                major_minor=VersionConstants.MASTER_VERSION_KEYWORD,
                numeric=None,
                is_master=True,
                is_latest=False
            )
        
        if version == VersionConstants.LATEST_VERSION_KEYWORD:
            return VersionManager._handle_latest_version_keyword(latest_tag)
        
        major_minor = VersionManager.extract_major_minor_from_version(version)
        numeric_version = VersionManager._convert_version_to_numeric(version)
        is_latest = VersionManager._does_version_match_latest_tag(version, latest_tag)
        
        return VersionInfo(
            original=version,
            major_minor=major_minor,
            numeric=numeric_version,
            is_master=False,
            is_latest=is_latest
        )
    
    @staticmethod
    def validate_version_format(version: str) -> bool:
        if not version:
            return False
        
        if version == VersionConstants.LATEST_VERSION_KEYWORD:
            return True
        
        if VersionManager._is_master_version(version):
            return True
        
        if RegexPatterns.VERSION.search(version):
            return True
        
        return False
    
    # Private helper methods
    
    @staticmethod
    def _handle_latest_version_keyword(latest_tag: str) -> VersionInfo:
        # Empty latest_tag is common in CI
        if not latest_tag:
            return VersionInfo(
                original=VersionConstants.LATEST_VERSION_KEYWORD,
                major_minor=VersionConstants.LATEST_VERSION_KEYWORD,
                numeric=None,
                is_master=False,
                is_latest=True
            )
        
        latest_match = RegexPatterns.VERSION.search(latest_tag)
        major_minor = latest_match.group(1) if latest_match else VersionConstants.LATEST_VERSION_KEYWORD
        numeric = VersionManager._convert_version_to_numeric(latest_tag) if latest_match else None
        
        return VersionInfo(
            original=VersionConstants.LATEST_VERSION_KEYWORD,
            major_minor=major_minor,
            numeric=numeric,
            is_master=False,
            is_latest=True
        )
    
    @staticmethod
    def _convert_version_to_numeric(version: str) -> Optional[int]:
        match = RegexPatterns.VERSION.search(version)
        if not match:
            return None
            
        try:
            major, minor = map(int, match.group(1).split('.'))
            return major * VersionConstants.VERSION_MULTIPLIER + minor
        except (ValueError, IndexError):
            return None
    
    @staticmethod
    def _does_version_match_latest_tag(version: str, latest_tag: str) -> bool:
        if not latest_tag:
            return False
        
        latest_match = RegexPatterns.VERSION.search(latest_tag)
        version_match = RegexPatterns.VERSION.search(version)
        
        if not latest_match or not version_match:
            return False
            
        return latest_match.group(1) == version_match.group(1)
    
    @staticmethod
    def _is_master_version(version: str) -> bool:
        return bool(RegexPatterns.MASTER.search(version))
    
    @staticmethod
    def _is_staging_version(version: str) -> bool:
        return bool(RegexPatterns.STAGING.search(version))
    
    @staticmethod
    def _is_saas_version(version: str) -> bool:
        return bool(RegexPatterns.SAAS.search(version))
    
    @staticmethod
    def _is_test_version(version: str) -> bool:
        return bool(RegexPatterns.TEST.search(version))
    
    @staticmethod
    def _is_semantic_version(version: str) -> bool:
        return bool(RegexPatterns.SEMVER.match(version))
    
    # Legacy aliases for backward compatibility
    
    @staticmethod
    def extract_major_minor(version: str) -> str:
        return VersionManager.extract_major_minor_from_version(version)
    
    @staticmethod
    def is_latest(version: str, latest_tag: str) -> bool:
        return VersionManager.is_latest_stable_version(version, latest_tag)
    
    @staticmethod
    def get_environment(env: str, version: str) -> Environment:
        return VersionManager.detect_environment_from_version_and_param(env, version)
    

    
    @staticmethod
    def parse_version_info(version: str, latest_tag: str) -> VersionInfo:
        return VersionManager.parse_version_into_components(version, latest_tag)

    @staticmethod
    def validate_version(version: str) -> bool:
        return VersionManager.validate_version_format(version) 