#!/usr/bin/env bash

# This script updates the branch_report.json file by running the string pattern scanner.

set -euo pipefail

cd "$(dirname "$0")"

poetry run python src/string_pattern_scanner.py ../../specs/runai-public.yaml -o branch_report.json
