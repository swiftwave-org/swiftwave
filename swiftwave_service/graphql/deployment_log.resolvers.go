package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.41

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/swiftwave-org/swiftwave/pubsub"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
)

// FetchDeploymentLog is the resolver for the fetchDeploymentLog field.
func (r *subscriptionResolver) FetchDeploymentLog(ctx context.Context, id string) (<-chan *model.DeploymentLog, error) {
	// find deployment status
	deploymentStatus, err := core.FindDeploymentStatusByID(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	// create a channel
	var channel = make(chan *model.DeploymentLog, 200)

	go func() {
		defer close(channel)
		// defer handle panic
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in FetchDeploymentLog: %v", r)
				return
			}
		}()

		// check if deployment is pending or deploy_pending stage
		if *deploymentStatus == core.DeploymentStatusPending || *deploymentStatus == core.DeploymentStatusDeployPending {
			// from pubsub
			// channel name
			channelName := fmt.Sprintf("deployment-log-%s", id)
			// create a subscription
			subscriptionId, subscriptionChannel, err := r.ServiceManager.PubSubClient.Subscribe(channelName)
			if err != nil {
				return
			}
			// defer unsubscribe
			defer func(PubSubClient pubsub.Client, topic string, subscriptionId string) {
				err := PubSubClient.Unsubscribe(topic, subscriptionId)
				if err != nil {
					log.Println("error while unsubscribing from pubsub")
				}
			}(r.ServiceManager.PubSubClient, channelName, subscriptionId)
			// iterate over channel
			for {
				select {
				case <-ctx.Done():
					return
				case data, ok := <-subscriptionChannel:
					if !ok {
						return
					}
					// create a deployment log object
					var deploymentLog = &model.DeploymentLog{
						Content:   data,
						CreatedAt: time.Now(),
					}
					// check if channel full
					if len(channel) == cap(channel) {
						// remove first element
						<-channel
					}
					select {
					case <-ctx.Done():
						return
					case channel <- deploymentLog:
					}
				}
			}
		} else {
			// fetch all deployment logs
			deploymentLogs, err := core.FindAllDeploymentLogsByDeploymentId(ctx, r.ServiceManager.DbClient, id)
			if err == nil {
				for _, deploymentLog := range deploymentLogs {
					var deploymentLogGraphqlObject = deploymentLogToGraphqlObject(&deploymentLog)
					// check if channel full
					if len(channel) == cap(channel) {
						// remove first element
						<-channel
					}
					select {
					case <-ctx.Done():
						return
					case channel <- deploymentLogGraphqlObject:
					}
				}
			} else {
				log.Println(err)
			}
		}
	}()

	return channel, nil
}