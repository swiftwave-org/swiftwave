package main

import (
	SERVER "swiftwave/m/server"
)

func main() {
	server := SERVER.Server{}
	server.Init()
	server.Start()
}
