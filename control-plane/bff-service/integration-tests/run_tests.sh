#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

VERBOSE=false
SHOW_ONLY_FAILED=false 
RUN_SPECIFIC_TEST=""
TIMEOUT="30m"
TEST_FILE="./integ-tests"
TEST_BINARY="./integ-tests"
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -f, --failed-only       Show only failed tests in output"
    echo "  -t, --test TEST_NAME    Run specific test (partial name matching)"
    echo "  --timeout DURATION      Set test timeout (default: 30m)"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                      # Run all tests with enhanced output"
    echo "  $0 -t TestHealth        # Run only tests containing 'TestHealth'"
    echo "  $0 -f                   # Show only failed tests"
}
while [[ $# -gt 0 ]]; do
    case $1 in
        -l | --local)
            TEST_FILE="./tests"
            LOCAL=true
            shift
            ;;
        -f|--failed-only)
            SHOW_ONLY_FAILED=true
            shift
            ;;
        -t|--test)
            RUN_SPECIFIC_TEST="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

echo -e "${BLUE}🧪 Integration Test Runner${NC}"
echo -e "${BLUE}==========================${NC}"

if [ ! -e "$TEST_FILE" ]; then
    echo -e "${RED}❌ Error: $TEST_FILE not found. Please build it first.${NC}"
    exit 1
fi

if [ "$LOCAL" = true ]; then
TEST_CMD="go test $TEST_FILE -v -parallel=8"
else
    TEST_CMD="$TEST_BINARY -test.v -test.parallel=8"
fi
if [ "$LOCAL" = true ]; then
    TEST_CMD="$TEST_CMD -timeout=$TIMEOUT"
else
    TEST_CMD="$TEST_CMD -test.timeout=$TIMEOUT"
fi

if [ -n "$RUN_SPECIFIC_TEST" ]; then
    TEST_CMD="$TEST_CMD -test.run=$RUN_SPECIFIC_TEST"
    echo -e "${YELLOW}🎯 Running tests matching: $RUN_SPECIFIC_TEST${NC}"
fi

echo -e "${BLUE}📋 Command: $TEST_CMD${NC}"
echo ""

TEMP_OUTPUT=$(mktemp)
TEMP_ERROR=$(mktemp)

cleanup() {
    rm -f "$TEMP_OUTPUT" "$TEMP_ERROR"
}
trap cleanup EXIT
echo -e "${BLUE}🚀 Starting tests...${NC}"
echo ""

if [ "$SHOW_ONLY_FAILED" = true ]; then
    $TEST_CMD 2>&1 | tee "$TEMP_OUTPUT" | grep -E "(FAIL|Error|panic|fatal)" || true
else
    $TEST_CMD 2>&1 | tee "$TEMP_OUTPUT"
fi

TEST_EXIT_CODE=${PIPESTATUS[0]}

echo ""
echo -e "${BLUE}📊 Test execution completed with exit code: $TEST_EXIT_CODE${NC}"

if [ -f "$TEMP_OUTPUT" ]; then
    FAILED_TESTS=$(grep -c "FAIL:" "$TEMP_OUTPUT" 2>/dev/null || echo "0")
    PASSED_TESTS=$(grep -c "PASS:" "$TEMP_OUTPUT" 2>/dev/null || echo "0")
    SKIPPED_TESTS=$(grep -c "SKIP:" "$TEMP_OUTPUT" 2>/dev/null || echo "0")
    
    echo ""
    echo -e "${BLUE}📈 Quick Summary:${NC}"
    echo -e "  ✅ Passed: $PASSED_TESTS"
    echo -e "  ❌ Failed: $FAILED_TESTS"
    echo -e "  ⏭️  Skipped: $SKIPPED_TESTS"

    if [ "$FAILED_TESTS" -gt 0 ]; then
        echo ""
        echo -e "${RED}❌ Failed Tests:${NC}"
        grep "FAIL:" "$TEMP_OUTPUT" | sed 's/^/  /' || true
    fi
    if [ "$SKIPPED_TESTS" -gt 0 ]; then
        echo ""
        echo -e "${YELLOW}⏭️  Skipped Tests:${NC}"
        grep "SKIP:" "$TEMP_OUTPUT" | sed 's/^/  /' || true
    fi

    TOTAL_TESTS=$((PASSED_TESTS + FAILED_TESTS + SKIPPED_TESTS))
    if [ "$TOTAL_TESTS" -eq 0 ] && [ "$TEST_EXIT_CODE" -eq 0 ]; then
        echo ""
        echo -e "${RED}❌ Error: no tests were collected. Check -t filter or build tags.${NC}"
        exit 1
    fi
fi

exit $TEST_EXIT_CODE
