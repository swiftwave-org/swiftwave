package main

import (
	SERVER "keroku/m/server"
)


func main() {
	server := SERVER.Server{}
	server.Init(3333)
	server.Start()
}
