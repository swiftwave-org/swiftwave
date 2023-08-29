package main

import (
	"log"
	SERVER "swiftwave/m/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printl("Error loading .env file. Ignoring")
	}
	server := SERVER.Server{}
	server.Init()
	server.Start()
}
