package task_queue

import "sync"

type WorkerFunctionType interface{}

type Client interface {
	RegisterFunction(name string, function WorkerFunctionType) error
}

type localTaskQueue struct {
	mutex                  *sync.RWMutex
	queueToFunctionMapping map[string]functionMetadata // map between queue name <---> function
}

type functionMetadata struct {
	function         WorkerFunctionType
	functionName     string
	argumentType     interface{}
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
