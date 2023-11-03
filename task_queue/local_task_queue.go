package task_queue

import (
	"errors"
	"log"
	"sync"
)

func (l *localTaskQueue) RegisterFunction(queueName string, function WorkerFunctionType) error {
	// acquire lock
	l.mutexQueueToFunctionMapping.Lock()
	// release lock when function returns
	defer l.mutexQueueToFunctionMapping.Unlock()

	if l.operationMode == ConsumerOnly || l.operationMode == Both {
		// acquire lock for queue to channel mapping
		l.mutexQueueToChannelMapping.Lock()
		// release lock when function returns
		defer l.mutexQueueToChannelMapping.Unlock()
	}

	// check if there is already a function registered for this queue
	if _, ok := l.queueToFunctionMapping[queueName]; ok {
		functionName := l.queueToFunctionMapping[queueName].functionName
		return errors.New("already a function [" + functionName + "] registered for this queue")
	}

	// inspect function
	metadata, err := inspectFunction(function)
	if err != nil {
		return err
	}

	// add function to mapping
	l.queueToFunctionMapping[queueName] = metadata

	// add channel to mapping
	if l.operationMode == ConsumerOnly || l.operationMode == Both {
		l.queueToChannelMapping[queueName] = make(chan ArgumentType, l.maxMessagesPerQueue)
	}

	return nil
}

func (l *localTaskQueue) EnqueueTask(queueName string, argument ArgumentType) error {
	// fetch function by queue name
	functionMetadata, err := l.getFunction(queueName)
	if err != nil {
		return err
	}
	// verify the argument type
	if functionMetadata.argumentTypeName != getTypeName(argument) {
		return errors.New("invalid argument type for this queue, expected [" + functionMetadata.argumentTypeName + "]")
	}

	// enqueue task
	if l.operationMode == ConsumerOnly || l.operationMode == Both {
		// check if channel is full
		if len(l.queueToChannelMapping[queueName]) == l.maxMessagesPerQueue {
			return errors.New("queue is full, cannot enqueue task")
		}
		l.queueToChannelMapping[queueName] <- argument
	} else {
		return errors.New("cannot enqueue task in producer only mode")
	}

	return nil
}

func (l *localTaskQueue) StartConsumers(nowait bool) {
	if l.operationMode == ProducerOnly {
		panic("cannot start consumers in producer mode")
	}
	// copy the queue names to a new slice
	queueNames := make([]string, 0, len(l.queueToChannelMapping))

	// acquire lock
	l.mutexQueueToChannelMapping.RLock()

	// copy the queue names
	for queueName, _ := range l.queueToChannelMapping {
		queueNames = append(queueNames, queueName)
	}

	// release lock when function returns
	l.mutexQueueToChannelMapping.RUnlock()

	// wait group
	wg := l.consumersWaitGroup

	// start consumers
	for _, queueName := range queueNames {
		wg.Add(1)
		go l.listenForTasks(queueName, wg)
	}

	if !nowait {
		// wait for all consumers to finish
		wg.Wait()
	}
}

func (l *localTaskQueue) WaitForConsumers() {
	l.consumersWaitGroup.Wait()
}

// private function
func (l *localTaskQueue) getFunction(queueName string) (functionMetadata, error) {
	// acquire lock
	l.mutexQueueToFunctionMapping.RLock()
	// release lock when function returns
	defer l.mutexQueueToFunctionMapping.RUnlock()

	// check if there is no function registered for this queue
	if _, ok := l.queueToFunctionMapping[queueName]; !ok {
		return functionMetadata{}, errors.New("no function registered for this queue")
	}

	// return function
	return l.queueToFunctionMapping[queueName], nil
}

func (l *localTaskQueue) getChannel(queueName string) (<-chan ArgumentType, error) {
	// acquire lock
	l.mutexQueueToChannelMapping.RLock()
	// release lock when function returns
	defer l.mutexQueueToChannelMapping.RUnlock()

	// check if there is no channel registered for this queue
	if _, ok := l.queueToChannelMapping[queueName]; !ok {
		return nil, errors.New("no channel registered for this queue")
	}

	// return channel
	return l.queueToChannelMapping[queueName], nil
}

func (l *localTaskQueue) listenForTasks(queueName string, wg *sync.WaitGroup) {
	// fetch function by queue name
	functionMetadata, err := l.getFunction(queueName)
	if err != nil {
		log.Println("error while fetching function for queue [" + queueName + "]")
		log.Println("error: " + err.Error())
	}

	// fetch channel by queue name
	channel, err := l.getChannel(queueName)
	if err != nil {
		log.Println("error while fetching channel for queue [" + queueName + "]")
		log.Println("error: " + err.Error())
	}

	// log message
	log.Println("starting consumer for queue [" + queueName + "]")

	// start consumer
	for {
		argument, ok := <-channel
		if !ok {
			// Channel is closed, exit the loop
			break
		}

		err := invokeFunction(functionMetadata.function, argument, functionMetadata.argumentType)
		if err != nil {
			// Enqueue the task again in case of error
			err = l.EnqueueTask(queueName, argument)
			if err != nil {
				log.Println("error while enqueuing task again for queue [" + queueName + "]")
				log.Println("error: " + err.Error())
			}
		}
	}

	// decrement wait group counter
	wg.Done()
}
