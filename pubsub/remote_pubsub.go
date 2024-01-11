package pubsub

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"log"
	"sync"
)

func (r *remotePubSub) CreateTopic(topic string) error {
	// lock
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// add this to `SET` of redis
	// docs: https://redis.io/commands/sadd/
	err := r.redisClient.SAdd(context.Background(), r.topicsChannelName, topic).Err()
	if err != nil {
		return err
	}
	// create a map for this topic
	r.subscriptions[topic] = make(map[string]remotePubSubSubscription)
	return nil
}

func (r *remotePubSub) RemoveTopic(topic string) error {
	// lock mutex
	r.mutex.Lock()
	// unlock mutex
	defer r.mutex.Unlock()
	// remove this from `SET` of redis
	// docs: https://redis.io/commands/srem/
	err := r.redisClient.SRem(context.Background(), r.topicsChannelName, topic).Err()
	if err != nil {
		return err
	}
	// check if topic exists
	if _, ok := r.subscriptions[topic]; !ok {
		return nil
	}
	// send a message to `eventsChannelName` to close all subscriptions
	// docs: https://redis.io/commands/publish/
	message := "close-topic-" + topic
	err = r.redisClient.Publish(context.Background(), r.eventsChannelName, message).Err()
	if err != nil {
		return errors.New("error in broadcasting close topic message")
	}
	return nil
}

func (r *remotePubSub) Subscribe(topic string) (string, <-chan string, error) {
	// lock mutex
	r.mutex.RLock()
	// check if topic exists in `SET` of redis
	// docs: https://redis.io/commands/sismember/
	exists, err := r.redisClient.SIsMember(context.Background(), r.topicsChannelName, topic).Result()
	r.mutex.RUnlock()
	if err != nil {
		return "", nil, err
	}
	if !exists {
		r.mutex.Lock()
		err := r.redisClient.SAdd(context.Background(), r.topicsChannelName, topic).Err()
		if err != nil {
			r.mutex.Unlock()
			return "", nil, err
		}
		// create a map for this topic
		r.subscriptions[topic] = make(map[string]remotePubSubSubscription)
		r.mutex.Unlock()
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// open a channel for this topic
	// docs: https://redis.io/commands/subscribe/
	// docs: https://redis.io/topics/pubsub
	// docs: https://pkg.go.dev/github.com/go-redis/redis/v8#example-PubSub-Subscribe
	redisSubscriptionRef := r.redisClient.Subscribe(context.Background(), topic)
	// add this subscription to map
	subscriptionId := topic + "_" + uuid.NewString()
	mutex := &sync.RWMutex{}
	subscription := remotePubSubSubscription{
		Mutex:   mutex,
		Channel: r.castChannelType(topic, subscriptionId, redisSubscriptionRef.Channel()),
		PubSub:  redisSubscriptionRef,
	}
	r.subscriptions[topic][subscriptionId] = subscription
	return subscriptionId, subscription.Channel, nil
}

func (r *remotePubSub) Unsubscribe(topic string, subscriptionId string) error {
	// lock mutex
	r.mutex.RLock()
	// check if topic exists in `SET` of redis
	// docs: https://redis.io/commands/sismember/
	exists, err := r.redisClient.SIsMember(context.Background(), r.topicsChannelName, topic).Result()
	// unlock mutex
	r.mutex.RUnlock()
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	// cancel subscription
	err = r.cancelSubscription(topic, subscriptionId)
	if err != nil {
		return err
	}
	r.mutex.Lock()
	// delete subscription
	delete(r.subscriptions[topic], subscriptionId)
	r.mutex.Unlock()
	return nil
}

func (r *remotePubSub) Publish(topic string, data string) error {
	// publish to redis
	// docs: https://redis.io/commands/publish/
	return r.redisClient.Publish(context.Background(), topic, data).Err()
}

func (r *remotePubSub) Close() error {
	// lock mutex
	r.mutex.RLock()
	subscriptions := r.subscriptions
	// unlock mutex
	r.mutex.RUnlock()
	// close all subscriptions
	for topic := range subscriptions {
		r.cancelAllSubscriptionsOfTopic(topic)
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// close redis client
	err := r.redisClient.Close()
	if err != nil {
		return err
	}
	// set closed to true
	r.closed = true
	return nil
}

// private functions
func (r *remotePubSub) castChannelType(topic string, subscriptionId string, channel <-chan *redis.Message) chan string {
	c := make(chan string, r.bufferLength)
	go func() {
		// defer recover from panic
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic", r)
			}
		}()
		for {
			msg, ok := <-channel
			// check if `input` is closed
			if !ok {
				// verify if `output` channel is not closed
				if _, ok := <-c; ok {
					// close
					err := r.cancelSubscription(topic, subscriptionId)
					if err != nil {
						log.Println("error in canceling subscription of topic", topic, "with id", subscriptionId, ":", err)
					}
				}
				// skip and break
				break
			}
			// send to `output` channel
			c <- msg.Payload
		}
	}()
	return c
}

func (r *remotePubSub) cancelAllSubscriptionsOfTopic(topic string) {
	// NOTE: a lock is need to be acquired before calling this function
	// verify topic exists with ok
	r.mutex.RLock()
	if _, ok := r.subscriptions[topic]; !ok {
		r.mutex.RUnlock()
		return
	}
	// iterate over all subscriptions
	subscriptions := r.subscriptions[topic]
	r.mutex.RUnlock()
	for id := range subscriptions {
		err := r.cancelSubscription(topic, id)
		if err != nil {
			log.Println("error in canceling subscription of topic", topic, "with id", id, ":", err)
		}
	}
}

func (r *remotePubSub) cancelSubscription(topic string, subscriptionId string) error {
	r.mutex.RLock()
	// verify topic exists locally
	if _, ok := r.subscriptions[topic]; !ok {
		r.mutex.RUnlock()
		return nil
	}
	// verify subscription exists locally
	if _, ok := r.subscriptions[topic][subscriptionId]; !ok {
		r.mutex.RUnlock()
		return nil
	}
	// fetch subscription record
	subscriptionRecord := r.subscriptions[topic][subscriptionId]
	r.mutex.RUnlock()
	// lock
	mutex := subscriptionRecord.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	// close channel
	close(subscriptionRecord.Channel)
	// close redis pubsub
	err := subscriptionRecord.PubSub.Close()
	if err != nil {
		return err
	}
	return nil
}

func (r *remotePubSub) removeTopicAndCleanup(topic string) {
	// cancel all subscriptions of this topic
	r.cancelAllSubscriptionsOfTopic(topic)
	r.mutex.Lock()
	// delete topic
	delete(r.subscriptions, topic)
	r.mutex.Unlock()
}

func (r *remotePubSub) listenForBroadcastEvents(ctx context.Context) {
	// subscribe to `eventsChannelName`
	// docs: https://redis.io/commands/subscribe/
	subscribeRef := r.redisClient.Subscribe(context.Background(), r.eventsChannelName)
	channel := subscribeRef.Channel()
	// listen for messages
	for {
		select {
		case <-ctx.Done():
			// close subscribeRef
			err := subscribeRef.Close()
			if err != nil {
				log.Println("Error in closing redis subscribeRef", err)
			}
			return
		case msg, ok := <-channel:
			// check if channel is closed
			if !ok {
				log.Println("Channel closed on consumer")
				return
			}
			// check if message is for closing topic
			if msg.Payload[:12] == "close-topic-" {
				// close all subscriptions of this topic
				topic := msg.Payload[12:]
				go r.removeTopicAndCleanup(topic)
			}
		}
	}
}
