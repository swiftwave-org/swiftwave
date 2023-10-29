package pubsub

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-set"
	"sync"
)

func CreatePubSubClient(config Config) (Client, error) {
	if config.Type == Local {
		if config.BufferLength == 0 {
			return nil, errors.New("buffer length must be greater than 0")
		}
		return createLocalPubSubClient(config.BufferLength)
	} else if config.Type == Remote {
		if config.RedisClient == nil {
			return nil, errors.New("redis client must not be nil")
		}
		return createRemotePubSubClient(config.RedisClient)
	} else {
		return nil, errors.New("invalid pubsub type")
	}
}

func createLocalPubSubClient(bufferLength int) (Client, error) {
	mutex := sync.RWMutex{}
	return &localPubSub{
		mutex:         &mutex,
		bufferLength:  bufferLength,
		subscriptions: make(map[string]map[string]localPubSubSubscription),
		topics:        set.New[string](0),
		closed:        false,
	}, nil
}

func createRemotePubSubClient(redisClient *redis.Client) (Client, error) {
	return &remotePubSub{
		redisClient: *redisClient,
		closed:      false,
	}, nil
}
