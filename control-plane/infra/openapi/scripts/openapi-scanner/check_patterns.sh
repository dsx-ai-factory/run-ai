#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BRANCH_REPORT="${SCRIPT_DIR}/branch_report.json"
CURRENT_REPORT="${SCRIPT_DIR}/current_report.json"
COMPARISON_REPORT="${SCRIPT_DIR}/comparison.json"
OPENAPI_SPEC="${SCRIPT_DIR}/../../specs/runai-public.yaml"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}OpenAPI Pattern Change Checker${NC}"
echo -e "${BLUE}========================================${NC}"

# Step 1: Check that branch_report.json exists
echo -e "\n${YELLOW}Step 1: Checking for branch_report.json...${NC}"
if [ ! -f "$BRANCH_REPORT" ]; then
    echo -e "${RED}Error: branch_report.json not found at $BRANCH_REPORT${NC}"
    echo -e "${RED}Please generate a branch report first using:${NC}"
    echo -e "${RED}  poetry run string-pattern-scanner $OPENAPI_SPEC -o branch_report.json${NC}"
    exit 1
fi
echo -e "${GREEN}✓ branch_report.json found${NC}"

# Step 2: Generate current report
echo -e "\n${YELLOW}Step 2: Generating current report from OpenAPI spec...${NC}"
if [ ! -f "$OPENAPI_SPEC" ]; then
    echo -e "${RED}Error: OpenAPI spec not found at $OPENAPI_SPEC${NC}"
    exit 1
fi

poetry run string-pattern-scanner "$OPENAPI_SPEC" -o "$CURRENT_REPORT"

if [ ! -f "$CURRENT_REPORT" ]; then
    echo -e "${RED}Error: Failed to generate current_report.json${NC}"
    exit 1
fi
echo -e "${GREEN}✓ current_report.json generated${NC}"

# Step 3: Create comparison report
echo -e "\n${YELLOW}Step 3: Creating comparison report...${NC}"
poetry run report-comparator "$BRANCH_REPORT" "$CURRENT_REPORT" --json-output "$COMPARISON_REPORT" --quiet

if [ ! -f "$COMPARISON_REPORT" ]; then
    echo -e "${RED}Error: Failed to generate comparison.json${NC}"
    exit 1
fi
echo -e "${GREEN}✓ comparison.json generated${NC}"

# Step 4: Check for modified_patterns or added_patterns
echo -e "\n${YELLOW}Step 4: Checking for breaking changes...${NC}"

# Use jq to check for added or modified patterns (if jq is available)
if command -v jq &> /dev/null; then
    ADDED_COUNT=$(jq '.statistics.added_count // 0' "$COMPARISON_REPORT")
    MODIFIED_COUNT=$(jq '.statistics.modified_count // 0' "$COMPARISON_REPORT")
    REMOVED_COUNT=$(jq '.statistics.removed_count // 0' "$COMPARISON_REPORT")
    
    echo -e "${BLUE}Comparison Results:${NC}"
    echo -e "  Added patterns: ${ADDED_COUNT}"
    echo -e "  Modified patterns: ${MODIFIED_COUNT}"
    echo -e "  Removed patterns: ${REMOVED_COUNT}"
    
    if [ "$ADDED_COUNT" -gt 0 ] || [ "$MODIFIED_COUNT" -gt 0 ]; then
        echo -e "\n${RED}⚠️  WARNING: Pattern changes detected!${NC}"
        echo -e "${RED}Added or modified patterns may indicate breaking changes.${NC}"
        echo -e "${YELLOW}Please review the comparison report: $COMPARISON_REPORT${NC}"
        
        # Display added patterns if any
        if [ "$ADDED_COUNT" -gt 0 ]; then
            echo -e "\n${YELLOW}Added Patterns:${NC}"
            jq -r '.added_patterns[] | "  • \(.schema_name): \(.pattern)"' "$COMPARISON_REPORT"
        fi
        
        # Display modified patterns if any
        if [ "$MODIFIED_COUNT" -gt 0 ]; then
            echo -e "\n${YELLOW}Modified Patterns:${NC}"
            jq -r '.modified_patterns[] | "  • \(.schema_name)"' "$COMPARISON_REPORT"
        fi
        
        exit 1
    else
        echo -e "\n${GREEN}✓ No breaking pattern changes detected${NC}"
        exit 0
    fi
else
    # Fallback: Use grep to check for changes (less precise)
    echo -e "${YELLOW}Note: jq not found, using basic pattern detection${NC}"
    if grep -q '"added_patterns": \[\]' "$COMPARISON_REPORT" && grep -q '"modified_patterns": \[\]' "$COMPARISON_REPORT"; then
        echo -e "\n${GREEN}✓ No breaking pattern changes detected${NC}"
        exit 0
    else
        echo -e "\n${RED}⚠️  WARNING: Pattern changes detected!${NC}"
        echo -e "${RED}Please review the comparison report: $COMPARISON_REPORT${NC}"
        echo -e "${YELLOW}Install jq for detailed output: brew install jq${NC}"
        exit 1
    fi
fi

