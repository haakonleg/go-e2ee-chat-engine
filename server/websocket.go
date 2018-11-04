package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/haakonleg/go-e2ee-chat-engine/user"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

// WebsockHandler is the handler for the server websocket
// it handles messages from a single client
func (s *Server) WebsockHandler(ws *websocket.Conn) {
	s.AddClient(ws, nil)
	log.Printf("Client connected: %s. Total connected: %d", ws.Request().RemoteAddr, len(s.ConnectedClients))

	// Listen for messages
	for {
		msg := new(websock.Message)
		if err := websocket.JSON.Receive(ws, msg); err != nil {
			break
		}

		// Check message type and forward to appropriate handlers
		switch msg.Type {
		case websock.RegisterUser:
			s.registerUser(ws, msg)
		case websock.LoginUser:
			s.loginUser(ws, msg)
		default:
			invalidMessage(ws)
		}
	}

	s.RemoveClient(ws)
	log.Printf("Client disconnected: %s. Total connected: %d\n", ws.Request().RemoteAddr, len(s.ConnectedClients))
}

func (s *Server) registerUser(ws *websocket.Conn, msg *websock.Message) {
	regUserMsg := new(websock.RegisterUserMessage)
	if err := json.Unmarshal(msg.Message, regUserMsg); err != nil {
		invalidFormat(ws)
		return
	}

	if err := user.RegisterUser(s.Db, regUserMsg.Username, regUserMsg.PublicKey); err != nil {
		websock.SendMessage(ws, websock.Error, err.Error(), websock.String)
		return
	}
	websock.SendMessage(ws, websock.MessageOK, "User registered", websock.String)
}

func (s *Server) loginUser(ws *websocket.Conn, msg *websock.Message) {
	username := string(msg.Message)

	// Generate user object
	user, err := user.AuthChallenge(s.Db, username)
	if err != nil {
		websock.SendMessage(ws, websock.Error, err.Error(), websock.String)
	}

	// Send auth challenge
	if err := websock.SendMessage(ws, websock.AuthChallenge, user.EncKey, websock.Bytes); err != nil {
		log.Println(err)
		return
	}

	// Recieve auth challenge
	res := new(websock.Message)
	if err := websocket.JSON.Receive(ws, res); err != nil {
		log.Println(err)
		return
	}

	// Check that the recieved decrypted key matches the original auth key
	if bytes.Compare(res.Message, user.AuthKey) == 0 {
		fmt.Printf("Client %s authenticated as user %s\n", ws.Request().RemoteAddr, user.Username)
		s.AddClient(ws, user)
		websock.SendMessage(ws, websock.MessageOK, "Logged in", websock.String)
	} else {
		websock.SendMessage(ws, websock.Error, "Invalid auth key", websock.String)
	}
}

func invalidMessage(ws *websocket.Conn) {
	log.Println("Error: Invalid websocket message type")
	websock.SendMessage(ws, websock.Error, "Invalid webscoket message type", websock.String)
}

func invalidFormat(ws *websocket.Conn) {
	log.Println("Error: invalid message format")
	websock.SendMessage(ws, websock.Error, "Invalid message format", websock.String)
}
