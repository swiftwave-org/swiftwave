package pubsub

import (
	"errors"
	"github.com/google/uuid"
	"github.com/hashicorp/go-set"
	"log"
	"sync"
)

func (l *localPubSub) CreateTopic(topic string) error {
	if l.closed {
		return errors.New("pubsub client is closed")
	}
	l.mutex.RLock()
	isContains := l.topics.Contains(topic)
	l.mutex.RUnlock()
	if isContains {
		return nil
	} else {
		l.mutex.Lock()
		// insert
		l.topics.Insert(topic)
		l.subscriptions[topic] = make(map[string]localPubSubSubscription)
		l.mutex.Unlock()
		return nil
	}
}

func (l *localPubSub) RemoveTopic(topic string) error {
	if l.closed {
		return errors.New("pubsub client is closed")
	}
	l.mutex.RLock()
	isContains := l.topics.Contains(topic)
	l.mutex.RUnlock()
	// check if topic exists
	if !isContains {
		return nil
	} else {
		l.mutex.Lock()
		// close all subscribers
		subscriptionRecords := l.subscriptions[topic]
		for _, subscription := range subscriptionRecords {
			m := subscription.Mutex
			m.Lock()
			close(subscription.Channel)
			m.Unlock()
		}
		// delete topic
		delete(l.subscriptions, topic)
		l.mutex.Unlock()
	}
	return nil
}

// Subscribe returns a subscription id and a channel to listen to
func (l *localPubSub) Subscribe(topic string) (string, <-chan string, error) {
	if l.closed {
		return "", nil, errors.New("pubsub client is closed")
	}
	// lock
	l.mutex.RLock()
	isContains := l.topics.Contains(topic)
	l.mutex.RUnlock()
	// check if topic exists
	if !isContains {
		l.mutex.Lock()
		// insert topic
		l.topics.Insert(topic)
		l.subscriptions[topic] = make(map[string]localPubSubSubscription)
		l.mutex.Unlock()
	}
	// create a new subscription id
	subscriptionId := topic + "_" + uuid.NewString()
	// create a new channel
	channel := make(chan string, l.bufferLength)
	// create a new subscription record
	subscriptionRecord := localPubSubSubscription{
		Mutex:   &sync.RWMutex{},
		Channel: channel,
	}
	// critical section
	l.mutex.Lock()
	// add subscription record to subscriptions
	l.subscriptions[topic][subscriptionId] = subscriptionRecord
	// unlock
	l.mutex.Unlock()
	// return subscription id and channel
	return subscriptionId, channel, nil
}

// Unsubscribe removes a subscription
func (l *localPubSub) Unsubscribe(topic string, subscriptionId string) error {
	if l.closed {
		return errors.New("pubsub client is closed")
	}
	// lock main mutex
	l.mutex.RLock()
	isContains := l.topics.Contains(topic)
	l.mutex.RUnlock()
	// check if topic exists
	if !isContains {
		return nil
	}
	// check if subscription exists
	l.mutex.RLock()
	if _, ok := l.subscriptions[topic][subscriptionId]; !ok {
		l.mutex.RUnlock()
		return nil
	}
	// fetch subscription record
	subscriptionRecord := l.subscriptions[topic][subscriptionId]
	l.mutex.RUnlock()
	// lock
	mutex := subscriptionRecord.Mutex
	mutex.Lock()
	// cleanup channel
	close(subscriptionRecord.Channel)
	mutex.Unlock()
	l.mutex.Lock()
	// delete subscription
	delete(l.subscriptions[topic], subscriptionId)
	l.mutex.Unlock()
	return nil
}

func (l *localPubSub) Publish(topic string, data string) error {
	// check if closed
	if l.closed {
		return errors.New("pubsub client is closed")
	}
	l.mutex.RLock()
	isContains := l.topics.Contains(topic)
	l.mutex.RUnlock()
	// check if topic exists
	if !isContains {
		l.mutex.Lock()
		// insert topic
		l.topics.Insert(topic)
		l.subscriptions[topic] = make(map[string]localPubSubSubscription)
		l.mutex.Unlock()
	}
	// fetch all subscriptions
	l.mutex.RLock()
	subscriptions := l.subscriptions[topic]
	l.mutex.RUnlock()
	// iterate over all subscriptions
	for _, subscriptionRecord := range subscriptions {
		// lock subscription mutex
		mutex := subscriptionRecord.Mutex
		mutex.Lock()
		channel := subscriptionRecord.Channel
		// clear channel if full
		if len(channel) == cap(channel) {
			<-channel
		}
		// send data
		channel <- data
		mutex.Unlock()
	}
	return nil
}

func (l *localPubSub) Close() error {
	// defer handle panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in localPubSub.Close: %v", r)
			return
		}
	}()
	// lock main mutex
	l.mutex.Lock()
	// unlock main mutex
	defer l.mutex.Unlock()
	// check if already closed
	if l.closed {
		return nil
	}
	// set closed to true
	l.closed = true
	// close
	for topic := range l.subscriptions {
		for _, subscription := range l.subscriptions[topic] {
			m := subscription.Mutex
			m.Lock()
			close(subscription.Channel)
			m.Unlock()
		}
	}
	// remove all topics
	l.topics = set.New[string](0)
	l.subscriptions = make(map[string]map[string]localPubSubSubscription)
	return nil
}
