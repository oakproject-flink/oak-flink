#!/usr/bin/env bash
#
# Run integration tests against a specific Flink version
# Usage: bash test-single-version.sh [VERSION]
#   VERSION: 1.18, 1.19, 1.20, 2.0, 2.1 (default: 2.1)
#
# Requirements:
#   - Docker must be running
#   - Port 8081 must be available
#

set -e

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default to Flink 2.1 if no argument provided
VERSION="${1:-2.1}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Oak - Single Version Test${NC}"
echo -e "${BLUE}  Flink ${VERSION}${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Validate version and set compose file
case "$VERSION" in
    1.18)
        COMPOSE_FILE="docker-compose-1.18.yml"
        FULL_VERSION="1.18.1"
        ;;
    1.19)
        COMPOSE_FILE="docker-compose-1.19.yml"
        FULL_VERSION="1.19.1"
        ;;
    1.20)
        COMPOSE_FILE="docker-compose-1.20.yml"
        FULL_VERSION="1.20.0"
        ;;
    2.0)
        COMPOSE_FILE="docker-compose-2.0.yml"
        FULL_VERSION="2.0.0"
        ;;
    2.1)
        COMPOSE_FILE="docker-compose-2.1.yml"
        FULL_VERSION="2.1.0"
        ;;
    *)
        echo -e "${RED}[ERROR] Invalid version: ${VERSION}${NC}"
        echo ""
        echo "Valid versions: 1.18, 1.19, 1.20, 2.0, 2.1"
        echo "Usage: bash test-single-version.sh [VERSION]"
        exit 1
        ;;
esac

# Check Docker
echo -e "${YELLOW}Checking Docker...${NC}"
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}[ERROR] Docker is not running${NC}"
    echo "Please start Docker and try again"
    exit 1
fi
echo -e "${GREEN}[OK] Docker is running${NC}"
echo ""

# Get script directory and navigate to test directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR/../oak-lib/flink/rest-api"

# Start Flink cluster
echo "Starting Flink ${FULL_VERSION} cluster..."
if ! docker compose -f "testdata/${COMPOSE_FILE}" up -d; then
    echo -e "${RED}[ERROR] Failed to start Flink cluster${NC}"
    exit 1
fi

# Wait for Flink to be ready
echo "Waiting for Flink to start..."
sleep 10

# Run integration tests
echo ""
echo "Running integration tests against Flink ${FULL_VERSION}..."
echo ""

# Run tests and capture result
set +e
go test -tags=integration -v -timeout 10m -count=1
TEST_RESULT=$?
set -e

# Cleanup
echo ""
echo "Stopping Flink cluster..."
docker compose -f "testdata/${COMPOSE_FILE}" down -v >/dev/null 2>&1

# Exit with test result
if [ $TEST_RESULT -eq 0 ]; then
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}[SUCCESS] Flink ${FULL_VERSION} tests passed!${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}[FAILED] Flink ${FULL_VERSION} tests failed${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
