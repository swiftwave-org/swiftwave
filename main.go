package main

import (
	"context"
	"fmt"
	SSL "keroku/m/ssl_manager"
	"net/http"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
		http.HandleFunc("/.well-known/acme-challenge/", func(w http.ResponseWriter, r *http.Request) {
			ssl_manager.ACMEHttpHandler(w, r)
		})
		http.HandleFunc("/.well-known/pre-authorize/", func(w http.ResponseWriter, r *http.Request) {
			ssl_manager.DNSConfigurationPreAuthorizeHttpHandler(w, r)
		})
		fmt.Println("Server listening on port 80...")
		err := http.ListenAndServe(":80", nil)
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	fmt.Println("Server started")

	// Request certificate
	wg.Add(1)
	go func() {
		err := ssl_manager.ObtainCertificate("apache.tanmoy.info")
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Done()
	}()

	wg.Wait()
}
