package main

import (
	"crypto/rsa"
	"log"

	"golang.org/x/net/websocket"
)

const (
	privKeyFile = "privatekey.pem"
)

type Client struct {
	sock        *websocket.Conn
	privateKey  *rsa.PrivateKey
	authKey     []byte
	chatSession *ChatSession
	gui         *GUI
}

// Connect connects to the websocket server
func (c *Client) Connect(server string) bool {
	if c.sock == nil {
		ws, err := websocket.Dial(server, "", "http://")
		if err != nil {
			c.gui.ShowDialog("Error connecting to server")
			return false
		}
		c.sock = ws
	}
	return true
}

func main() {
	c := &Client{}
	guiConfig := &GUIConfig{
		DefaultServerText:     "ws://localhost:5000",
		ChatRoomsPollInterval: 5,
		CreateUserHandler:     c.createUserHandler,
		LoginUserHandler:      c.loginUserHandler,
		CreateRoomHandler:     c.createRoomHandler,
		JoinChatHandler:       c.joinChatHandler}

	c.gui = NewGUI(guiConfig)

	// Enter GUI event loop
	if err := c.gui.app.Run(); err != nil {
		log.Fatal(err)
	}
}
