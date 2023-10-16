package main

import (
	"log"

	"github.com/joho/godotenv"
	SERVER "github.com/swiftwave-org/swiftwave/server"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file. Ignoring")
	}
	server := SERVER.Server{}
	server.Init()
	data, err := server.DOCKER_MANAGER.RealtimeInfoRunningServices()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(data)
	// server.Start()
}
