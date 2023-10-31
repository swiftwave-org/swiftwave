package task_queue

import (
	"errors"
	"reflect"
)

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
	return nil
}

func (l *localTaskQueue) StartConsumers() error {
	// TODO implement this
	return nil
}

// private function
func (l *localTaskQueue) getFunction(queueName string) (functionMetadata, error) {
	// acquire lock
	l.mutex.RLock()
	// release lock when function returns
	defer l.mutex.RUnlock()

	// check if there is no function registered for this queue
	if _, ok := l.queueToFunctionMapping[queueName]; !ok {
		return functionMetadata{}, errors.New("no function registered for this queue")
	}

	// return function
	return l.queueToFunctionMapping[queueName], nil
}

func getTypeName(object interface{}) string {
	val := reflect.ValueOf(object)
	if val.Kind() == reflect.Ptr {
		return val.Elem().Type().Name()
	} else {
		return val.Type().Name()
	}
}
