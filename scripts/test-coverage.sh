#!/usr/bin/env bash
#
# Run tests with coverage report for Oak project
# Works on Windows (Git Bash), Linux, and macOS
#
# Generates HTML coverage report and displays coverage percentages
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
echo -e "${BLUE}  Oak Project - Coverage Report${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Get the project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Create coverage directory
mkdir -p coverage

echo -e "${YELLOW}Running tests with coverage (unit tests only)...${NC}"
echo ""

# Run unit tests with coverage for oak-lib
if go test -short -coverprofile=coverage/coverage.out -covermode=atomic ./oak-lib/...; then
    echo ""
    echo -e "${GREEN}✓ Tests completed successfully${NC}"
    echo ""

    # Display coverage summary
    echo -e "${CYAN}Coverage Summary:${NC}"
    go tool cover -func=coverage/coverage.out | tail -1
    echo ""

    # Generate HTML report
    echo -e "${YELLOW}Generating HTML coverage report...${NC}"
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html

    echo -e "${GREEN}✓ Coverage report generated: coverage/coverage.html${NC}"
    echo ""
    echo -e "${CYAN}To view the report, open:${NC}"
    echo -e "  ${BLUE}coverage/coverage.html${NC}"

    exit 0
else
    echo ""
    echo -e "${RED}✗ Tests failed!${NC}"
    exit 1
fi
