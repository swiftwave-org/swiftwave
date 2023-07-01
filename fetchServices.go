package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)


func fetchService(cli *client.Client, wg *sync.WaitGroup){
	defer wg.Done()
	ctx := context.Background()
	services, err := cli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		panic(err)
	}
	for _, service := range services {
		marshalled, err := json.Marshal(service)
		if err == nil {
			fmt.Println(string(marshalled))
		}
		fmt.Println("===================")
	}

}