"""
Link replacement module for OpenAPI spec build processing.

Handles documentation URL building and link replacement in OpenAPI specs.
"""

import re
from .config import Config
from .version_manager import VersionInfo


class LinkReplacer:
    """Builds documentation URLs and replaces links in content"""

    def __init__(self, config: Config):
        self.config = config

    def get_docs_url(self, version_info: VersionInfo) -> str:
        """Get documentation URL for the local build"""
        return self._get_build_mode_url(version_info)

    def replace_docs_links(self, content: str, docs_base_url: str) -> str:
        """Replace documentation links in the content"""
        print(f"Replacing documentation links with base URL: {docs_base_url}")

        # Create escaped pattern for regex
        from .config import RegexPatterns
        source_pattern = RegexPatterns.get_docs_pattern(self.config.NVIDIA_DOCS_DOMAIN)

        # Apply replacements based on target URL
        if "/saas" in docs_base_url:
            saas_replacement = f"{self.config.NVIDIA_DOCS_DOMAIN}/saas/"
            content = re.sub(source_pattern, saas_replacement, content)
        elif "/self-hosted" in docs_base_url:
            content = re.sub(source_pattern, f"{docs_base_url}/", content)
        elif "docs.run.ai" in docs_base_url:
            print("Skipping automatic replacement for legacy docs - will be handled manually")

        return content

    def _get_build_mode_url(self, version_info: VersionInfo) -> str:
        """Get documentation URL for build mode (always self-hosted)"""
        if not version_info.numeric:
            return f"{self.config.NVIDIA_DOCS_DOMAIN}/self-hosted"
        
        if version_info.numeric >= self.config.MIN_NEW_DOCS_VERSION:
            if version_info.is_latest:
                return f"{self.config.NVIDIA_DOCS_DOMAIN}/self-hosted"
            else:
                return f"{self.config.NVIDIA_DOCS_DOMAIN}/self-hosted/{version_info.major_minor}"
        else:
            return f"{self.config.LEGACY_DOCS_DOMAIN}/v{version_info.major_minor}"
    
 