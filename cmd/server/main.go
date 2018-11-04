package main

import (
	"log"
	"os"

	"github.com/haakonleg/go-e2ee-chat-engine/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Error: environment variable PORT is not set")
	}

	serverConfig := server.Config{
		ListenPort: port,
		DBName:     "go-e2ee-chat-engine",
		MongoURL:   "mongo:27017"}

	server := server.CreateServer(serverConfig)
	server.Start()
}
