package main

import (
	"log"
	"net/http"
	"os"

	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Error: environment variable PORT is not set")
	}

	// Connect to the database
	db := &mdb.Database{
		DBName:   "go-e2ee-chat-engine",
		MongoURL: "mongo:27017"}

	if err := db.CreateConnection(); err != nil {
		log.Fatal(err)
	}

	// Listen for websocket connections
	log.Println("Listening for connections...")
	http.ListenAndServe(":"+port, websocket.Handler(websock.ServerHandler))
}
