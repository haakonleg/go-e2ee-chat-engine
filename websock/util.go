package websock

import (
	"encoding/json"

	"golang.org/x/net/websocket"
)

// SendMessage sends a message to the websocket reciever
func SendMessage(ws *websocket.Conn, msgType MessageType, msg interface{}, format MessageFormat) error {
	var msgData []byte
	var err error

	switch format {
	case JSON:
		msgData, err = json.Marshal(msg)
	case String:
		msgData = []byte(msg.(string))
	case Bytes:
		msgData = msg.([]byte)
	}

	if err != nil {
		return err
	}

	wMsg := &Message{
		Type:    msgType,
		Message: msgData}

	if err := websocket.JSON.Send(ws, wMsg); err != nil {
		return err
	}

	return nil
}

// GetResponse waits for a response fom the websocket client/server
func GetResponse(ws *websocket.Conn) (*Message, error) {
	response := new(Message)
	if err := websocket.JSON.Receive(ws, response); err != nil {
		return nil, err
	}

	return response, nil
}
