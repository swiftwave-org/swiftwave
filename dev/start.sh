#! /usr/bin/env bash


# Color
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Variables
IP=$(hostname -i | awk '{print $1}')
STACK_NAME="swiftwave"
SWARM_NETWORK="swarm_network"
SWIFTWAVE_DATA_FOLDER="/app/.data"
SWIFTWAVE_APP_TARBALL_FOLDER="/app/.data/tarball"
SWIFTWAVE_REDIS_FOLDER="/app/.data/redis"
SWIFTWAVE_HAPROXY_FOLDER="/app/.data/haproxy"

# Generate pem file
# Input : $1 - file name
generate_pem_file() {
    openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048 >/dev/null 2>&1
    openssl req -new -key private_key.pem -out csr.pem -subj "/C=XX" >/dev/null 2>&1
    openssl x509 -req -days 365 -in csr.pem -signkey private_key.pem -out certificate.pem >/dev/null 2>&1
    cat certificate.pem private_key.pem >"$1"
    # Remove temp files
    rm csr.pem
    rm private_key.pem
    rm certificate.pem
}

# DRIVER

# Reset the terminal
reset

# Set docker host
export DOCKER_HOST="unix:///var/run/docker.sock"

# Wait for docker to start by checking if docker ps is working .
echo "Waiting for docker to start..."
# Check failed if contains string "Cannot connect"
while true; do
    # Check if Docker is running
    if ! sudo docker info &>/dev/null; then
        echo "Docker is not running. Cannot connect."
        sleep 5
    else
        echo "Docker is running!"
        break
    fi
done
echo "Docker started successfully"

# Check if docker swarm is initialized
if ! sudo docker info | grep -q "Swarm: active"; then
    echo "Docker swarm is not initialized. Initializing docker swarm..."
    docker swarm init >/dev/null 2>&1
    echo "Docker swarm initialized successfully"
fi

# Find current node details from docker swarm
node_id=$(sudo docker node ls | grep $(hostname) | awk '{print $1}')
if [ -z "$node_id" ]; then
    echo "Node id not found. Exiting..."
    exit 1
fi

# Add label swiftwave_controlplane_node=true to current node
sudo docker node update --label-add swiftwave_controlplane_node=true "$node_id" >/dev/null 2>&1

# Create docker network
if ! sudo docker network ls | grep -q $SWARM_NETWORK; then
    echo "Creating docker network..."
    sudo docker network create --driver=overlay --attachable $SWARM_NETWORK >/dev/null 2>&1
    echo "Docker network created successfully"
else
    echo "Docker network already exists"
fi

# Create subfolders
mkdir -p "$SWIFTWAVE_APP_TARBALL_FOLDER"
mkdir -p "$SWIFTWAVE_REDIS_FOLDER"
mkdir -p "$SWIFTWAVE_HAPROXY_FOLDER"
mkdir -p "$SWIFTWAVE_HAPROXY_FOLDER/ssl"

# Update permissions
sudo chown -R user:user "$SWIFTWAVE_APP_TARBALL_FOLDER"
sudo chown -R user:user "$SWIFTWAVE_REDIS_FOLDER"
sudo chown -R user:user "$SWIFTWAVE_HAPROXY_FOLDER"
sudo chown -R user:user "$SWIFTWAVE_DATA_FOLDER"

# Generate default pem file
# as haproxy ssl_sni is enabled, without atleast one pem file, haproxy will not start
generate_pem_file "$SWIFTWAVE_HAPROXY_FOLDER/ssl/default.pem"

# Fetch this admin email, username and password from a file # TODO
admin_email="test@gmail.com"
admin_username="admin"
admin_password="admin"

# Generate brypt hash of admin password
admin_password_hash=$(htpasswd -bnBC 8 "" "$admin_password" | grep -oE '\$2[ayb]\$.{56}' | base64 -w 0)

# admin username and password hash
SWIFTWAVE_ADMIN_EMAIL="$admin_email"
SWIFTWAVE_ADMIN_USERNAME="$admin_username"
SWIFTWAVE_ADMIN_PASSWORD_HASH="$admin_password_hash"

# Read haproxy.cfg
haproxy_cfg=$(cat /app/dev/haproxy.sample.cfg)

# Get IP address of current node
ip_address=$IP

# Use sed to replace env variables in haproxy.cfg
haproxy_cfg=$(echo "$haproxy_cfg" | sed "s|\${PUBLIC_IP}|$ip_address|g")

# Write haproxy.cfg
echo "$haproxy_cfg" >"$SWIFTWAVE_HAPROXY_FOLDER/haproxy.cfg"

# Start redis
(cd "$SWIFTWAVE_REDIS_FOLDER" && echo "dbfilename dump.rdb" | redis-server -) > /dev/null 2>&1 &

# Pull docker images
sudo docker pull ghcr.io/swiftwave-org/swiftwave-dashboard:latest
sudo docker pull haproxytech/haproxy-debian:2.9
# Start the services
sudo docker stack deploy -c /app/dev/docker-compose.dev.yml $STACK_NAME

# Cd to /app
cd /app

# Message
echo -e "${GREEN}HaProxy and Redis started successfully ! ${NC}"
echo -e "${GREEN}Admin username : $SWIFTWAVE_ADMIN_USERNAME ${NC}"
echo -e "${GREEN}Admin password : $admin_password ${NC}"

# Set environment variables
export PORT="3333"
export ENVIRONMENT="development"
export ADMIN_USERNAME="$SWIFTWAVE_ADMIN_USERNAME"
export ADMIN_PASSWORD="$SWIFTWAVE_ADMIN_PASSWORD_HASH"
export CODE_TARBALL_DIR="$SWIFTWAVE_APP_TARBALL_FOLDER"
export SWARM_NETWORK="$SWARM_NETWORK"
export HAPROXY_SERVICE_NAME="swiftwave_haproxy"
export DATABASE_TYPE="sqlite"
export SQLITE_DATABASE="/app/.data/gorm.db"
export POSTGRESQL_URI=""
export REDIS_ADDRESS="127.0.0.1:6379"
export REDIS_PASSWORD=""
export ACCOUNT_EMAIL_ID="$SWIFTWAVE_ADMIN_EMAIL"
export ACCOUNT_PRIVATE_KEY_FILE_PATH="/app/.data/account_private_key.key"
export HAPROXY_MANAGER_HOST="127.0.0.1"
export HAPROXY_MANAGER_PORT="5555"
export HAPROXY_MANAGER_USERNAME="admin"
export HAPROXY_MANAGER_PASSWORD="mypassword"
export DOCKER_HOST="unix:///var/run/docker.sock"
export RESTRICTED_PORTS="80,443,5555,3333,1212"
export SESSION_TOKEN_EXPIRY_MINUTES="720"
export CGO_ENABLED="1"
echo -e "${GREEN}Environment variables set successfully ! ${NC}"

# Start bash as normal user
(sudo -H -E -u user ttyd -W -w ~ /bin/bash) & >/dev/null 2>&1
echo -e "${GREEN}TTYD Bash started successfully at port 7681 ! ${NC}"