package pubsub

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-set"
	"sync"
)

func CreatePubSubClient(config Config) (Client, error) {
	if config.Type == Local {
		return createLocalPubSubClient()
	} else if config.Type == Remote {
		return createRemotePubSubClient(config.RedisClient)
	} else {
		return nil, errors.New("invalid pubsub type")
	}
}

func createLocalPubSubClient() (Client, error) {
	mutex := sync.RWMutex{}
	return &localPubSub{
		mutex:         &mutex,
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
