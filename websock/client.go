package websock

import (
	"log"
	"os"

	"golang.org/x/net/websocket"
)

func NewClient(server string) (*websocket.Conn, error) {
	origin, _ := os.Hostname()

	ws, err := websocket.Dial(server, "", "http://"+origin)
	if err != nil {
		log.Fatal(err)
	}
	return ws, nil
}
