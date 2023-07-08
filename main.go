package main

import (
	"context"
	"fmt"
	SSL "keroku/m/ssl_manager"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/labstack/echo/v4"
)

func main() {
	var wg sync.WaitGroup
	ctx := context.Background()

	// Initiating database
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	options := SSL.ManagerOptions{
		Email:                     "tanmoysrt@gmail.com",
		AccountPrivateKeyFilePath: "/home/ubuntu/client_program/data/account_private_key.pem",
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
		fmt.Println("Server listening on port 80...")
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
