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
	mappings := make(map[string]functionMetadata)
	mutex := &sync.RWMutex{}
	return &localTaskQueue{
		mutex:                  mutex,
		queueToFunctionMapping: mappings,
	}, nil
}
