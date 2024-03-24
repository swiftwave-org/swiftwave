package task_queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
)

func (r *remoteTaskQueue) RegisterFunction(queueName string, function WorkerFunctionType) error {
	// acquire lock
	r.mutexQueueToFunctionMapping.Lock()
	// release lock when function returns
	defer r.mutexQueueToFunctionMapping.Unlock()
	// check if there is already a function registered for this queue
	if _, ok := r.queueToFunctionMapping[queueName]; ok {
		functionName := r.queueToFunctionMapping[queueName].functionName
		return errors.New("already a function [" + functionName + "] registered for this queue")
	}
	// inspect function
	metadata, err := inspectFunction(function)
	if err != nil {
		return err
	}
	// establish connection
	err = r.establishConnection()
	if err != nil {
		return err
	}
	// declare queue
	err = r.declareQueue(queueName)
	if err != nil {
		return err
	}
	// add function to mapping
	r.queueToFunctionMapping[queueName] = metadata
	return nil
}

func (r *remoteTaskQueue) EnqueueTask(queueName string, argument ArgumentType) error {
	// marshal argument to json
	jsonBytes, err := json.Marshal(argument)
	if err != nil {
		return errors.New("error while marshalling argument to json")
	}

	// check if queueName is registered
	r.mutexQueueToFunctionMapping.RLock()
	if _, ok := r.queueToFunctionMapping[queueName]; !ok {
		return errors.New("no function registered for this queue")
	}
	r.mutexQueueToFunctionMapping.RUnlock()

	// establish connection if not already established
	err = r.establishConnection()
	if err != nil {
		return errors.New("error while establishing connection to AMQP server")
	}

	// push to queue
	if r.queueType == AmqpQueue {
		dConfirmation, err := r.amqpChannel.PublishWithDeferredConfirmWithContext(
			context.Background(),
			"",
			queueName,
			true,
			false,
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				DeliveryMode:    amqp.Persistent,
				Priority:        0,
				Body:            jsonBytes,
			},
		)
		if err != nil {
			log.Println("error while publishing message to queue [" + queueName + "]")
			log.Println(err.Error())
			return errors.New("error while publishing message to queue")
		}
		// Check acknowledgement
		ack := dConfirmation.Wait()
		if !ack {
			log.Println("error while publishing message to queue [" + queueName + "] publish ack > false")
			return errors.New("error while publishing message to queue")
		}
	} else if r.queueType == RedisQueue {
		// push to redis
		_, err := r.redisClient.RPush(context.Background(), queueName, jsonBytes).Result()
		if err != nil {
			log.Println("error while pushing message to queue [" + queueName + "]")
			log.Println(err.Error())
			return errors.New("error while pushing message to queue")
		}
	}

	return nil
}

func (r *remoteTaskQueue) StartConsumers(nowait bool) error {
	// establish connection if not already established
	err := r.establishConnection()
	if err != nil {
		return err
	}

	// create the queue names for copy to a new slice
	queueNames := make([]string, 0, len(r.queueToFunctionMapping))

	// acquire lock
	r.mutexQueueToFunctionMapping.RLock()

	// copy the queue names
	for queueName := range r.queueToFunctionMapping {
		queueNames = append(queueNames, queueName)
	}

	// release lock when function returns
	r.mutexQueueToFunctionMapping.RUnlock()

	// wait group for consumers
	wg := r.consumersWaitGroup

	// start consumers
	for _, queueName := range queueNames {
		for i := 1; i <= r.NoOfWorkersPerQueue; i++ {
			wg.Add(1)
			go r.listenForTasks(queueName, wg)
		}
	}

	if !nowait {
		// wait for all consumers to finish
		wg.Wait()
	}

	return nil
}

func (r *remoteTaskQueue) WaitForConsumers() {
	// wait for all consumers to finish
	r.consumersWaitGroup.Wait()
}

func (r *remoteTaskQueue) EnqueueProcessingQueueExpiredTask() error {
	if r.queueType == AmqpQueue {
		return nil
	}
	for queueName := range r.queueToFunctionMapping {
		// move from processing queue to original queue
		_ = r.redisClient.LMove(context.Background(), queueName+"_processing", queueName, "right", "left")
	}
	return nil
}

// private functions
// getFunction: getFunction returns the function registered for a queue
func (r *remoteTaskQueue) getFunction(queueName string) (functionMetadata, error) {
	// acquire lock
	r.mutexQueueToFunctionMapping.RLock()
	// release lock when function returns
	defer r.mutexQueueToFunctionMapping.RUnlock()

	// check if there is no function registered for this queue
	if _, ok := r.queueToFunctionMapping[queueName]; !ok {
		return functionMetadata{}, errors.New("no function registered for this queue")
	}

	// return function
	return r.queueToFunctionMapping[queueName], nil
}

// establishConnection: connect connects to the AMQP server
func (r *remoteTaskQueue) establishConnection() error {
	if r.queueType == RedisQueue {
		return nil
	}
	// if there is already a connection, return
	if r.amqpConnection != nil && !r.amqpConnection.IsClosed() {
		return nil
	}
	// dial connection
	connection, err := amqp.DialConfig(r.amqpURI, r.amqpConfig)
	if err != nil {
		return err
	}
	// set connection
	r.amqpConnection = connection
	// get a channel from the connection
	channel, err := r.amqpConnection.Channel()
	if err != nil {
		return err
	}
	// set channel
	r.amqpChannel = channel
	err = r.amqpChannel.Confirm(false)
	if err != nil {
		return err
	}
	return nil
}

// declareQueue: create a queue
func (r *remoteTaskQueue) declareQueue(queueName string) error {
	if r.queueType == RedisQueue {
		return nil
	}
	if r.amqpConnection == nil || r.amqpChannel == nil {
		return errors.New("connection not established")
	}
	// create queue
	_, err := r.amqpChannel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	return err
}

// listenForTasks: listen for tasks on a queue
func (r *remoteTaskQueue) listenForTasks(queueName string, wg *sync.WaitGroup) {
	if r.queueType == RedisQueue {
		r.listenForTasksUsingRQ(queueName, wg)
	} else if r.queueType == AmqpQueue {
		r.listenForTasksUsingAMQP(queueName, wg)
	}
}

func (r *remoteTaskQueue) listenForTasksUsingRQ(queueName string, _ *sync.WaitGroup) {
	// fetch function by queue name
	functionMetadata, err := r.getFunction(queueName)
	if err != nil {
		log.Println("error while fetching function for queue [" + queueName + "]")
		log.Println("error: " + err.Error())
	}

	// log message
	log.Println("starting consumer for redis queue [" + queueName + "]")

	for {
		stringCmd := r.redisClient.BLMove(context.Background(), queueName, queueName+"_processing", "right", "left", 1*time.Minute)
		if stringCmd.Err() != nil {
			if strings.Contains(stringCmd.Err().Error(), "redis: nil") {
				continue
			}
			log.Println("error while fetching message from queue [" + queueName + "]")
			continue
		}
		content, err := stringCmd.Bytes()
		if err != nil {
			continue
		}

		// create a new object of an argument type
		argument := reflect.New(functionMetadata.argumentType).Interface()

		// string to json unmarshal
		err = json.Unmarshal(content, &argument)
		if err != nil {
			log.Println(err)
			continue
		}

		// argument is a pointer, dereference it
		argument = reflect.ValueOf(argument).Elem().Interface()

		// type cast to an argument type
		if err != nil {
			log.Println("error while de-referencing argument from pointer for queue [" + queueName + "]")
			log.Println("error: " + err.Error())
			continue
		}
		// execute function
		err = invokeFunction(functionMetadata.function, argument, functionMetadata.argumentType)
		// remove from processing queue
		r.redisClient.LRem(context.Background(), queueName+"_processing", 0, content)
		if err != nil {
			log.Println("error while executing function for queue [" + queueName + "]")
			log.Println("error: " + err.Error())
			// enqueue to original queue
			r.redisClient.LPush(context.Background(), queueName, content)
		}
	}
}

func (r *remoteTaskQueue) listenForTasksUsingAMQP(queueName string, wg *sync.WaitGroup) {
	// fetch function by queue name
	functionMetadata, err := r.getFunction(queueName)
	if err != nil {
		log.Println("error while fetching function for queue [" + queueName + "]")
		log.Println("error: " + err.Error())
	}

	// log message
	log.Println("starting consumer for amqp queue [" + queueName + "]")

	// consumer tag
	consumerTag := r.amqpClientName + "_" + queueName
	// start consumer
	deliveries, err := r.amqpChannel.Consume(
		queueName,   // name
		consumerTag, // consumerTag,
		false,       // autoAck
		false,       // exclusive
		false,       // noLocal
		false,       // noWait
		nil,         // arguments
	)
	if err != nil {
		println(err.Error())
		panic("error while listening for queue [" + queueName + "], maybe some connection error")
	}

	for {
		delivery, ok := <-deliveries
		if !ok {
			// Channel is closed, exit the loop
			break
		}

		// fetch the content
		content := delivery.Body

		// create a new object of an argument type
		argument := reflect.New(functionMetadata.argumentType).Interface()

		// string to json unmarshal
		err := json.Unmarshal(content, &argument)
		if err != nil {
			log.Println(err)
			ackMessage(delivery)
			continue
		}

		// argument is a pointer, dereference it
		argument = reflect.ValueOf(argument).Elem().Interface()

		// type cast to an argument type
		if err != nil {
			log.Println("error while de-referencing argument from pointer for queue [" + queueName + "]")
			log.Println("error: " + err.Error())
			// acknowledge message
			ackMessage(delivery)
			continue
		}
		// execute function
		err = invokeFunction(functionMetadata.function, argument, functionMetadata.argumentType)
		if err != nil {
			log.Println("error while executing function for queue [" + queueName + "]")
			log.Println("error: " + err.Error())
			nackMessage(delivery)
			continue
		}
		// acknowledge message
		ackMessage(delivery)
	}
	// wait group done
	wg.Done()
}

func ackMessage(delivery amqp.Delivery) {
	err := delivery.Ack(false)
	if err != nil {
		log.Println("error while acknowledging message for queue [" + delivery.RoutingKey + "]")
		log.Println("error: " + err.Error())
	}
}

func nackMessage(delivery amqp.Delivery) {
	err := delivery.Nack(false, true)
	if err != nil {
		log.Println("error while nacknowledging message for queue [" + delivery.RoutingKey + "]")
	}
}

func (r *remoteTaskQueue) PurgeQueue(queueName string) error {
	if r.queueType == RedisQueue {
		err := r.redisClient.Del(context.Background(), queueName).Err()
		if err != nil {
			return err
		}
		return r.redisClient.Del(context.Background(), queueName+"_processing").Err()
	}
	if r.queueType == AmqpQueue {
		if r.amqpChannel == nil {
			err := r.establishConnection()
			if err != nil {
				return fmt.Errorf("error while establishing connection to AMQP server: %s", err.Error())
			}
		}
		_, err := r.amqpChannel.QueuePurge(queueName, true)
		return err
	}
	return errors.New("invalid queue type")
}

func (r *remoteTaskQueue) ListMessages(queueName string) ([]string, error) {
	if r.queueType == RedisQueue {
		return r.inspectQueueUsingRQ(queueName)
	}
	if r.queueType == AmqpQueue {
		return r.inspectQueueUsingAMQP(queueName)
	}
	return nil, errors.New("invalid queue type")
}

func (r *remoteTaskQueue) inspectQueueUsingRQ(queueName string) ([]string, error) {
	// fetch all the messages from the queue
	result, err := r.redisClient.LRange(context.Background(), queueName, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *remoteTaskQueue) inspectQueueUsingAMQP(queueName string) ([]string, error) {
	if r.amqpChannel == nil {
		err := r.establishConnection()
		if err != nil {
			return nil, fmt.Errorf("error while establishing connection to AMQP server: %s", err.Error())
		}
	}
	// fetch all the messages from the queue, by consuming all the messages and then nack them
	consumerTag := r.amqpClientName + "_" + queueName
	deliveries, err := r.amqpChannel.Consume(
		queueName,   // name
		consumerTag, // consumerTag,
		false,       // autoAck
		false,       // exclusive
		false,       // noLocal
		true,        // noWait
		nil,         // arguments
	)
	if err != nil {
		return nil, err
	}
	var result []string
	var lastDelivery *amqp.Delivery
	var lastMessageTime = time.Now()
	var ticker = time.NewTicker(1 * time.Second)
	for {
		select {
		case delivery, ok := <-deliveries:
			if !ok {
				// Channel is closed, exit the loop
				break
			}
			// fetch the content
			content := delivery.Body
			result = append(result, string(content))
			lastDelivery = &delivery
			lastMessageTime = time.Now()
		case <-ticker.C:
			if time.Since(lastMessageTime) > 10*time.Second {
				if lastDelivery != nil {
					_ = lastDelivery.Nack(false, true)
				}
				return result, nil
			}
		}
	}
}
