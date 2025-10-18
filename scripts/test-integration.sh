#!/usr/bin/env bash
#
# Run integration tests for Oak project
# Works on Windows (Git Bash), Linux, and macOS
#
# Requirements:
#   - Docker must be running
#   - Port 8081 must be available (for Flink)
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Oak Project - Integration Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Get the project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Check if Docker is running
echo -e "${YELLOW}Checking Docker...${NC}"
if ! docker ps >/dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running!${NC}"
    echo -e "${YELLOW}Please start Docker and try again.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker is running${NC}"
echo ""

echo -e "${YELLOW}Running integration tests...${NC}"
echo -e "${YELLOW}(This will start a Flink cluster in Docker)${NC}"
echo ""

# Run integration tests with tags (-count=1 disables test caching)
if go test -tags=integration -v -timeout 10m -count=1 ./oak-lib/flink/rest-api/...; then
    echo ""
    echo -e "${GREEN}✓ All integration tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Integration tests failed!${NC}"
    exit 1
fi
