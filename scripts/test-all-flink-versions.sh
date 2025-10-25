#!/bin/bash
#
# Run integration tests against all Flink versions (1.18-2.1)
# Usage: bash test-all-flink-versions.sh [-y|--yes]
#   -y, --yes: Skip confirmation prompt (for CI/automation)
#
# Environment variables:
#   CI=true or CONTINUOUS_INTEGRATION=true: Auto-skip prompt
#
# Requirements:
#   - Docker must be running
#   - Port 8081 must be available
#   - 10-15 minutes to complete
#

set -e

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check for -y or --yes flag, or CI environment
SKIP_PROMPT=0
if [ "$1" = "-y" ] || [ "$1" = "--yes" ]; then
    SKIP_PROMPT=1
fi

# Auto-detect CI environments
if [ "$CI" = "true" ] || [ "$CONTINUOUS_INTEGRATION" = "true" ] || [ "$GITHUB_ACTIONS" = "true" ] || [ -n "$JENKINS_URL" ] || [ -n "$GITLAB_CI" ]; then
    SKIP_PROMPT=1
fi

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
echo "  - Flink 2.0.0"
echo "  - Flink 2.1.0"
echo ""
echo -e "${YELLOW}WARNING: This will take 10-15 minutes to complete${NC}"
echo "Each version will start a Docker container, run tests, and clean up"
echo ""

# Skip prompt in CI or with -y flag
if [ $SKIP_PROMPT -eq 1 ]; then
    echo -e "${GREEN}[CI MODE] Skipping confirmation prompt${NC}"
    echo ""
else
    read -p "Press Enter to continue or Ctrl+C to cancel..."
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Running Multi-Version Integration Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Get script directory and navigate to test directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR/../oak-lib/flink/rest-api"

go test -tags=integration_versions -v -timeout 30m -count=1

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ All Flink version tests passed!${NC}"
echo -e "${GREEN}========================================${NC}"
