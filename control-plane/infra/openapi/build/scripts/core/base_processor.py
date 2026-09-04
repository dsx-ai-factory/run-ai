"""
Base processor class for shared functionality of the local build processor.

Provides common initialization, validation, and error handling patterns.
"""

from abc import ABC, abstractmethod
from pathlib import Path
from .config import Config, ProcessingMode, Environment
from .version_manager import VersionManager
from .link_replacer import LinkReplacer
from .file_processor import FileProcessor
from .reporter import Reporter


class BaseProcessor(ABC):
    """Base class for spec processors with shared functionality"""

    def __init__(self, version: str, spec_file_path: str, latest_tag: str,
                 env: str, spec_file_src: str, mode: ProcessingMode):
        # Common parameters
        self.version = version
        self.spec_file_path = Path(spec_file_path)
        self.latest_tag = latest_tag
        self.env = env
        self.spec_file_src = Path(spec_file_src)

        # Common utilities
        self.config = Config.get_default()
        self.version_manager = VersionManager()
        self.link_replacer = LinkReplacer(self.config)
        self.file_processor = FileProcessor(self.spec_file_path, self.spec_file_src, self.config)
        self.reporter = Reporter(mode)

        # Parse version info
        self.version_info = self.version_manager.parse_version_into_components(
            version, latest_tag
        )

        # Validate inputs
        self._validate_common_inputs()
        self._validate_specific_inputs()
    
    def _validate_common_inputs(self) -> None:
        """Validate common inputs shared by all processors"""
        if not self.version:
            raise ValueError(
                "Version parameter is required. "
                "Provide a version like '2.22.0', 'latest', or '2.21.0.master'"
            )
        
        if not VersionManager.validate_version_format(self.version):
            raise ValueError(
                f"Invalid version format: '{self.version}'. "
                f"Expected formats: '2.22.0', 'latest', 'v2.21.0-saas.1', or '2.20.0.master'"
            )
        
        try:
            self.file_processor.validate_files()
        except FileNotFoundError as e:
            raise FileNotFoundError(
                f"Required file missing: {e}. "
                f"Ensure both spec file ({self.spec_file_path}) and source file ({self.spec_file_src}) exist."
            ) from e
    
    @abstractmethod
    def _validate_specific_inputs(self) -> None:
        """Validate processor-specific inputs (implemented by subclasses)"""
        ...
    
    def get_environment(self) -> Environment:
        """Get detected environment"""
        return self.version_manager.detect_environment_from_version_and_param(
            self.env, self.version
        )
    
    def get_docs_url(self, environment: Environment, mode: ProcessingMode) -> str:
        """Get documentation URL for given environment and mode"""
        return self.link_replacer.get_docs_url(self.version_info)

    @abstractmethod
    def run(self) -> bool:
        """Main execution method (implemented by subclasses)"""
        ... 