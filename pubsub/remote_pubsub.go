package pubsub

func (r *remotePubSub) CreateTopic(topic string) error {
	return nil
}

func (r *remotePubSub) RemoveTopic(topic string) error {
	return nil
}

func (r *remotePubSub) Subscribe(topic string) (string, <-chan string, error) {
	return "", nil, nil
}

func (r *remotePubSub) Unsubscribe(topic string, subscriptionId string) error {
	return nil
}

func (r *remotePubSub) Publish(topic string, data string) error {
	return nil
}

func (r *remotePubSub) Close() error {
	return nil
}
