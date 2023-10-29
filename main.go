package main

import (
	"fmt"
	"github.com/swiftwave-org/swiftwave/pubsub"
	"log"
	"sync"
)

//func main() {
//	err := godotenv.Load()
//	if err != nil {
//		log.Println("WARN: error loading .env file. Ignoring")
//	}
//	// Load the manager
//	config := core.ServiceConfig{}
//	manager := core.ServiceManager{}
//	config.Load()
//	manager.Load()
//
//	// Create Echo Server
//	echoServer := echo.New()
//	// Start the swift wave server
//	swiftwave.StartServer(&config, &manager, echoServer, true)
//}

var wg sync.WaitGroup

func producer(c pubsub.Client, topic string) {
	for i := 0; i < 100; i++ {
		msg := fmt.Sprintf("Message %d", i)
		fmt.Printf("Publishing %s\n", msg)
		err := c.Publish(topic, msg)
		if err != nil {
			panic(err)
		}
	}
	err := c.RemoveTopic(topic)
	if err != nil {
		fmt.Println(err)
	}
	wg.Done()
	fmt.Println("Done publishing")
}

func consumer(prefixLog string, channel <-chan string) {
	//time.Sleep(10 * time.Second)
	// check if channel is closed
	for {
		msg, ok := <-channel
		if !ok {
			log.Println("Channel closed on consumer : " + prefixLog)
			break
		}
		log.Println("Received", msg, " on consumer ", prefixLog)
	}
	wg.Done()
	fmt.Println("Done consuming on consumer : ", prefixLog)
}

func main() {
	pubsubclient, err := pubsub.CreatePubSubClient(pubsub.Config{
		Type:         pubsub.Local,
		BufferLength: 1000,
		RedisClient:  nil,
	})
	if err != nil {
		panic(err)
	}
	log.Println("Created pubsub client", pubsubclient)
	// Create a topic
	err = pubsubclient.CreateTopic("test")
	if err != nil {
		panic(err)
	}
	wg.Add(3)

	// Create a subscription
	subscriptionId, subscriptionChannel, err := pubsubclient.Subscribe("test")
	if err != nil {
		panic(err)
	}
	log.Println("Created subscription", subscriptionId, subscriptionChannel)
	go consumer("A", subscriptionChannel)

	// Create a second subscription
	subscriptionId, subscriptionChannel, err = pubsubclient.Subscribe("test")
	if err != nil {
		panic(err)
	}
	log.Println("Created subscription", subscriptionId, subscriptionChannel)

	go consumer("B", subscriptionChannel)

	// Run producer
	go producer(pubsubclient, "test")
	// Wait for goroutines to finish
	wg.Wait()
}
