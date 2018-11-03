package websock

import (
	"log"

	"golang.org/x/net/websocket"
)

func ServerHandler(ws *websocket.Conn) {
	msg := new(Message)

	// Listen for messages
	for {
		if err := websocket.JSON.Receive(ws, msg); err != nil {
			log.Println(err)
			break
		}

		log.Printf("Recieved message: %v, from client %s", msg, ws.RemoteAddr())

		res := &Message{0, "Thanks"}
		if err := websocket.JSON.Send(ws, res); err != nil {
			log.Println(err)
			break
		}
	}
}

func handleMessage(ws *websocket.Conn, msg *Message) {

}
