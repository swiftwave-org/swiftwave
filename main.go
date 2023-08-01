package main

import (
	"context"
	DOCKER "keroku/m/container_manager"
	DOCKER_CONFIG_GENERATOR "keroku/m/docker_config_generator"
	HAPROXY "keroku/m/haproxy_manager"
	SSL "keroku/m/ssl_manager"

	DOCKER_CLIENT "github.com/docker/docker/client"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Initiating database client
	db_client, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Initiating SSL Manager
	options := SSL.ManagerOptions{
		IsStaging:                 false,
		Email:                     "tanmoysrt@gmail.com",
		AccountPrivateKeyFilePath: "/home/ubuntu/client_program/data/account_private_key.key",
		DomainPrivateKeyStorePath: "/home/ubuntu/client_program/data/domain/private_key",
		DomainFullChainStorePath:  "/home/ubuntu/client_program/data/domain/full_chain",
	}
	ssl_manager := SSL.Manager{}
	ssl_manager.Init(context.Background(), *db_client, options)

	// Initiating HAPROXY Manager
	var haproxy_manager = HAPROXY.Manager{}
	haproxy_manager.InitTcpSocket("localhost", 5555)
	haproxy_manager.Auth("admin", "mypassword")

	// Initiating Docker Manager
	docker_client, err := DOCKER_CLIENT.NewClientWithOpts(DOCKER_CLIENT.WithHost("tcp://127.0.0.1:2375"))
	if err != nil {
		panic(err)
	}
	docker_manager := DOCKER.Manager{}
	docker_manager.Init(context.Background(), *docker_client)

	// Initiating Docker Image Manager
	docker_config_generator := DOCKER_CONFIG_GENERATOR.Manager{}
	err = docker_config_generator.Init()
	if err != nil {
		panic(err)
	}
}
