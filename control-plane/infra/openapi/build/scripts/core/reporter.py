# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Shared reporting and output formatting for API documentation build processing.

Provides consistent, reusable printing functions for the build operation.
"""

import sys
from datetime import datetime
from pathlib import Path
from .config import Colors, ProcessingMode


class Reporter:
    """Handles all output formatting and reporting for API documentation operations"""

    def __init__(self, mode: ProcessingMode):
        self.mode = mode
        self.start_time = datetime.now()

    def print_header(self, version: str, spec_file_path: Path,
                    spec_file_src: Path) -> None:
        """Print operation header with all key information"""
        mode_info = {
            ProcessingMode.BUILD: ("🔨", "Build Mode: Processing OpenAPI spec for local development"),
        }

        icon, title = mode_info[self.mode]

        print(f"{Colors.BLUE}{icon} {title}{Colors.NC}")
        print("=" * 80)
        print("📋 Configuration:")
        print(f"   • Version: {version}")
        print(f"   • Target spec: {spec_file_path}")
        print(f"   • Source spec: {spec_file_src}")
        
        # File size info
        if spec_file_src.exists():
            size_kb = spec_file_src.stat().st_size / 1024
            print(f"   • Source size: {size_kb:.1f} KB")
        
        print()
    
    def print_transformation_start(self) -> None:
        """Print transformation start message"""
        print(f"{Colors.YELLOW}⚙️  Applying transformations for local development...{Colors.NC}")
    
    def print_transformations_applied(self, version_transform: str) -> None:
        """Print detailed transformation results"""
        print("✨ Transformations applied:")
        print(f"   • Version: 2.0.0 → {version_transform}")
        print()
    
    def print_build_success(self, version: str, output_file: Path) -> None:
        """Print build success summary with enhanced details"""
        elapsed = self._get_elapsed_time()
        
        print(f"{Colors.GREEN}✅ Build completed successfully! ({elapsed}){Colors.NC}")
        print()
        
        # Enhanced transformation summary
        self.print_transformations_applied(version)

    def print_error(self, error: Exception, context: str = "") -> None:
        """Print error with context and troubleshooting"""
        print(f"{Colors.RED}❌ {context}Error: {error}{Colors.NC}", file=sys.stderr)
        print("", file=sys.stderr)
        self._print_build_troubleshooting()

    def _get_elapsed_time(self) -> str:
        """Get formatted elapsed time"""
        elapsed = datetime.now() - self.start_time
        seconds = elapsed.total_seconds()
        if seconds < 60:
            return f"{seconds:.1f}s"
        else:
            minutes = int(seconds // 60)
            remaining_seconds = seconds % 60
            return f"{minutes}m {remaining_seconds:.1f}s"
    
    def _print_build_troubleshooting(self) -> None:
        """Print build-specific troubleshooting tips"""
        print("🔧 Build Troubleshooting:", file=sys.stderr)
        print("   • Verify source and target files exist and are readable", file=sys.stderr)
        print("   • Check OpenAPI spec is valid YAML format", file=sys.stderr)
        print("   • Ensure sufficient disk space for file operations", file=sys.stderr)
        print("   • Validate version format (e.g., '2.22.0', 'latest')", file=sys.stderr)
        print("   • Check file permissions in _build/ directory", file=sys.stderr)