package main

import (
	SERVER "swiftwave/m/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	server := SERVER.Server{}
	server.Init()
	server.Start()
}
