"""
File processor module for OpenAPI spec build processing.

Handles file operations and content transformations for OpenAPI specs.
"""

import shutil
from pathlib import Path
from typing import Optional
from .config import Config, RegexPatterns, BrandingConstants
from .link_replacer import LinkReplacer


class FileProcessor:
    """Handles file operations and content transformations"""

    # Allowed file extensions to prevent operating on unexpected file types
    _ALLOWED_EXTENSIONS = {'.yaml', '.yml', '.json'}

    def __init__(self, spec_file_path: Path, spec_file_src: Path, config: Config,
                 trusted_root: Optional[Path] = None):
        root = trusted_root.resolve() if trusted_root is not None else None
        self.spec_file_path = self._validate_path(Path(spec_file_path), root)
        self.spec_file_src = self._validate_path(Path(spec_file_src), root)
        self.config = config
        self.link_replacer = LinkReplacer(config)
        self._cached_content: Optional[str] = None

    @classmethod
    def _validate_path(cls, path: Path, trusted_root: Optional[Path] = None) -> Path:
        """Resolve path to canonical form, verify extension, and enforce root boundary.

        Prevents path injection (S2083):
        - resolve() eliminates '..' traversal so the canonical path is used for all I/O.
        - Extension allowlist prevents operating on unexpected file types.
        - trusted_root, when provided, rejects any path that escapes the boundary.
        """
        resolved = path.resolve()
        if resolved.suffix.lower() not in cls._ALLOWED_EXTENSIONS:
            raise ValueError(
                f"Path {path!r} has disallowed extension {resolved.suffix!r}. "
                f"Allowed: {cls._ALLOWED_EXTENSIONS}"
            )
        if trusted_root is not None:
            try:
                resolved.relative_to(trusted_root)
            except ValueError:
                raise ValueError(
                    f"Path {path!r} resolves to {resolved!r} which is outside "
                    f"the trusted root {trusted_root!r}"
                )
        return resolved
    
    def read_spec_file(self) -> str:
        """Read spec file content with caching and proper error handling"""
        if self._cached_content is None:
            try:
                self._cached_content = self.spec_file_path.read_text(encoding=self.config.FILE_ENCODING)
            except (IOError, OSError) as e:
                raise RuntimeError(f"Failed to read spec file {self.spec_file_path}: {e}") from e
            except UnicodeDecodeError as e:
                raise RuntimeError(f"Failed to decode spec file {self.spec_file_path}: {e}") from e
        
        return self._cached_content
    
    def write_spec_file(self, content: str) -> None:
        """Write spec file content with proper error handling"""
        try:
            self.spec_file_path.write_text(content, encoding=self.config.FILE_ENCODING)
            self._cached_content = content  # Update cache
        except (IOError, OSError) as e:
            raise RuntimeError(f"Failed to write spec file {self.spec_file_path}: {e}") from e
    
    def reset_spec_file(self) -> None:
        """Reset spec file to original state"""
        try:
            shutil.copy2(self.spec_file_src, self.spec_file_path)
            self._cached_content = None  # Clear cache
        except Exception as e:
            raise RuntimeError(f"Failed to reset spec file: {e}") from e
    
    def apply_build_transformations(self, version: str, docs_base_url: str,
                                   version_info: Optional['VersionInfo'] = None) -> None:
        """Apply all content transformations to the spec file for local build"""
        content = self.read_spec_file()

        # Apply documentation link replacements
        content = self.link_replacer.replace_docs_links(content, docs_base_url)

        # Apply content transformations
        content = self._replace_version(content, version)
        content = self._replace_branding(content)

        self.write_spec_file(content)

    def _replace_version(self, content: str, version: str) -> str:
        """Replace version in the content"""
        return RegexPatterns.VERSION_REPLACE.sub(f'version: "{version}"', content)
    
    def _replace_branding(self, content: str) -> str:
        """Replace branding in the content"""
        return RegexPatterns.BRANDING.sub(BrandingConstants.NEW_BRAND, content)
    
    def validate_files(self) -> None:
        """Validate that required files exist"""
        if not self.spec_file_path.exists():
            raise FileNotFoundError(f"Spec file not found: {self.spec_file_path}")
        
        if not self.spec_file_src.exists():
            raise FileNotFoundError(f"Source spec file not found: {self.spec_file_src}")
    
 