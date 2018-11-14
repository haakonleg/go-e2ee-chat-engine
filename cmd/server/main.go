package main

import (
	"log"
	"os"

	"github.com/haakonleg/go-e2ee-chat-engine/server"
	"golang.org/x/net/websocket"
	"net/http"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Error: environment variable PORT is not set")
	}
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("Error: environment variable MONGODB_URI is not set")
	}
	dbName := os.Getenv("MONGODB_NAME")
	if dbName == "" {
		log.Fatal("Error: environment variable MONGODB_NAME is not set")
	}

	serverConfig := server.Config{
		DBName:   dbName,
		MongoURL: mongoURI}

	server := server.CreateServer(serverConfig)

	log.Printf("Listening on port: %s\n", port)

	err := http.ListenAndServe(":"+port, websocket.Handler(server.WebsockHandler))
	log.Printf("Error occured in http listener: %s\n", err)
}
