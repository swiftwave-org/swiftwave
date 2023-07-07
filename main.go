package main

import (
	"context"
	"fmt"
	. "keroku/m/ssl_manager"
	"net/http"
	"sync"

	"github.com/redis/go-redis/v9"
)

func main()  {
	var wg sync.WaitGroup
	ctx := context.Background()
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
	options := SSLManagerOptions{
		Email: "tanmoysrt@gmail.com",
		AccountPrivateKeyFilePath: "/home/ubuntu/client_program/data/account_private_key.pem",
		DomainPrivateKeyStorePath: "/home/ubuntu/client_program/data/domain/private_key",
		DomainFullChainStorePath: "/home/ubuntu/client_program/data/domain/full_chain",
	}

	// Initialize SSLManager
	ssl_manager := SSLManager{}
	ssl_manager.Init(ctx, *rdb, options)

	// Start the HTTP server
	wg.Add(1)
	go func() {
		http.HandleFunc("/.well-known/acme-challenge/", func (w http.ResponseWriter, r *http.Request) {
			ssl_manager.ACMEHttpHandler(w, r);
		})
		http.HandleFunc("/.well-known/pre-authorize/", func (w http.ResponseWriter, r *http.Request) {
			ssl_manager.DNSConfigurationPreAuthorizeHttpHandler(w, r);
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
		err := ssl_manager.ObtainCertificate("minc.tanmoy.info");
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Done();
	}()

	wg.Wait()
}