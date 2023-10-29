package pubsub

import (
	"errors"
	"github.com/google/uuid"
	"log"
	"sync"
)

func (l *localPubSub) CreateTopic(topic string) error {
	if l.closed {
		return errors.New("pubsub client is closed")
	}
	if l.topics.Contains(topic) {
		return nil
	} else {
		// lock
		l.mutex.Lock()
		defer l.mutex.Unlock()
		// insert
		l.topics.Insert(topic)
		l.subscriptions[topic] = make(map[string]localPubSubSubscription)
		return nil
	}
}

func (l *localPubSub) RemoveTopic(topic string) error {
	if l.closed {
		return errors.New("pubsub client is closed")
	}
	// lock
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// check if topic exists
	if !l.topics.Contains(topic) {
		return nil
	} else {
		// close and remove all subscribers
		for subscriptionRecord := range l.subscriptions[topic] {
			l.cancelSubscriptionsOfTopic(topic, subscriptionRecord)
		}
		// delete subscribers
		delete(l.subscriptions, topic)
	}
	return nil
}

func (l *localPubSub) cancelSubscriptionsOfTopic(topic string, subscriptionId string) {
	if l.closed {
		return
	}
	// verify topic exists with ok
	if _, ok := l.subscriptions[topic]; !ok {
		return
	}
	// verify subscription exists
	if _, ok := l.subscriptions[topic][subscriptionId]; !ok {
		return
	}
	// lock
	mutex := l.subscriptions[topic][subscriptionId].Mutex
	mutex.Lock()
	defer mutex.Unlock()
	// close channel
	close(l.subscriptions[topic][subscriptionId].Channel)
	// delete subscription
	delete(l.subscriptions[topic], subscriptionId)
}

// Subscribe returns a subscription id and a channel to listen to
func (l *localPubSub) Subscribe(topic string) (string, <-chan string, error) {
	if l.closed {
		return "", nil, errors.New("pubsub client is closed")
	}
	// lock
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// check if topic exists
	if !l.topics.Contains(topic) {
		return "", nil, errors.New("topic does not exist")
	}
	// create a new subscription id
	subscriptionId := topic + "-" + uuid.NewString()
	// create a new channel
	channel := make(chan string, l.bufferLength)
	// create a new subscription record
	subscriptionRecord := localPubSubSubscription{
		Mutex:   &sync.RWMutex{},
		Channel: channel,
	}
	// add subscription record to subscriptions
	l.subscriptions[topic][subscriptionId] = subscriptionRecord
	// return subscription id and channel
	return subscriptionId, channel, nil
}

// Unsubscribe removes a subscription
func (l *localPubSub) Unsubscribe(topic string, subscriptionId string) error {
	if l.closed {
		return errors.New("pubsub client is closed")
	}
	// lock main mutex
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// check if topic exists
	if !l.topics.Contains(topic) {
		return errors.New("topic does not exist")
	}
	// check if subscription exists
	if _, ok := l.subscriptions[topic][subscriptionId]; !ok {
		return errors.New("subscription does not exist")
	}
	// fetch subscription record
	subscriptionRecord := l.subscriptions[topic][subscriptionId]
	// lock
	mutex := subscriptionRecord.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	// close channel
	close(subscriptionRecord.Channel)
	// delete subscription
	delete(l.subscriptions[topic], subscriptionId)
	return nil
}

func (l *localPubSub) Publish(topic string, data string) error {
	// check if closed
	if l.closed {
		return errors.New("pubsub client is closed")
	}
	// lock main mutex
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// check if topic exists
	if !l.topics.Contains(topic) {
		l.mutex.Unlock()
		return errors.New("topic does not exist")
	}
	// fetch all subscriptions
	subscriptions := l.subscriptions[topic]
	// iterate over all subscriptions
	for _, subscriptionRecord := range subscriptions {
		// lock subscription mutex
		mutex := subscriptionRecord.Mutex
		mutex.Lock()
		// clear channel if full
		if len(subscriptionRecord.Channel) == cap(subscriptionRecord.Channel) {
			<-subscriptionRecord.Channel
		}
		// send data
		subscriptionRecord.Channel <- data
		// unlock subscription mutex
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
	// close and remove all subscribers
	for topic := range l.subscriptions {
		for subscriptionRecord := range l.subscriptions[topic] {
			l.cancelSubscriptionsOfTopic(topic, subscriptionRecord)
		}
	}
	return nil
}
