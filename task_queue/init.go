package task_queue

import (
	"errors"
	"sync"
)

func NewClient(options Options) (Client, error) {
	if options.Type == Local {
		return createLocalTaskQueueClient(options)
		//} else if options.Type == Remote {
		//	return createRemoteTaskQueueClient(options)
	} else {
		return nil, errors.New("invalid task queue type")
	}
}

func createLocalTaskQueueClient(options Options) (Client, error) {
	if options.MaxMessagesPerQueue == 0 {
		return nil, errors.New("max messages per queue cannot be zero")
	}
	functionsMapping := make(map[string]functionMetadata)
	channelsMapping := make(map[string]chan ArgumentType)
	mutex := &sync.RWMutex{}
	mutex2 := &sync.RWMutex{}

	if options.Type == Local && options.Mode != Both {
		return nil, errors.New("for local task queue, mode should be both")
	}

	return &localTaskQueue{
		mutexQueueToFunctionMapping: mutex,
		mutexQueueToChannelMapping:  mutex2,
		queueToFunctionMapping:      functionsMapping,
		queueToChannelMapping:       channelsMapping,
		operationMode:               options.Mode,
		maxMessagesPerQueue:         options.MaxMessagesPerQueue,
	}, nil
}
