package main

import (
	"context"
	"encoding/json"
	"fmt"
	DOCKER "keroku/m/container_manager"
	IMAGE_MANAGER "keroku/m/docker_config_generator"
	GIT "keroku/m/git_manager"
	HAProxy "keroku/m/haproxy_manager"
	SSL "keroku/m/ssl_manager"
	"os"
	"sync"

	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GitTest() {
	manager := GIT.Manager{}
	manager.Init(GIT.GitUser{
		Username: "tanmoysrt",
		Password: "",
	})
	// fmt.Println(manager.FetchRepositories())
	var repo GIT.Repository = GIT.Repository{
		Name:      "dsa-leetcode-solutions",
		Username:  "tanmoysrt",
		Branch:    "main",
		IsPrivate: false,
	}
	fmt.Println(manager.FetchLatestCommitHash(repo))
	fmt.Println(manager.FetchFolderStructure(repo))
	fmt.Println(manager.FetchFileContent(repo, "941.valid-mountain-array.java"))
}

func DockerconfiggeneratorTest() {
	manager := IMAGE_MANAGER.Manager{}
	manager.Init()
	fmt.Println(manager.DefaultArgs("nextjs"))
	// git_manager := GIT.Manager{}
	// git_manager.Init(GIT.GitUser{
	// 	Username: "tanmoysrt",
	// 	Password: "",
	// })
	// var repo GIT.Repository = GIT.Repository{
	// 	Name: "spring-petclinic",
	// 	Username: "spring-projects",
	// 	Branch: "main",
	// 	IsPrivate: false,
	// }
	// fmt.Println(manager.DetectService(git_manager, repo))
}

func ImageGenerateTest() {
	// image manager
	image_manager := IMAGE_MANAGER.Manager{}
	image_manager.Init()

	// git manager
	git_manager := GIT.Manager{}
	git_manager.Init(GIT.GitUser{
		Username: "tanmoysrt",
		Password: "",
	})
	var git_repo GIT.Repository = GIT.Repository{
		Name:      "react-todo-app",
		Username:  "kabirbaidhya",
		Branch:    "master",
		IsPrivate: false,
	}

	// container manager
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost("tcp://127.0.0.1:2375"))
	if err != nil {
		panic(err)
	}
	dClient := DOCKER.Manager{}
	dClient.Init(ctx, *cli)

	serviceName, err := image_manager.DetectService(git_manager, git_repo)
	if err != nil {
		fmt.Println(err)
		return
	}
	// path, err := git_manager.CloneRepository(git_repo)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	path := "/tmp/keroku/756d3a78-3a2b-44dd-a813-0b9e69475747"
	dockerfile := image_manager.DockerTemplates[serviceName]
	buildargs := image_manager.DefaultArgs(serviceName)

	// Create image
	scanner, err := dClient.CreateImage(dockerfile, buildargs, path, "todo-app-vvvv")
	if err != nil {
		fmt.Println(err)
		return
	}
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func RunSSLSystem() {
	var wg sync.WaitGroup
	ctx := context.Background()

	// Initiating database
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	options := SSL.ManagerOptions{
		IsStaging:                 false,
		Email:                     "tanmoysrt@gmail.com",
		AccountPrivateKeyFilePath: "/home/ubuntu/client_program/data/account_private_key.key",
		DomainPrivateKeyStorePath: "/home/ubuntu/client_program/data/domain/private_key",
		DomainFullChainStorePath:  "/home/ubuntu/client_program/data/domain/full_chain",
	}

	// Initialize Manager
	ssl_manager := SSL.Manager{}
	ssl_manager.Init(ctx, *db, options)

	// Start the HTTP server
	wg.Add(1)
	go func() {
		server := echo.New()
		ssl_manager.InitHttpHandlers(server)
		fmt.Println("Server listening on port 8888...")
		err := server.Start(":80")
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	fmt.Println("Server started")

	// Request certificate
	wg.Add(1)
	go func() {
		err := ssl_manager.ObtainCertificate("nginx.tanmoy.info")
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Done()
	}()

	wg.Wait()
}

func SSLUpdate() {

	// RunSSLSystem()
	// return;
	var wg sync.WaitGroup

	// Create a new Manager
	var Manager = HAProxy.Manager{}
	Manager.InitTcpSocket("localhost", 5555)
	Manager.Auth("admin", "mypassword")
	errFound := false
	transaction_id, err := Manager.FetchNewTransactionId()
	if err != nil {
		print("Error while fetching HAProxy version: " + err.Error())
		os.Exit(1)
		return
	}

	// Add backend switch
	// err = Manager.AddHTTPSLink(transaction_id, "be_minc-service_3000", "nginx.tanmoy.info")
	// if err != nil {
	// 	errFound = true;
	// 	fmt.Println(err)
	// }

	// Delete backend switch
	// err = Manager.DeleteHTTPLink(transaction_id, "be_minc-service_3000", "minc.tanmoy.info")
	// if err != nil {
	// 	errFound = true;
	// 	fmt.Println(err)
	// }

	// Add SSL certificate
	privateKey, err := os.ReadFile("/home/ubuntu/client_program/data/domain/private_key/nginx.tanmoy.info.key")
	if err != nil {
		errFound = true
		fmt.Println(err)
	}
	fullChain, err := os.ReadFile("/home/ubuntu/client_program/data/domain/full_chain/nginx.tanmoy.info.crt")
	if err != nil {
		errFound = true
		fmt.Println(err)
	}
	err = Manager.UpdateSSL(transaction_id, "nginx.tanmoy.info", privateKey, fullChain)
	fmt.Println(err)

	if errFound {
		fmt.Println("Deleting transaction: " + transaction_id)
		Manager.DeleteTransaction(transaction_id)
		fmt.Println("Error found")
	} else {
		fmt.Println("Committing transaction: " + transaction_id)
		Manager.CommitTransaction(transaction_id)
		fmt.Println("No error found")
	}

	// Wait for events
	wg.Wait()
	fmt.Println("done")
}

func TestDocker() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost("tcp://127.0.0.1:2375"))
	if err != nil {
		panic(err)
	}
	dClient := DOCKER.Manager{}
	dClient.Init(ctx, *cli)

	spec := DOCKER.Service{
		Name:    "nginx-service",
		Image:   "nginx:latest",
		Command: []string{},
		Env: map[string]string{
			"APP_NAME":  "nginx",
			"APP_COLOR": "blue",
		},
		VolumeMounts: []DOCKER.VolumeMount{},
		Networks:     []string{},
		Replicas:     5,
	}

	// dClient.CreateService(spec)

	data, err := dClient.StatusService(spec)
	if err != nil {
		fmt.Println(err)
	} else {
		d, _ := json.Marshal(data)
		fmt.Println(string(d))
	}

	// fmt.Println(dClient.CreateVolume())
	// fmt.Println(dClient.VolumeUsage())
	// fmt.Println(dClient.VolumeUsage("a3ca139b3619064a066eacf9645ec2d6df4bf2d2e1ca5820fbaabe45f57e7e56"))
	// fmt.Println(dClient.RemoveVolume("835b690a85584c0f59902b4ff730c30740591527fcba89820534ac5cad6d14eb"))
	// fmt.Println(dClient.VolumeExists("c39afeeef0d0473a99eb4a8c86ea63ccb341b4ed63f14cc183ba4e7264232bb9e89cb225f931484aad8d3b62709cf261d08ee740037f4a2eaf03389c1d202ec4"))

	// dClient.CreateService("mincc-service", 2);
	// dClient.UpdateService("mincc-service", "njryymk1cayxvme5m6sex5jlu", 5);
}

func TestDockerNetwork() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost("tcp://127.0.0.1:2375"))
	if err != nil {
		panic(err)
	}
	dClient := DOCKER.Manager{}
	dClient.Init(ctx, *cli)

	// fmt.Println(dClient.DeleteNetwork("my-attachable-overlay"))
	fmt.Println(dClient.ExistsNetwork("swarm-network"))
	fmt.Println(dClient.CIDRNetwork("swarm-network"))
	fmt.Println(dClient.GatewayNetwork("swarm-network"))
}

func FullFledgeTest() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost("tcp://127.0.0.1:2375"))
	if err != nil {
		panic(err)
	}
	dockerClient := DOCKER.Manager{}
	dockerClient.Init(ctx, *cli)

	// Main overlay network
	mainNetwork := "swarm-network"

	// Create overlay network
	overlayNetwork := "nginx-service-network"
	err = dockerClient.CreateNetwork(overlayNetwork)
	if err != nil {
		panic(err)
	}

	// Create volume
	err = dockerClient.CreateVolume("nginx-service-volume")
	if err != nil {
		panic(err)
	}

	// Create nginx service
	spec := DOCKER.Service{
		Name:    "nginx-service",
		Image:   "nginx:latest",
		Command: []string{},
		Env:     map[string]string{},
		VolumeMounts: []DOCKER.VolumeMount{
			{
				Source:   "nginx-service-volume",
				Target:   "/etc/nginx",
				ReadOnly: false,
			},
		},
		Networks: []string{
			mainNetwork,
			overlayNetwork,
		},
		Replicas: 3,
	}
	err = dockerClient.CreateService(spec)
	if err != nil {
		panic(err)
	}

	// HAProxy
	var wg sync.WaitGroup

	// Create a new Manager
	var Manager = HAProxy.Manager{}
	Manager.InitTcpSocket("localhost", 5555)
	Manager.Auth("admin", "mypassword")
	errFound := false
	transaction_id, err := Manager.FetchNewTransactionId()
	if err != nil {
		print("Error while fetching HAProxy version: " + err.Error())
		os.Exit(1)
		return
	}

	// Add backend
	err = Manager.AddBackend(transaction_id, "nginx-service", 80, 3)
	if err != nil {
		errFound = true
		fmt.Println(err)
	}

	// Add backend switch
	err = Manager.AddHTTPSLink(transaction_id, "be_nginx-service_80", "nginx.tanmoy.info")
	if err != nil {
		errFound = true
		fmt.Println(err)
	}

	// Generate SSL certificate  --> Run seperately -- in production use queue

	// Add SSL certificate
	privateKey, err := os.ReadFile("/home/ubuntu/client_program/data/domain/private_key/nginx.tanmoy.info.key")
	if err != nil {
		errFound = true
		fmt.Println(err)
	}
	fullChain, err := os.ReadFile("/home/ubuntu/client_program/data/domain/full_chain/nginx.tanmoy.info.crt")
	if err != nil {
		errFound = true
		fmt.Println(err)
	}
	err = Manager.UpdateSSL(transaction_id, "nginx.tanmoy.info", privateKey, fullChain)
	fmt.Println(err)

	if errFound {
		fmt.Println("Deleting transaction: " + transaction_id)
		Manager.DeleteTransaction(transaction_id)
		fmt.Println("Error found")
	} else {
		fmt.Println("Committing transaction: " + transaction_id)
		Manager.CommitTransaction(transaction_id)
		fmt.Println("No error found")
	}

	// Wait for events
	wg.Wait()
	fmt.Println("done")
}

func CustomFrontendTCPTest() {
	// Create a new Manager
	var Manager = HAProxy.Manager{}
	Manager.InitTcpSocket("localhost", 5555)
	Manager.Auth("admin", "mypassword")
	transaction_id, err := Manager.FetchNewTransactionId()
	if err != nil {
		print("Error while fetching HAProxy version: " + err.Error())
		os.Exit(1)
		return
	}

	err = Manager.AddTCPLink(transaction_id, "be_minc-service_3000", 5555, "test2.tanmoy.info", HAProxy.HTTPMode)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Committing transaction: " + transaction_id)
	err = Manager.CommitTransaction(transaction_id)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("No error found")
	}
}
