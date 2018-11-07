package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"log"

	"github.com/haakonleg/go-e2ee-chat-engine/util"

	"github.com/haakonleg/go-e2ee-chat-engine/websock"

	"golang.org/x/net/websocket"
)

type ChatSession struct {
	Socket     *websocket.Conn
	AuthKey    []byte
	PrivateKey *rsa.PrivateKey
	Users      map[string]*websock.User
}

// NewChatSession creates a new instance of ChatSession
func NewChatSession(ws *websocket.Conn, authKey []byte, privateKey *rsa.PrivateKey) *ChatSession {
	return &ChatSession{
		Socket:     ws,
		AuthKey:    authKey,
		PrivateKey: privateKey,
		Users:      make(map[string]*websock.User, 0)}
}

// ChatSession runs in a separate goroutine and listens for new chat messages and users when a user is in a chat session
// TODO: Maybe move the callbacks into the struct, to make it more consistent with rest of the code
func (cs *ChatSession) ChatSession(
	onChatInfo func(error, *ChatSession, *websock.ChatInfoMessage),
	onChatMessage func(error, *ChatSession, *websock.ChatMessage),
	onUserJoined func(error, *ChatSession, *websock.User),
	onUserLeft func(*ChatSession, string)) {

	for {
		msg, err := websock.GetResponse(cs.Socket)
		if err != nil {
			log.Println(err)
			break
		}

		switch msg.Type {
		case websock.ChatInfo:
			chatInfo := new(websock.ChatInfoMessage)
			if err = json.Unmarshal(msg.Message, chatInfo); err == nil {
				// Decrypt chat messages
				err = cs.DecryptChatMessages(chatInfo.Messages...)

				// Add users to the user list
				for i := range chatInfo.Users {
					cs.Users[chatInfo.Users[i].Username] = &chatInfo.Users[i]
				}
			}
			onChatInfo(err, cs, chatInfo)

		case websock.ChatMessageReceived:
			chatMessage := new(websock.ChatMessage)
			if err = json.Unmarshal(msg.Message, chatMessage); err == nil {
				err = cs.DecryptChatMessages(chatMessage)
			}
			onChatMessage(err, cs, chatMessage)

		case websock.UserJoined:
			user := new(websock.User)
			if err = json.Unmarshal(msg.Message, user); err == nil {
				cs.Users[user.Username] = user
			}
			onUserJoined(err, cs, user)

		case websock.UserLeft:
			// Remove user from the user list
			username := string(msg.Message)
			delete(cs.Users, username)
			onUserLeft(cs, username)
		}
	}
}

// DecryptChatMessages decrypts chat messages using an RSA private key
func (cs *ChatSession) DecryptChatMessages(chatMessages ...*websock.ChatMessage) error {
	for i := range chatMessages {
		decMsg, err := rsa.DecryptPKCS1v15(rand.Reader, cs.PrivateKey, chatMessages[i].Message)
		if err != nil {
			return err
		}
		chatMessages[i].Message = decMsg
	}

	return nil
}

// SendChatMessage sends a chat message in the chat room of the chat session
// The message is encrypted with every participants public key, and sent to the server
func (cs *ChatSession) SendChatMessage(message string) {
	req := &websock.SendChatMessage{
		EncryptedContent: make(map[string][]byte),
		AuthKey:          cs.AuthKey}

	// For every user in the chat, encrypt the message with their public key
	for _, user := range cs.Users {
		pubKey, err := util.UnmarshalPublic(user.PublicKey)
		if err != nil {
			continue
		}
		encMsg, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(message))
		if err != nil {
			continue
		}
		req.EncryptedContent[user.Username] = encMsg
	}

	websock.SendMessage(cs.Socket, websock.SendChat, req, websock.JSON)
}
