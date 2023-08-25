#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# verify if script runninf as non-root user
if [[ "$EUID" -eq 0 ]]; then
    echo -e "${RED}Please do not run as root${NC}"
    exit 1
fi

# verify if docker is available
if ! [ -x "$(command -v docker)" ]; then
    echo -e "${RED}Docker is not installed. Please install docker first${NC}"
    exit 1
fi

# Get base path
base_path=$(pwd)
dockerfile_path="$base_path/dev/Dockerfile"

# Remove ./data folder
sudo rm -rf ./data

# build docker image
sudo docker build -t swiftwave_dev_env:latest -f $dockerfile_path .

# Remove existing container
sudo docker rm -f swiftwave_dev_env --force > /dev/null 2>&1

echo "Starting swiftwave dev environment..."
# run dev/start.sh script
container_id=$(sudo docker run --privileged -d -v go_cache:/home/user/go/pkg/mod/ -v .:/app:rw -w /app --name swiftwave_dev_env swiftwave_dev_env:latest)

echo "Perform post installation tasks..."
# Run startup script by exec
sudo docker exec swiftwave_dev_env bash -c "cd /app && ./dev/start.sh"  >> "./dev.log" 2>&1
echo "Post installation tasks completed"
# Print IP
ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $container_id)
echo -e "${GREEN}Swiftwave dev environment is running at $ip${NC}"
echo -e "---------------------------------------------------------"
echo -e "${GREEN}Access Bash Terminal${NC} : ${BLUE}http://$ip:7681${NC}"
echo -e "---------------------------------------------------------"
echo -e "Start Swiftwave :"
echo -e "  1. Open ${BLUE}http://$ip:7681${NC} in browser"
echo -e "  2. Run ${BLUE}go run .${NC} in terminal"
echo -e "---------------------------------------------------------"

# Stop the container
# echo "Stopping the container..."
# sudo docker stop swiftwave_dev_env > /dev/null 2>&1
