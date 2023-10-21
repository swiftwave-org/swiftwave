function Green
{
    process { Write-Host $_ -ForegroundColor Green }
}


function Yellow
{
    process { Write-Host $_ -ForegroundColor Yellow }
}


# Check if docker is available
if (Get-Command -Name "docker" -ErrorAction SilentlyContinue) {
    # Get base path
    $base_path=$PWD
    $dockerfile_path="$base_path\dev\Dockerfile"

    # Remove data folder
    Remove-Item -LiteralPath "data" -Force -Recurse -ErrorAction Ignore

    # Add .images folder if not exists
    New-Item -ItemType Directory -Force -Path "$base_path\.images" -ErrorAction SilentlyContinue >$null

    # Fetch required images
    Write-Output "Fetching required images..."
    docker pull ghcr.io/swiftwave-org/dashboard:develop
    docker pull haproxytech/haproxy-debian:2.9
    docker save -o .images\swiftwave-dashboard.tar ghcr.io/swiftwave-org/dashboard:develop
    docker save -o .images\haproxy-debian.tar haproxytech/haproxy-debian:2.9

    # build docker image
    Write-Output "Building docker image..."
    docker build -t swiftwave_dev_env:latest -f $dockerfile_path .

    # Remove existing container
    docker rm -f swiftwave_dev_env --force >$null

    Write-Output "Starting swiftwave dev environment..."
    # run dev/start.sh script
    $container_id = docker run --privileged -d -v go_cache:/home/user/go/pkg/mod/ -v ${PWD}:/app -w /app --name swiftwave_dev_env swiftwave_dev_env:latest 2>&1

    Write-Output "Perform post installation tasks..."
    # Run startup script by exec
    docker exec swiftwave_dev_env bash -c "cd /app && ./dev/start.sh"
    Write-Output "Post installation tasks completed"
    # Print IP
    $ip = docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $container_id 2>&1
    Write-Output "`n`nSwiftwave dev environment is started successfully" | Green
    Write-Output "Follow the steps below to start the server"
    Write-Output "---------------------------------------------------------"
    Write-Output "Open Bash Terminal : http://${ip}:7681" | Green
    Write-Output "---------------------------------------------------------"
    Write-Output "Start Swiftwave Server :" | Green
    Write-Output "  1. Open http://${ip}:7681 in browser"
    Write-Output "  2. Run go run . in terminal"
    Write-Output "  3. Wait for 3~4 minutes to start the server"
    Write-Output "---------------------------------------------------------"
    Write-Output "Access Swiftwave Dashboard :" | Green
    Write-Output "  1. Open http://${ip}:1212 in browser"
    Write-Output "  2. Configure server details from bottom bar -> Host ${ip} | Port 3333"
    Write-Output "  2. Login with admin as username and admin as password"
    Write-Output "---------------------------------------------------------"
    Write-Output "Stop Development Container :" | Green
    Write-Output "docker stop swiftwave_dev_env" | Yellow
    Write-Output "---------------------------------------------------------"
}
else {
    Write-Output "Docker is not available. Follow this documentation to install docker: https://docs.docker.com/desktop/install/windows-install/"
}
