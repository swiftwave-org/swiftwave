package main

import (
	"github.com/joho/godotenv"
	"log"
	SERVER "github.com/swiftwave-org/swiftwave/server"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file. Ignoring")
	}
	server := SERVER.Server{}
	server.Init()
	server.Start()
}
