package main

import (
	SERVER "swiftwave/m/server"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	server := SERVER.Server{}
	server.Init()
	server.Start()
}
