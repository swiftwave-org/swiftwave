package task_queue

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

func NewClient(options Options) (Client, error) {
	if options.Type == Local {
		return createLocalTaskQueueClient(options)
	} else if options.Type == Remote {
		return createRemoteTaskQueueClient(options)
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

	return &localTaskQueue{
		db:                          options.DbClient,
		mutexQueueToFunctionMapping: mutex,
		mutexQueueToChannelMapping:  mutex2,
		queueToFunctionMapping:      functionsMapping,
		queueToChannelMapping:       channelsMapping,
		maxMessagesPerQueue:         options.MaxMessagesPerQueue,
		NoOfWorkersPerQueue:         options.NoOfWorkersPerQueue,
		consumersWaitGroup:          &sync.WaitGroup{},
	}, nil
}

func createRemoteTaskQueueClient(options Options) (Client, error) {
	functionsMapping := make(map[string]functionMetadata)
	mutex := &sync.RWMutex{}

	if options.RemoteQueueType == NoneRemoteQueue {
		return nil, errors.New("remote queue type is not provided")
	}

	if options.RemoteQueueType == AmqpQueue {
		// declare connection
		amqpConfig := amqp.Config{
			Vhost:      options.AMQPVhost,
			Properties: amqp.NewConnectionProperties(),
		}
		// set client name
		amqpConfig.Properties.SetClientConnectionName(options.AMQPClientName)
		return &remoteTaskQueue{
			queueType:                   AmqpQueue,
			mutexQueueToFunctionMapping: mutex,
			NoOfWorkersPerQueue:         options.NoOfWorkersPerQueue,
			queueToFunctionMapping:      functionsMapping,
			amqpURI:                     options.AMQPUri,
			amqpConfig:                  amqpConfig,
			amqpClientName:              options.AMQPClientName,
			consumersWaitGroup:          &sync.WaitGroup{},
		}, nil
	} else if options.RemoteQueueType == RedisQueue {
		if options.RedisClient == nil {
			return nil, errors.New("redis client is nil")
		}
		return &remoteTaskQueue{
			queueType:                   RedisQueue,
			mutexQueueToFunctionMapping: mutex,
			NoOfWorkersPerQueue:         options.NoOfWorkersPerQueue,
			queueToFunctionMapping:      functionsMapping,
			redisClient:                 options.RedisClient,
			consumersWaitGroup:          &sync.WaitGroup{},
		}, nil

	} else {
		return nil, errors.New("invalid remote queue type")
	}
}
