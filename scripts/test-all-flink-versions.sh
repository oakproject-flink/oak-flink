#!/bin/bash
set -e

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Oak Project - All Flink Versions Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Docker is running
echo -e "${YELLOW}Checking Docker...${NC}"
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running${NC}"
    echo "Please start Docker and try again"
    exit 1
fi
echo -e "${GREEN}✓ Docker is running${NC}"

echo ""
echo "This will test against ALL supported Flink versions:"
echo "  - Flink 1.18.1"
echo "  - Flink 1.19.1"
echo "  - Flink 1.20.0"
echo "  - Flink 2.0.1"
echo "  - Flink 2.1.0"
echo ""
echo -e "${YELLOW}WARNING: This will take 10-15 minutes to complete${NC}"
echo "Each version will start a Docker container, run tests, and clean up"
echo ""
read -p "Press Enter to continue or Ctrl+C to cancel..."

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Running Multi-Version Integration Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd oak-lib/flink/rest-api
go test -tags=integration_versions -v -timeout 30m -count=1

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ All Flink version tests passed!${NC}"
echo -e "${GREEN}========================================${NC}"
