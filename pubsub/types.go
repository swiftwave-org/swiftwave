package pubsub

import (
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-set"
	"sync"
)

type PubSub interface {
	CreateTopic(topic string) error
	RemoveTopic(topic string) error
	Subscribe(topic string) (<-chan interface{}, error)
	Publish(topic string, data interface{}) error
	Close() error
}

type localPubSub struct {
	mutex         *sync.RWMutex
	subscriptions map[string]map[string]localPubSubSubscription
	// <topic> -> [<subscriber> -> <channel>]
	topics set.Set[string]
	closed bool
}

type localPubSubSubscription struct {
	Mutex   *sync.RWMutex
	Channel chan interface{}
}

type remotePubSub struct {
	RedisClient redis.Client
	Closed      bool
}

type Type string

const (
	Local  Type = "local"
	Remote Type = "remote"
)

type Config struct {
	Type        Type
	RedisClient *redis.Client
}
