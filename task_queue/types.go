package task_queue

import "sync"

type WorkerFunctionType interface{}
type ArgumentType interface{}

type Client interface {
	// RegisterFunction registers a consumer function for a queue
	RegisterFunction(queueName string, function WorkerFunctionType) error
	// EnqueueTask enqueues a task to a queue
	EnqueueTask(queueName string, argument ArgumentType) error
	// StartConsumers is a blocking function that starts the consumers for all the registered queues
	StartConsumers()
}

type localTaskQueue struct {
	mutexQueueToFunctionMapping *sync.RWMutex
	mutexQueueToChannelMapping  *sync.RWMutex
	queueToFunctionMapping      map[string]functionMetadata // map between queue name <---> function
	queueToChannelMapping       map[string]chan ArgumentType
	operationMode               Mode
	maxMessagesPerQueue         int
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

type Mode string

const (
	ProducerOnly Mode = "producer_only"
	ConsumerOnly Mode = "consumer_only"
	Both         Mode = "both"
)

type Options struct {
	Type                Type
	Mode                Mode
	MaxMessagesPerQueue int
}
