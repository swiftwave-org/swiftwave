package main

import (
	"context"
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