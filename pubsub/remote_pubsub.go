package pubsub

func (r *remotePubSub) CreateTopic(topic string) error {
	return nil
}

func (r *remotePubSub) RemoveTopic(topic string) error {
	return nil
}

func (r *remotePubSub) Subscribe(topic string) (<-chan interface{}, error) {
	return nil, nil
}

func (r *remotePubSub) Publish(topic string, data interface{}) error {
	return nil
}

func (r *remotePubSub) Close() error {
	return nil
}
