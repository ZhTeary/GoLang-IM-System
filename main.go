package main

import "Golang-IM-System/src"

func main() {
	server := src.NewServer("127.0.0.1", 8888)
	server.Start()
}
