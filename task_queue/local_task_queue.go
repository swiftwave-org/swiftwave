package task_queue

import "errors"

func (l *localTaskQueue) RegisterFunction(queueName string, function WorkerFunctionType) error {
	// acquire lock
	l.mutex.Lock()
	// release lock when function returns
	defer l.mutex.Unlock()

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

	return nil
}

// private function
func (l *localTaskQueue) getFunction(queueName string) (WorkerFunctionType, error) {
	// acquire lock
	l.mutex.RLock()
	// release lock when function returns
	defer l.mutex.RUnlock()

	// check if there is no function registered for this queue
	if _, ok := l.queueToFunctionMapping[queueName]; !ok {
		return nil, errors.New("no function registered for this queue")
	}

	// return function
	return l.queueToFunctionMapping[queueName].function, nil
}
