# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Version validation for OpenAPI spec build processing.
"""

from .config import RegexPatterns, VersionConstants


class VersionManager:

    @staticmethod
    def validate_version_format(version: str) -> bool:
        if not version:
            return False

        if version == VersionConstants.LATEST_VERSION_KEYWORD:
            return True

        if RegexPatterns.MASTER.search(version):
            return True

        return bool(RegexPatterns.VERSION.search(version))
