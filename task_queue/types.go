package task_queue

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"reflect"
	"sync"
)

type WorkerFunctionType interface{}
type ArgumentType interface{}

type Client interface {
	// RegisterFunction registers a consumer function for a queue
	RegisterFunction(queueName string, function WorkerFunctionType) error
	// EnqueueTask enqueues a task to a queue
	EnqueueTask(queueName string, argument ArgumentType) error
	// StartConsumers is a blocking function that starts the consumers for all the registered queues
	StartConsumers(nowait bool) error
	// WaitForConsumers is a blocking function that waits for all the consumers to finish
	WaitForConsumers()
}

type localTaskQueue struct {
	mutexQueueToFunctionMapping *sync.RWMutex
	mutexQueueToChannelMapping  *sync.RWMutex
	queueToFunctionMapping      map[string]functionMetadata // map between queue name <---> function
	queueToChannelMapping       map[string]chan ArgumentType
	operationMode               Mode
	maxMessagesPerQueue         int
	consumersWaitGroup          *sync.WaitGroup
}

type remoteTaskQueue struct {
	mutexQueueToFunctionMapping *sync.RWMutex
	queueToFunctionMapping      map[string]functionMetadata // map between queue name <---> function
	amqpConfig                  amqp.Config
	amqpURI                     string
	amqpClientName              string
	operationMode               Mode
	consumersWaitGroup          *sync.WaitGroup
	// internal use
	amqpConnection *amqp.Connection
	amqpChannel    *amqp.Channel
}

type functionMetadata struct {
	function         WorkerFunctionType
	functionName     string
	argumentType     reflect.Type
	argumentTypeName string
}

type ServiceType string

const (
	Local  ServiceType = "local"
	Remote ServiceType = "remote"
)

type Mode string

const (
	ProducerOnly Mode = "producer_only"
	ConsumerOnly Mode = "consumer_only"
	Both         Mode = "both"
)

type Options struct {
	Type                ServiceType
	Mode                Mode
	MaxMessagesPerQueue int // only applicable for local task queue
	// Extra options for remote task queue
	AMQPUri        string
	AMQPVhost      string
	AMQPClientName string
}
