package task_queue

import (
	"github.com/go-redis/redis/v8"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
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
	// EnqueueProcessingQueueExpiredTask enqueues tasks from processing queue to the original queue
	EnqueueProcessingQueueExpiredTask() error
	// PurgeQueue purges all the messages from a queue
	PurgeQueue(queueName string) error
	// ListMessages returns the messages of a queue
	// Note: Should be called when no consumers are running
	ListMessages(queueName string) ([]string, error)
}

type localTaskQueue struct {
	mutexQueueToFunctionMapping *sync.RWMutex
	mutexQueueToChannelMapping  *sync.RWMutex
	queueToFunctionMapping      map[string]functionMetadata // map between queue name <---> function
	queueToChannelMapping       map[string]chan ArgumentType
	maxMessagesPerQueue         int
	NoOfWorkersPerQueue         int
	consumersWaitGroup          *sync.WaitGroup
	db                          *gorm.DB
}

type RemoteQueueType string

const (
	AmqpQueue       RemoteQueueType = "amqp"
	RedisQueue      RemoteQueueType = "redis"
	NoneRemoteQueue RemoteQueueType = "none"
)

type remoteTaskQueue struct {
	mutexQueueToFunctionMapping *sync.RWMutex
	queueToFunctionMapping      map[string]functionMetadata // map between queue name <---> function
	consumersWaitGroup          *sync.WaitGroup
	NoOfWorkersPerQueue         int
	queueType                   RemoteQueueType
	// redis specific
	redisClient *redis.Client
	// amqp specific
	amqpConfig     amqp.Config
	amqpURI        string
	amqpClientName string
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

type Options struct {
	Type                ServiceType
	NoOfWorkersPerQueue int
	MaxMessagesPerQueue int      // only applicable for local task queue
	DbClient            *gorm.DB // only applicable for local task queue
	// Extra options for remote task queue
	RemoteQueueType RemoteQueueType
	// Redis specific options
	RedisClient *redis.Client
	// AMQP specific options
	AMQPUri        string
	AMQPVhost      string
	AMQPClientName string
}
