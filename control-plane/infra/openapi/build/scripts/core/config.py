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
    LATEST_VERSION_KEYWORD = "latest"
    MASTER_VERSION_SUFFIX = ".master"
    MASTER_VERSION_KEYWORD = "master"
    VERSION_MULTIPLIER = 100  # For converting "2.20" to 220


@dataclass
class Config:
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
    LOGO = re.compile(r'url: \./runai-logo-api\.png')
    
