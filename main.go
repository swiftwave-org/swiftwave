package main

import (
	"github.com/joho/godotenv"
	"log"
	SERVER "swiftwave/m/server"
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
