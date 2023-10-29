package pubsub

import (
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-set"
	"sync"
)

type Client interface {
	CreateTopic(topic string) error
	RemoveTopic(topic string) error
	Subscribe(topic string) (string, <-chan string, error)
	Publish(topic string, data string) error
	Close() error
}

type localPubSub struct {
	mutex         *sync.RWMutex
	bufferLength  int
	subscriptions map[string]map[string]localPubSubSubscription
	// <topic> -> [<subscriber> -> <channel>]
	topics *set.Set[string]
	closed bool
}

type localPubSubSubscription struct {
	Mutex   *sync.RWMutex
	Channel chan string
}

type remotePubSub struct {
	redisClient redis.Client
	closed      bool
}

type Type string

const (
	Local  Type = "local"
	Remote Type = "remote"
)

type Config struct {
	Type Type
	// Only for local pubsub
	// to store max number of messages in channel if no subscriber is listening
	BufferLength int
	// Only for remote pubsub, to store redis client
	RedisClient *redis.Client
}
