#!/usr/bin/env bash
#
# Run all tests for Oak project
# Usage: bash test-all.sh [--full-compat]
#   --full-compat: Include multi-version compatibility tests (10-15 min)
#
# Requirements for integration tests:
#   - Docker must be running
#   - Port 8081 must be available (for Flink)
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check for --full-compat flag
FULL_COMPAT=0
if [ "$1" = "--full-compat" ]; then
    FULL_COMPAT=1
fi

if [ $FULL_COMPAT -eq 1 ]; then
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  Oak Project - FULL COMPATIBILITY TEST${NC}"
    echo -e "${BLUE}  (Unit + Integration + Multi-Version)${NC}"
    echo -e "${BLUE}========================================${NC}"
else
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  Oak Project - All Tests${NC}"
    echo -e "${BLUE}  (Unit + Integration)${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${YELLOW}TIP: Use --full-compat to test all Flink versions (1.18-2.1)${NC}"
fi
echo ""

# Get the project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Track overall status
UNIT_PASSED=0
INTEGRATION_PASSED=0
MULTIVERSION_PASSED=0

# Run unit tests
echo -e "${CYAN}[1/3] Running unit tests...${NC}"
echo ""
if bash "$SCRIPT_DIR/test-unit.sh"; then
    UNIT_PASSED=1
    echo ""
else
    echo ""
    echo -e "${RED}Unit tests failed. Stopping.${NC}"
    exit 1
fi

# Run integration tests (single version - Flink 2.1)
echo -e "${CYAN}[2/3] Running integration tests (Flink 2.1)...${NC}"
echo ""
if bash "$SCRIPT_DIR/test-integration.sh"; then
    INTEGRATION_PASSED=1
    echo ""
else
    echo ""
    echo -e "${RED}Integration tests failed.${NC}"
    exit 1
fi

# Optionally run multi-version compatibility tests
if [ $FULL_COMPAT -eq 1 ]; then
    echo -e "${CYAN}[3/3] Running multi-version compatibility tests...${NC}"
    echo "This will test against ALL Flink versions 1.18-2.1 (~10-15 min)"
    echo ""

    # Disable exit on error temporarily to handle test failure gracefully
    set +e
    bash "$SCRIPT_DIR/test-all-flink-versions.sh"
    MULTIVERSION_RESULT=$?
    set -e

    if [ $MULTIVERSION_RESULT -eq 0 ]; then
        MULTIVERSION_PASSED=1
        echo ""
    else
        echo ""
        echo -e "${RED}Multi-version compatibility tests failed.${NC}"
        exit 1
    fi

    # Full compat summary
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  FULL COMPATIBILITY TEST SUMMARY${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "${GREEN}✓ Unit tests: PASSED${NC}"
    echo -e "${GREEN}✓ Integration tests: PASSED${NC}"
    echo -e "${GREEN}✓ Multi-version tests: PASSED${NC}"
    echo ""
    echo -e "${GREEN}All compatibility tests passed! 🎉${NC}"
    exit 0
else
    # Standard summary
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  Test Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "${GREEN}✓ Unit tests: PASSED${NC}"
    echo -e "${GREEN}✓ Integration tests: PASSED${NC}"
    echo ""
    echo -e "${GREEN}All tests passed! 🎉${NC}"
    echo ""
    echo -e "${YELLOW}Run with --full-compat to test all Flink versions (1.18-2.1)${NC}"
    exit 0
fi
