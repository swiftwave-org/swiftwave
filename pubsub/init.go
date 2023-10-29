package pubsub

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-set"
	"sync"
)

func CreatePubSubClient(config Config) (PubSub, error) {
	if config.Type == Local {
		return createLocalPubSubClient()
	} else if config.Type == Remote {
		return createRemotePubSubClient(config.RedisClient)
	} else {
		return nil, errors.New("invalid pubsub type")
	}
}

func createLocalPubSubClient() (PubSub, error) {
	return &localPubSub{
		mutex:         sync.RWMutex{},
		subscriptions: make(map[string]map[string][]chan interface{}),
		topics:        set.Set[string]{},
		closed:        false,
	}, nil
}

func createRemotePubSubClient(redisClient *redis.Client) (PubSub, error) {
	return &remotePubSub{
		redisClient: *redisClient,
	}, nil
}
