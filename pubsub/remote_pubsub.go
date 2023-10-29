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
	// remove this from `SET` of redis
	// docs: https://redis.io/commands/srem/
	err := r.redisClient.SRem(context.Background(), r.topicsChannelName, topic).Err()
	if err != nil {
		return err
	}
	// lock mutex
	r.mutex.Lock()
	// unlock mutex
	defer r.mutex.Unlock()
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
	r.mutex.Lock()
	// unlock mutex
	defer r.mutex.Unlock()
	// check if topic exists in `SET` of redis
	// docs: https://redis.io/commands/sismember/
	exists, err := r.redisClient.SIsMember(context.Background(), r.topicsChannelName, topic).Result()
	if err != nil {
		return "", nil, err
	}
	if !exists {
		return "", nil, errors.New("topic does not exist")
	}
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
		Channel: castChannelType(redisSubscriptionRef.Channel()),
		PubSub:  redisSubscriptionRef,
	}
	r.subscriptions[topic][subscriptionId] = subscription
	return subscriptionId, subscription.Channel, nil
}

func (r *remotePubSub) Unsubscribe(topic string, subscriptionId string) error {
	// lock mutex
	r.mutex.Lock()
	// unlock mutex
	defer r.mutex.Unlock()
	// check if topic exists in `SET` of redis
	// docs: https://redis.io/commands/sismember/
	exists, err := r.redisClient.SIsMember(context.Background(), r.topicsChannelName, topic).Result()
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("topic does not exist")
	}
	// check if topic exists
	if _, ok := r.subscriptions[topic]; !ok {
		return errors.New("topic does not exist")
	}
	// check if subscription exists
	if _, ok := r.subscriptions[topic][subscriptionId]; !ok {
		return errors.New("subscription does not exist")
	}
	// close channel
	lock := r.subscriptions[topic][subscriptionId].Mutex
	lock.Lock()
	defer lock.Unlock()
	// channel
	channel := r.subscriptions[topic][subscriptionId].Channel
	// close channel
	if _, ok := <-channel; ok {
		close(channel)
	}
	// close redis pubsub
	err = r.subscriptions[topic][subscriptionId].PubSub.Close()
	if err != nil {
		return err
	}
	// delete subscription
	delete(r.subscriptions, subscriptionId)
	return nil
}

func (r *remotePubSub) Publish(topic string, data string) error {
	// publish to redis
	// docs: https://redis.io/commands/publish/
	return r.redisClient.Publish(context.Background(), topic, data).Err()
}

func (r *remotePubSub) Close() error {
	// TODO: close local channels
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
func castChannelType(channel <-chan *redis.Message) chan string {
	c := make(chan string, 100)
	go func() {
		for {
			msg, ok := <-channel
			// check if `input` is closed
			if !ok {
				// verify if `output` channel is not closed
				if _, ok := <-c; ok {
					// close
					close(c)
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
	if _, ok := r.subscriptions[topic]; !ok {
		return
	}
	// iterate over all subscriptions
	for id, _ := range r.subscriptions[topic] {
		r.cancelSubscription(topic, id)
	}
}

func (r *remotePubSub) cancelSubscription(topic string, subscriptionId string) {
	// verify topic exists locally
	if _, ok := r.subscriptions[topic]; !ok {
		return
	}
	// verify subscription exists locally
	if _, ok := r.subscriptions[topic][subscriptionId]; !ok {
		return
	}
	// fetch subscription record
	subscriptionRecord := r.subscriptions[topic][subscriptionId]
	// lock
	mutex := subscriptionRecord.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	// close channel
	close(subscriptionRecord.Channel)
	// close redis pubsub
	err := subscriptionRecord.PubSub.Close()
	if err != nil {
		log.Println("Error in closing redis pubsub", err)
	}
}

func (r *remotePubSub) removeTopicAndCleanup(topic string) {
	// lock
	r.mutex.Lock()
	// unlock
	defer r.mutex.Unlock()
	// cancel all subscriptions of this topic
	r.cancelAllSubscriptionsOfTopic(topic)
	// delete topic
	delete(r.subscriptions, topic)
}

func (r *remotePubSub) listenForBroadcastEvents(ctx context.Context) {
	// subscribe to `eventsChannelName`
	// docs: https://redis.io/commands/subscribe/
	pubsub := r.redisClient.Subscribe(context.Background(), r.eventsChannelName)
	channel := pubsub.Channel()
	// listen for messages
	for {
		select {
		case <-ctx.Done():
			// close pubsub
			err := pubsub.Close()
			if err != nil {
				log.Println("Error in closing redis pubsub", err)
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
