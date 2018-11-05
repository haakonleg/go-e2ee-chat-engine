package websock

import (
	"encoding/json"
	"log"

	"golang.org/x/net/websocket"
)

// SendMessage sends a message to the websocket reciever
func SendMessage(ws *websocket.Conn, msgType MessageType, msg interface{}, format MessageFormat) {
	var msgData []byte

	switch format {
	case JSON:
		msgData, _ = json.Marshal(msg)
	case String:
		msgData = []byte(msg.(string))
	case Bytes:
		msgData = msg.([]byte)
	}

	wMsg := &Message{
		Type:    msgType,
		Message: msgData}

	if err := websocket.JSON.Send(ws, wMsg); err != nil {
		log.Println(err)
	}
}

// GetResponse waits for a response fom the websocket client/server
func GetResponse(ws *websocket.Conn) (*Message, error) {
	response := new(Message)
	if err := websocket.JSON.Receive(ws, response); err != nil {
		return nil, err
	}

	return response, nil
}

func InvalidMessage(ws *websocket.Conn) {
	log.Println("Error: Invalid websocket message type")
	SendMessage(ws, Error, "Invalid webscoket message type", String)
}

func InvalidFormat(ws *websocket.Conn) {
	log.Println("Error: invalid message format")
	SendMessage(ws, Error, "Invalid message format", String)
}
