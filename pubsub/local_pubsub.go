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
		// create a wait group
		var wg sync.WaitGroup
		// close and remove all subscribers
		for subscriptionRecord := range l.subscriptions[topic] {
			wg.Add(1)
			go l.cancelSubscriptionsOfTopic(topic, subscriptionRecord, &wg)
		}
		// wait for all goroutines to finish
		wg.Wait()
		// delete subscribers
		delete(l.subscriptions, topic)
	}
	return nil
}

func (l *localPubSub) cancelSubscriptionsOfTopic(topic string, subscriptionId string, wg *sync.WaitGroup) {
	if l.closed {
		return
	}
	// defer wait group
	defer wg.Done()
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
	channel := make(chan string, 50)
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

func (l *localPubSub) Publish(topic string, data string) error {
	if l.closed {
		return errors.New("pubsub client is closed")
	}
	// check if topic exists
	if !l.topics.Contains(topic) {
		return errors.New("topic does not exist")
	}
	// fetch all subscriptions
	subscriptions := l.subscriptions[topic]
	// iterate over all subscriptions

	for _, subscriptionRecord := range subscriptions {
		// lock
		mutex := subscriptionRecord.Mutex
		mutex.Lock()
		// clear channel if full
		if len(subscriptionRecord.Channel) == cap(subscriptionRecord.Channel) {
			<-subscriptionRecord.Channel
		}
		// send data
		subscriptionRecord.Channel <- data
		// unlock
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
	// lock
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// check if already closed
	if l.closed {
		return nil
	}
	// set closed to true
	l.closed = true
	// create a wait group
	var wg sync.WaitGroup
	// close and remove all subscribers
	for topic := range l.subscriptions {
		for subscriptionRecord := range l.subscriptions[topic] {
			wg.Add(1)
			go l.cancelSubscriptionsOfTopic(topic, subscriptionRecord, &wg)
		}
	}
	// wait for all goroutines to finish
	wg.Wait()
	return nil
}
