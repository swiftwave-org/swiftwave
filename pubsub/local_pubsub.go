package pubsub

func (l *localPubSub) CreateTopic(topic string) error {
	return nil
}

func (l *localPubSub) RemoveTopic(topic string) error {
	return nil
}

func (l *localPubSub) Subscribe(topic string) (<-chan interface{}, error) {
	return nil, nil
}

func (l *localPubSub) Publish(topic string, data interface{}) error {
	return nil
}

func (l *localPubSub) Close() error {
	return nil
}
