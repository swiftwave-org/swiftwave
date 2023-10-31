package task_queue

import "sync"

type WorkerFunctionType interface{}
type ArgumentType interface{}

type Client interface {
	RegisterFunction(queueName string, function WorkerFunctionType) error
	EnqueueTask(queueName string, argument ArgumentType) error
}

type localTaskQueue struct {
	mutex                  *sync.RWMutex
	queueToFunctionMapping map[string]functionMetadata // map between queue name <---> function
}

type functionMetadata struct {
	function         WorkerFunctionType
	functionName     string
	argumentType     ArgumentType
	argumentTypeName string
}

type Type string

const (
	Local  Type = "local"
	Remote Type = "remote"
)

type Options struct {
	Type Type
}
