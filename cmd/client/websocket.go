package main

import (
	"log"
	"os"

	"golang.org/x/net/websocket"
)

// NewClient creates a new websocket client
func NewClient(server string) (*websocket.Conn, error) {
	origin, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	ws, err := websocket.Dial(server, "", "http://"+origin)
	if err != nil {
		return nil, err
	}

	return ws, nil
}
