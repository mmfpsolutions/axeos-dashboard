#!/bin/bash
#
# Docker run script for AxeOS Dashboard
# This script properly mounts the config directory and starts the container
#

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== AxeOS Dashboard - Docker Runner ===${NC}\n"

# Check if config directory exists
if [ ! -d "./config" ]; then
    echo -e "${RED}ERROR: ./config directory not found!${NC}"
    echo "Please create a config directory with the required files:"
    echo "  - config.json"
    echo "  - access.json"
    echo "  - jsonWebTokenKey.json"
    exit 1
fi

# Create data directory if it doesn't exist
if [ ! -d "./data" ]; then
    echo -e "${YELLOW}Creating ./data directory for metrics storage...${NC}"
    mkdir -p ./data
fi

# Check if required config files exist
REQUIRED_FILES=("config.json" "access.json" "jsonWebTokenKey.json")
MISSING_FILES=()

for file in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "./config/$file" ]; then
        MISSING_FILES+=("$file")
    fi
done

if [ ${#MISSING_FILES[@]} -gt 0 ]; then
    echo -e "${RED}ERROR: Missing required configuration files:${NC}"
    for file in "${MISSING_FILES[@]}"; do
        echo "  - $file"
    done
    exit 1
fi

echo -e "${GREEN}✓ Configuration files found${NC}"

# Check if Docker image exists
if ! docker images | grep -q "axeos-dashboard"; then
    echo -e "${YELLOW}Docker image 'axeos-dashboard:latest' not found.${NC}"
    echo -e "${YELLOW}Building image now...${NC}\n"
    docker build -t axeos-dashboard:latest .
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to build Docker image!${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Image built successfully${NC}\n"
fi

# Stop and remove existing container
if docker ps -a | grep -q axeos-dashboard; then
    echo -e "${YELLOW}Stopping and removing existing container...${NC}"
    docker stop axeos-dashboard 2>/dev/null
    docker rm axeos-dashboard 2>/dev/null
fi

# Get absolute paths to config and data directories
CONFIG_DIR="$(cd "$(dirname "$0")/config" && pwd)"
DATA_DIR="$(cd "$(dirname "$0")/data" && pwd)"

echo -e "${GREEN}✓ Config directory: $CONFIG_DIR${NC}"
echo -e "${GREEN}✓ Data directory: $DATA_DIR${NC}"

# Run the container
echo -e "\n${GREEN}Starting AxeOS Dashboard container...${NC}\n"

# Detect OS and use appropriate network mode
if [[ "$OSTYPE" == "darwin"* ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    # macOS or Windows - use port mapping (--network host doesn't work)
    echo -e "${YELLOW}Detected macOS/Windows - using port mapping${NC}"
    docker run -d \
      --name axeos-dashboard \
      -p 3000:3000 \
      -v "$CONFIG_DIR:/app/config" \
      -v "$DATA_DIR:/app/data" \
      axeos-dashboard:latest
else
    # Linux - use host network for better local device access
    echo -e "${YELLOW}Detected Linux - using host network${NC}"
    docker run -d \
      --name axeos-dashboard \
      --network host \
      -v "$CONFIG_DIR:/app/config" \
      -v "$DATA_DIR:/app/data" \
      axeos-dashboard:latest
fi

# Check if container started successfully
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}✓ Container started successfully!${NC}\n"
    echo "Container name: axeos-dashboard"
    echo "Dashboard URL:  http://localhost:3000"
    echo ""
    echo "View logs:      docker logs axeos-dashboard"
    echo "Follow logs:    docker logs -f axeos-dashboard"
    echo "Stop container: docker stop axeos-dashboard"
    echo ""

    # Wait a moment and show initial logs
    sleep 2
    echo -e "${YELLOW}=== Initial Logs ===${NC}"
    docker logs axeos-dashboard
else
    echo -e "\n${RED}✗ Failed to start container!${NC}"
    exit 1
fi
