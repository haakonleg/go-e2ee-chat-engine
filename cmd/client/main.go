package main

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/haakonleg/go-e2ee-chat-engine/websock"

	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"golang.org/x/net/websocket"
)

// Result is used by WSReader to communicate the websocket messages between threads
type Result struct {
	Message *websock.Message
	Err     error
}

// WSReader reads messages from the websocket in the background
type WSReader struct {
	OnDisconnect func()
	Ws           *websocket.Conn
	c            chan Result
}

// Reader runs in a separate goroutine and listens for messages on the websocket
func (wr *WSReader) Reader() {
	for {
		msg := new(websock.Message)
		if err := websocket.JSON.Receive(wr.Ws, msg); err != nil {
			break
		} else if msg.Type == websock.Error {
			wr.c <- Result{Message: nil, Err: errors.New(string(msg.Message))}
		} else {
			wr.c <- Result{Message: msg, Err: nil}
		}
	}

	wr.OnDisconnect()
}

// GetNext retrieves the next websocket message from the message pool
func (wr *WSReader) GetNext() (*websock.Message, error) {
	result := <-wr.c
	return result.Message, result.Err
}

type Client struct {
	wsReader    *WSReader
	ws          *websocket.Conn
	privateKey  *rsa.PrivateKey
	authKey     []byte
	chatSession *ChatSession
	gui         *GUI
}

func (c *Client) Disconnected() {
	c.gui.ShowDialog("Disconnected from server", func() {
		c.gui.app.Stop()
	})
}

// Connect connects to the websocket server
func (c *Client) Connect(server string) bool {
	if c.wsReader == nil {
		ws, err := websocket.Dial(server, "", "http://")
		if err != nil {
			c.gui.ShowDialog("Error connecting to server", nil)
			return false
		}

		c.ws = ws
		c.wsReader = &WSReader{
			OnDisconnect: c.Disconnected,
			Ws:           ws,
			c:            make(chan Result, 10)}
		go c.wsReader.Reader()
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
		DefaultServerText:     "wss://go-e2ee-chat-engine.herokuapp.com/",
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
