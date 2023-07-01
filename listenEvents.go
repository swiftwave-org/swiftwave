package main

import (
	"context"
	"fmt"
	"sync"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func listenForEvents(cli *client.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	events, errs := cli.Events(ctx, types.EventsOptions{})
	for{
		select {
			case err := <-errs:{
				print(err)
			}
			case msg := <-events:{
				tmp:=""
				for k,v := range msg.Actor.Attributes{
					tmp+=k+"="+v+" "
				}
				print(msg.Type+" "+msg.Action+" "+msg.Actor.ID+" "+tmp+" "+msg.Scope+"\n")
			}
		  }
	}
}

func main() {
	var wg sync.WaitGroup
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	// Start listening for events
	wg.Add(1)
	go listenForEvents(cli, &wg)

	// Wait for events
	wg.Wait()
	fmt.Println("done")
}