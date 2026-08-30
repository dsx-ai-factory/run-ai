#!/usr/bin/env python3
"""
Build-only main script for OpenAPI spec build processing.
"""

import sys
import argparse
from core import (
    ProcessingMode,
    Reporter, BaseProcessor
)


class BuildProcessor(BaseProcessor):
    
    def __init__(self, version: str, spec_file_path: str, latest_tag: str, 
                 env: str, spec_file_src: str):
        super().__init__(version, spec_file_path, latest_tag, env, spec_file_src, ProcessingMode.BUILD)
        # Store parsed version info for compatibility (some methods may still reference it)
        self.parsed_version_info = self.version_info
    
    def _validate_specific_inputs(self) -> None:
        """Build processor has no additional validation beyond common inputs"""
        pass
    
    def run(self) -> bool:
        try:
            # Print header with all configuration details
            self.reporter.print_header(
                version=self.version,
                spec_file_path=self.spec_file_path,
                latest_tag=self.latest_tag,
                env=self.env,
                spec_file_src=self.spec_file_src
            )
            
            # Use base class helper methods
            detected_environment = self.get_environment()
            documentation_base_url = self.get_docs_url(detected_environment, ProcessingMode.BUILD)
            
            # Print environment detection results
            self.reporter.print_environment_detection(detected_environment, documentation_base_url)
            
            # Print transformation start
            self.reporter.print_transformation_start()
            
            self.file_processor.apply_build_transformations(
                version=self.version,
                docs_base_url=documentation_base_url,
                version_info=self.version_info
            )
            
            # Print success summary using Reporter method
            self.reporter.print_build_success(
                version=self.version,
                docs_url=documentation_base_url,
                output_file=self.spec_file_path
            )
            return True
            
        except Exception as transformation_error:
            self.reporter.print_error(transformation_error, "Build ")
            return False
    



def create_argument_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser()
    
    parser.add_argument('version')
    parser.add_argument('spec_file_path')
    parser.add_argument('latest_tag')
    parser.add_argument('env')
    parser.add_argument('spec_file_src')
    
    return parser


def main() -> None:
    argument_parser = create_argument_parser()
    parsed_args = argument_parser.parse_args()
    
    try:
        build_processor = BuildProcessor(
            version=parsed_args.version,
            spec_file_path=parsed_args.spec_file_path,
            latest_tag=parsed_args.latest_tag,
            env=parsed_args.env,
            spec_file_src=parsed_args.spec_file_src
        )
        
        processing_succeeded = build_processor.run()
        sys.exit(0 if processing_succeeded else 1)
        
    except (ValueError, FileNotFoundError) as user_input_error:
        # Create a reporter for error handling 
        error_reporter = Reporter(ProcessingMode.BUILD)
        error_reporter.print_error(user_input_error, "Input ")
        print("Use --help for usage information", file=sys.stderr)
        sys.exit(1)
        
    except Exception as unexpected_error:
        # Create a reporter for error handling
        error_reporter = Reporter(ProcessingMode.BUILD)
        error_reporter.print_error(unexpected_error, "Unexpected ")
        sys.exit(1)


if __name__ == "__main__":
    main() 