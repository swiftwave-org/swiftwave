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

# If .images folder does not exist, create it
if [ ! -d "./.images" ]; then
    mkdir ./.images
fi

# Fetch required images
echo -e "${GREEN}Fetching required images...${NC}"
sudo docker pull ghcr.io/swiftwave-org/swiftwave-dashboard:latest
sudo docker pull haproxytech/haproxy-debian:2.9
sudo docker save -o ./.images/swiftwave-dashboard.tar ghcr.io/swiftwave-org/swiftwave-dashboard:latest
sudo docker save -o ./.images/haproxy-debian.tar haproxytech/haproxy-debian:2.9

# build docker image
echo "${GREEN}Building docker image...${NC}"
sudo docker build -t swiftwave_dev_env:latest -f $dockerfile_path .

# Remove existing container
sudo docker rm -f swiftwave_dev_env --force > /dev/null 2>&1

echo "Starting swiftwave dev environment..."
# run dev/start.sh script
container_id=$(sudo docker run --privileged -d -v go_cache:/home/user/go/pkg/mod/ -v .:/app:rw -w /app --name swiftwave_dev_env swiftwave_dev_env:latest)

echo "Perform post installation tasks..."
# Run startup script by exec
sudo docker exec swiftwave_dev_env bash -c "cd /app && ./dev/start.sh" | tee -a dev.log
echo "Post installation tasks completed"
# Print IP
ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $container_id)
echo -e "\n\n${GREEN}Swiftwave dev environment is started successfully${NC}"
echo -e "Follow the steps below to start the server"
echo -e "---------------------------------------------------------"
echo -e "${GREEN}Open Bash Terminal${NC} : ${BLUE}http://$ip:7681${NC}"
echo -e "---------------------------------------------------------"
echo -e "${GREEN}Start Swiftwave Server :${NC}"
echo -e "  1. Open ${BLUE}http://$ip:7681${NC} in browser"
echo -e "  2. Run ${BLUE}go run .${NC} in terminal"
echo -e "  3. Wait for 3~4 minutes to start the server"
echo -e "---------------------------------------------------------"
echo -e "${GREEN}Access Swiftwave Dashboard :${NC}"
echo -e "  1. Open ${BLUE}http://$ip:1212${NC} in browser"
echo -e "  2. Configure server details from bottom bar -> ${BLUE}Host ${ip}${NC} | ${BLUE}Port 3333${NC}"
echo -e "  2. Login with ${BLUE}admin${NC} as username and ${BLUE}admin${NC} as password"
echo -e "---------------------------------------------------------"
echo -e "${GREEN}Stop Development Container :${NC}"
echo -e "${YELLOW}sudo docker stop swiftwave_dev_env${NC}"
echo -e "---------------------------------------------------------"