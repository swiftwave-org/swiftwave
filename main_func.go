package main

import (
	"context"
	"encoding/json"
	"fmt"
	DOCKER "keroku/m/container_manager"
	HAProxy "keroku/m/haproxy_manager"
	SSL "keroku/m/ssl_manager"
	"os"
	"sync"

	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
		err := ssl_manager.ObtainCertificate("minc.tanmoy.info")
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

	// Create a new HAProxySocket
	var haproxySocket = HAProxy.HAProxySocket{}
	haproxySocket.InitTcpSocket("localhost", 5555)
	haproxySocket.Auth("admin", "mypassword")
	errFound := false
	transaction_id, err := haproxySocket.FetchNewTransactionId()
	if err != nil {
		print("Error while fetching HAProxy version: " + err.Error())
		os.Exit(1)
		return
	}

	// Add backend switch
	// err = haproxySocket.AddHTTPSLink(transaction_id, "be_minc-service_3000", "minc.tanmoy.info")
	// if err != nil {
	// 	errFound = true;
	// 	fmt.Println(err)
	// }

	// Delete backend switch
	// err = haproxySocket.DeleteHTTPLink(transaction_id, "be_minc-service_3000", "minc.tanmoy.info")
	// if err != nil {
	// 	errFound = true;
	// 	fmt.Println(err)
	// }

	// Add SSL certificate
	privateKey, err := os.ReadFile("/home/ubuntu/client_program/data/domain/private_key/minc.tanmoy.info.key")
	if err != nil {
		errFound = true
		fmt.Println(err)
	}
	fullChain, err := os.ReadFile("/home/ubuntu/client_program/data/domain/full_chain/minc.tanmoy.info.crt")
	if err != nil {
		errFound = true
		fmt.Println(err)
	}
	err = haproxySocket.UpdateSSL(transaction_id, "minc.tanmoy.info", privateKey, fullChain)
	fmt.Println(err)

	if errFound {
		fmt.Println("Deleting transaction: " + transaction_id)
		haproxySocket.DeleteTransaction(transaction_id)
		fmt.Println("Error found")
	} else {
		fmt.Println("Committing transaction: " + transaction_id)
		haproxySocket.CommitTransaction(transaction_id)
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
		Name: "nginx-service",
		Image: "nginx:latest",
		Command: []string{},
		Env: map[string]string{
			"APP_NAME": "nginx",
			"APP_COLOR": "blue",
		},
		VolumeMounts: []DOCKER.VolumeMount{},
		Networks: []string{
		},
		Replicas: 5,
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
