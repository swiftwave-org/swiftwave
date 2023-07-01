package main

import (
	"fmt"
	"sync"

	"github.com/docker/docker/client"
)



func main() {
	var wg sync.WaitGroup
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	if cli == nil {
		panic("cli is nil")
	}
	// Start listening for events
	// wg.Add(1)
	// go listenForEvents(cli, &wg)

	// Fetch services
	// wg.Add(1)
	// go fetchService(cli, &wg)

	// Fetch DNS
	fmt.Println(fetchDNSRecord("hashnode.network", "A"))

	// Wait for events
	wg.Wait()
	fmt.Println("done")
}
