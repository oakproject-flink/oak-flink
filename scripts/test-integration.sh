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
echo ""

# Track test failures
FAILED=0

# Run Flink integration tests (requires Docker)
echo -e "${YELLOW}[*] Flink REST API integration tests...${NC}"
echo -e "${YELLOW}    (This will start a Flink cluster in Docker)${NC}"
echo ""
if ! go test -tags=integration -v -timeout 10m -count=1 ./oak-lib/flink/rest-api/...; then
    FAILED=1
    echo ""
    echo -e "${RED}✗ Flink integration tests failed!${NC}"
else
    echo ""
    echo -e "${GREEN}✓ Flink integration tests passed!${NC}"
fi

echo ""

# Run Oak Server integration tests (no Docker needed)
echo -e "${YELLOW}[*] Oak Server gRPC integration tests...${NC}"
echo -e "${YELLOW}    (Uses in-memory connections, no Docker needed)${NC}"
echo ""
if ! go test -tags=integration -v -timeout 5m -count=1 ./oak-server/internal/grpc/...; then
    FAILED=1
    echo ""
    echo -e "${RED}✗ Server integration tests failed!${NC}"
else
    echo ""
    echo -e "${GREEN}✓ Server integration tests passed!${NC}"
fi

echo ""

# Final result
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All integration tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some integration tests failed!${NC}"
    exit 1
fi
