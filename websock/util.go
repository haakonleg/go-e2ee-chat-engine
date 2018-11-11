package websock

import (
	"encoding/json"
	"errors"
	"log"

	"golang.org/x/net/websocket"
)

// SendMessage sends a message to the websocket receiver
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
// an error is returned if the response is of type websock.Error (means server returned an error)
func GetResponse(ws *websocket.Conn) (*Message, error) {
	response := new(Message)
	if err := websocket.JSON.Receive(ws, response); err != nil {
		return nil, err
	}

	if response.Type == Error {
		return nil, errors.New(string(response.Message))
	}

	return response, nil
}

func InvalidMessage(ws *websocket.Conn) {
	log.Printf("%s sent invalid websocket message type\n", ws.Request().RemoteAddr)
	SendMessage(ws, Error, "Invalid webscoket message type", String)
}

func InvalidFormat(ws *websocket.Conn) {
	log.Printf("%s sent invalid message format\n", ws.Request().RemoteAddr)
	SendMessage(ws, Error, "Invalid message format", String)
}

func InvalidAuth(ws *websocket.Conn) {
	log.Printf("%s sent invalid auth token\n", ws.Request().RemoteAddr)
	SendMessage(ws, Error, "Invalid authentication token", String)
}
