package main

import (
	"crypto/rsa"
	"io/ioutil"
	"log"
	"os"

	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"golang.org/x/net/websocket"
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

func savePrivKey(username string, privKey *rsa.PrivateKey) {
	pem := util.MarshalPrivate(privKey)
	if err := ioutil.WriteFile(username+".pem", pem, 0644); err != nil {
		log.Fatal(err)
	}
}

func main() {
	f, _ := os.OpenFile("client_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)

	c := &Client{}
	guiConfig := &GUIConfig{
		DefaultServerText:     "ws://localhost:5000",
		ChatRoomsPollInterval: 2,
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
