#!/usr/bin/env bash
#
# Run all tests (unit + integration) for Oak project
# Works on Windows (Git Bash), Linux, and macOS
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

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Oak Project - All Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Get the project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Track overall status
UNIT_PASSED=0
INTEGRATION_PASSED=0

# Run unit tests
echo -e "${CYAN}[1/2] Running unit tests...${NC}"
echo ""
if bash "$SCRIPT_DIR/test-unit.sh"; then
    UNIT_PASSED=1
    echo ""
else
    echo ""
    echo -e "${RED}Unit tests failed. Stopping.${NC}"
    exit 1
fi

# Run integration tests
echo -e "${CYAN}[2/2] Running integration tests...${NC}"
echo ""
if bash "$SCRIPT_DIR/test-integration.sh"; then
    INTEGRATION_PASSED=1
    echo ""
else
    echo ""
    echo -e "${RED}Integration tests failed.${NC}"
    exit 1
fi

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
if [ $UNIT_PASSED -eq 1 ] && [ $INTEGRATION_PASSED -eq 1 ]; then
    echo -e "${GREEN}✓ Unit tests: PASSED${NC}"
    echo -e "${GREEN}✓ Integration tests: PASSED${NC}"
    echo ""
    echo -e "${GREEN}All tests passed! 🎉${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
