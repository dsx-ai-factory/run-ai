#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


# This script updates the branch_report.json file by running the string pattern scanner.

set -euo pipefail

cd "$(dirname "$0")"

poetry run python src/string_pattern_scanner.py ../../specs/runai-public.yaml -o branch_report.json
