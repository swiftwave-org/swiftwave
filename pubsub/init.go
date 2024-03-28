package pubsub

import (
	"context"
	"errors"
	"github.com/hashicorp/go-set"
	"strings"
	"sync"
)

func NewClient(options Options) (Client, error) {
	if options.Type == Local {
		return createLocalPubSubClient(options)
	} else if options.Type == Remote {
		return createRemotePubSubClient(options)
	} else {
		return nil, errors.New("invalid pubsub type")
	}
}

func createLocalPubSubClient(options Options) (Client, error) {
	// validate options
	if options.BufferLength <= 0 {
		return nil, errors.New("buffer length cannot be less than or equal to 0")
	}
	mutex := sync.RWMutex{}
	return &localPubSub{
		mutex:         &mutex,
		bufferLength:  options.BufferLength,
		subscriptions: make(map[string]map[string]localPubSubSubscription),
		topics:        set.New[string](0),
		closed:        false,
	}, nil
}

func createRemotePubSubClient(options Options) (Client, error) {
	// validate options
	if options.RedisClient == nil {
		return nil, errors.New("redis client is nil")
	}
	if strings.Compare(options.TopicsChannelName, "") == 0 {
		return nil, errors.New("topics channel name is empty")
	}
	if strings.Compare(options.EventsChannelName, "") == 0 {
		return nil, errors.New("events channel name is empty")
	}
	if options.BufferLength <= 0 {
		return nil, errors.New("buffer length cannot be less than or equal to 0")
	}
	mutex := sync.RWMutex{}
	client := remotePubSub{
		mutex:             &mutex,
		redisClient:       *options.RedisClient,
		bufferLength:      options.BufferLength,
		subscriptions:     make(map[string]map[string]remotePubSubSubscription),
		topicsChannelName: options.TopicsChannelName,
		eventsChannelName: options.EventsChannelName,
		eventsContext:     context.Background(),
		closed:            false,
	}
	go client.listenForBroadcastEvents(client.eventsContext)
	return &client, nil
}
