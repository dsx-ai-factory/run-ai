# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

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


class FileConstants:
    DEFAULT_ENCODING = 'utf-8'
    SPEC_FILE_BACKUP_SUFFIX = '.backup'
    

class VersionConstants:
    LATEST_VERSION_KEYWORD = "latest"


@dataclass
class Config:
    FILE_ENCODING: str = FileConstants.DEFAULT_ENCODING
    
    @classmethod
    def get_default(cls) -> 'Config':
        return cls()


class RegexPatterns:
    # Version parsing
    VERSION = re.compile(r'v?(\d+\.\d+)')
    MASTER = re.compile(r'\.master')

    # Content replacement
    VERSION_REPLACE = re.compile(r'version: 2\.0\.0')
    
