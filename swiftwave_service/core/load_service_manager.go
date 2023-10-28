package core

import (
	"context"
	DOCKER_CLIENT "github.com/docker/docker/client"
	DOCKER "github.com/swiftwave-org/swiftwave/container_manager"
	DOCKER_CONFIG_GENERATOR "github.com/swiftwave-org/swiftwave/docker_config_generator"
	HAPROXY "github.com/swiftwave-org/swiftwave/haproxy_manager"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"log"
	"os"
	"strconv"
	"strings"
)

func (manager *ServiceManager) Load() {
	// Initiating database client
	dbClient, err := createDbClient()
	if err != nil {
		panic(err.Error())
	}
	manager.DbClient = *dbClient

	// Initiating SSL Manager
	options := SSL.ManagerOptions{
		IsStaging:                 strings.Compare(os.Getenv("ENVIRONMENT"), "production") != 0,
		Email:                     os.Getenv("ACCOUNT_EMAIL_ID"),
		AccountPrivateKeyFilePath: os.Getenv("ACCOUNT_PRIVATE_KEY_FILE_PATH"),
	}
	sslManager := SSL.Manager{}
	err = sslManager.Init(context.Background(), *dbClient, options)
	if err != nil {
		panic(err)
	}
	manager.SslManager = sslManager

	// Initiating HAPROXY Manager
	var haproxyManager = HAPROXY.Manager{}
	haproxyPort, err := strconv.Atoi(os.Getenv("HAPROXY_MANAGER_PORT"))
	if err != nil {
		log.Fatal("HAPROXY_MANAGER_PORT environment variable is not set")
	}
	haproxyManager.InitTcpSocket(os.Getenv("HAPROXY_MANAGER_HOST"), haproxyPort)
	haproxyManager.Auth(os.Getenv("HAPROXY_MANAGER_USERNAME"), os.Getenv("HAPROXY_MANAGER_PASSWORD"))
	manager.HaproxyManager = haproxyManager

	// Initiating Docker Manager
	dockerClient, err := DOCKER_CLIENT.NewClientWithOpts(DOCKER_CLIENT.WithHost(os.Getenv("DOCKER_HOST")))
	if err != nil {
		panic(err)
	}
	dockerManager := DOCKER.Manager{}
	err = dockerManager.Init(context.Background(), *dockerClient)
	if err != nil {
		panic(err)
	}
	manager.DockerManager = dockerManager

	// Initiating Docker Config Generator
	dockerConfigGenerator := DOCKER_CONFIG_GENERATOR.Manager{}
	err = dockerConfigGenerator.Init()
	if err != nil {
		panic(err)
	}
	manager.DockerConfigGenerator = dockerConfigGenerator
}
